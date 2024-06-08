package service

import (
	"FanCode/config"
	"FanCode/constants"
	e "FanCode/error"
	"FanCode/models/dto"
	"FanCode/models/vo"
	"FanCode/service/debug"
	"FanCode/service/debug/debugger"
	"FanCode/utils"
	json2 "encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"path"
)

// DebugService
// 用户调试相关
type DebugService interface {
	// CreateDebugSession 创建调试session
	CreateDebugSession(ctx *gin.Context, language constants.LanguageType) (string, *e.Error)
	// CreateSseConnect 创建sse连接
	CreateSseConnect(ctx *gin.Context, key string)
	// CloseDebugSession 关闭用户程序并关闭调试session
	CloseDebugSession(key string) *e.Error

	// Start 加载并启动用户程序
	Start(ctx *gin.Context, startReq dto.StartDebugRequest) *e.Error
	SendToConsole(key string, input string) *e.Error
	StepIn(key string) *e.Error
	StepOver(key string) *e.Error
	StepOut(key string) *e.Error
	Continue(key string) *e.Error
	AddBreakpoints(key string, breakpoints []int) *e.Error
	RemoveBreakpoints(key string, breakpoints []int) *e.Error
}

type debugService struct {
	config       *config.AppConfig
	judgeService JudgeService
}

func NewDebugService(cf *config.AppConfig, js JudgeService) DebugService {
	return &debugService{
		config:       cf,
		judgeService: js,
	}
}

func (d *debugService) CreateDebugSession(ctx *gin.Context, language constants.LanguageType) (string, *e.Error) {
	key := utils.GetUUID()
	if err := debug.DebugSessionManage.CreateDebugSession(key, language); err != nil {
		debug.DebugSessionManage.DestroyDebugSession(key)
		return "", e.NewCustomMsg("系统出错")
	}
	return key, nil
}

func (d *debugService) CreateSseConnect(ctx *gin.Context, key string) {
	result := vo.NewResult(ctx)
	session, y := debug.DebugSessionManage.GetDebugSession(key)
	if !y {
		result.SimpleErrorMessage("key 不存在")
		return
	}
	w := ctx.Writer
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	fmt.Fprintf(w, "data: %s\n\n", "connect success")
	// 刷新缓冲，确保立即发送到客户端
	flusher, _ := w.(http.Flusher)
	flusher.Flush()
	// 遍历channel获取event并发送给前端
	channel := session.DtoEventChan
	stopChannel := session.StopProcessDtoEventChan
	for {
		select {
		case event := <-channel:
			json, err := json2.Marshal(event)
			if err != nil {
				continue
			}
			// 写入事件数据
			fmt.Fprintf(w, "data: %s\n\n", string(json))
			// 刷新缓冲，确保立即发送到客户端
			flusher.Flush()
		case <-stopChannel:
			return
		}
	}
}

func (d *debugService) CloseDebugSession(key string) *e.Error {
	// 获取调试上下文
	debug.DebugSessionManage.DestroyDebugSession(key)
	return nil
}

func (d *debugService) Start(ctx *gin.Context, startReq dto.StartDebugRequest) *e.Error {
	// 创建工作目录, 用户的临时文件
	executePath := getExecutePath(d.config)
	if err := os.MkdirAll(executePath, os.ModePerm); err != nil {
		log.Printf("MkdirAll error: %v\n", err)
		return e.ErrUnknown
	}
	// 保存用户代码到用户的执行路径，并获取编译文件列表
	var compileFiles []string
	var err2 *e.Error
	if compileFiles, err2 = d.saveUserCode(startReq.Language,
		startReq.Code, executePath); err2 != nil {
		return e.ErrUnknown
	}

	// 原本的处理event的协程
	debugSession, ok := debug.DebugSessionManage.GetDebugSession(startReq.Key)
	if !ok {
		return e.ErrUnknown
	}
	debugge := debugSession.Debugger

	//启动用户程序
	err := debugge.Launch(compileFiles, executePath, startReq.Language)
	if err != nil {
		return e.ErrUnknown
	}
	go func() {
		for {
			data := <-debugSession.DebuggerEventChan
			d.sendEventToSse(startReq.Key, d.getDebuggerEventToDtoEvent(data))
			if event, ok := data.(*debugger.LaunchEvent); ok {
				if event.Success {
					break
				} else {
					return
				}
			}
		}

		// 设置断点
		breakpoints := make([]*debugger.Breakpoint, len(startReq.Breakpoints))
		mainFile, _ := getMainFileNameByLanguage(startReq.Language)
		for i, bp := range startReq.Breakpoints {
			breakpoints[i] = &debugger.Breakpoint{
				File: mainFile,
				Line: bp,
			}
		}
		debugge.AddBreakpoints(breakpoints)

		// 确保所有断点都添加成功
		j := 0
		for j < len(breakpoints) {
			data := <-debugSession.DebuggerEventChan
			d.sendEventToSse(startReq.Key, d.getDebuggerEventToDtoEvent(data))
			if _, ok := data.(*debugger.BreakpointEvent); ok {
				j++
			}
		}

		// 启动用户程序
		_ = debugge.Start()
		for {
			select {
			case data := <-debugSession.DebuggerEventChan:
				d.sendEventToSse(startReq.Key, d.getDebuggerEventToDtoEvent(data))
			case _ = <-debugSession.StopProcessDtoEventChan:
				return
			}
		}
	}()
	return nil
}

func (d *debugService) sendEventToSse(key string, event interface{}) error {
	session, ok := debug.DebugSessionManage.GetDebugSession(key)
	if !ok {
		return errors.New("key 不存在")
	}
	channel := session.DtoEventChan
	channel <- event
	return nil
}

func (d *debugService) getDebuggerEventToDtoEvent(data interface{}) interface{} {
	var event interface{}
	if bevent, ok := data.(*debugger.BreakpointEvent); ok {
		bps := make([]int, len(bevent.Breakpoints))
		for i, bp := range bevent.Breakpoints {
			bps[i] = bp.Line
		}
		event = dto.BreakpointEvent{
			Event:       constants.BreakpointEvent,
			Reason:      bevent.Reason,
			Breakpoints: bps,
		}
	}
	if oevent, ok := data.(*debugger.OutputEvent); ok {
		event = dto.OutputEvent{
			Event:  constants.OutputEvent,
			Output: oevent.Output,
		}
	}
	if sevent, ok := data.(*debugger.StoppedEvent); ok {
		event = dto.StoppedEvent{
			Event:  constants.StoppedEvent,
			Reason: sevent.Reason,
			Line:   sevent.Line,
		}
	}
	if _, ok := data.(*debugger.ContinuedEvent); ok {
		event = dto.ContinuedEvent{
			Event: constants.ContinuedEvent,
		}
	}
	if eevent, ok := data.(*debugger.ExitedEvent); ok {
		event = dto.ExitedEvent{
			Event:    constants.ExitedEvent,
			ExitCode: eevent.ExitCode,
		}
	}
	if levent, ok := data.(*debugger.LaunchEvent); ok {
		event = dto.LaunchEvent{
			Event:   constants.LaunchEvent,
			Success: levent.Success,
			Message: levent.Message,
		}
	}
	if cevent, ok := data.(*debugger.CompileEvent); ok {
		event = dto.CompileEvent{
			Event:   constants.CompileEvent,
			Success: cevent.Success,
			Message: cevent.Message,
		}
	}
	return event
}

func (d *debugService) SendToConsole(key string, input string) *e.Error {
	// 获取调试上下文
	debugContext, _ := debug.DebugSessionManage.GetDebugSession(key)
	if err := debugContext.Debugger.SendToConsole(input); err != nil {
		log.Println(err)
		return e.ErrUnknown
	}
	return nil
}

func (d *debugService) StepIn(key string) *e.Error {
	// 获取调试上下文
	debugContext, _ := debug.DebugSessionManage.GetDebugSession(key)
	if err := debugContext.Debugger.StepIn(); err != nil {
		log.Println(err)
		return e.ErrUnknown
	}
	return nil
}

func (d *debugService) StepOut(key string) *e.Error {
	// 获取调试上下文
	debugContext, _ := debug.DebugSessionManage.GetDebugSession(key)
	if err := debugContext.Debugger.StepOut(); err != nil {
		log.Println(err)
		return e.ErrUnknown
	}
	return nil
}

func (d *debugService) StepOver(key string) *e.Error {
	// 获取调试上下文
	debugContext, _ := debug.DebugSessionManage.GetDebugSession(key)
	if err := debugContext.Debugger.StepOver(); err != nil {
		log.Println(err)
		return e.ErrUnknown
	}
	return nil
}

func (d *debugService) Continue(key string) *e.Error {
	// 获取调试上下文
	debugContext, _ := debug.DebugSessionManage.GetDebugSession(key)
	if err := debugContext.Debugger.Continue(); err != nil {
		log.Println(err)
		return e.ErrUnknown
	}
	return nil
}

func (d *debugService) AddBreakpoints(key string, breakpoints []int) *e.Error {
	// 获取调试上下文
	debugContext, _ := debug.DebugSessionManage.GetDebugSession(key)
	bps := make([]*debugger.Breakpoint, len(breakpoints))
	mainFile, err := getMainFileNameByLanguage(debugContext.Language)
	if err != nil {
		return err
	}
	for i, breakpoint := range breakpoints {
		bps[i] = &debugger.Breakpoint{
			File: mainFile,
			Line: breakpoint,
		}
	}
	if err := debugContext.Debugger.AddBreakpoints(bps); err != nil {
		log.Println(err)
		return e.ErrUnknown
	}
	return nil
}

func (d *debugService) RemoveBreakpoints(key string, breakpoints []int) *e.Error {
	// 获取调试上下文
	debugContext, _ := debug.DebugSessionManage.GetDebugSession(key)
	bps := make([]*debugger.Breakpoint, len(breakpoints))
	mainFile, err := getMainFileNameByLanguage(debugContext.Language)
	if err != nil {
		return err
	}
	for i, breakpoint := range breakpoints {
		bps[i] = &debugger.Breakpoint{
			File: mainFile,
			Line: breakpoint,
		}
	}
	if err := debugContext.Debugger.RemoveBreakpoints(bps); err != nil {
		log.Println(err)
		return e.ErrUnknown
	}
	return nil
}

// saveUserCode
// 保存用户代码到用户的executePath，并返回需要编译的文件列表
func (d *debugService) saveUserCode(language constants.LanguageType, codeStr string, executePath string) ([]string, *e.Error) {
	var compileFiles []string
	var mainFile string
	var err2 *e.Error

	if mainFile, err2 = getMainFileNameByLanguage(language); err2 != nil {
		log.Println(err2)
		return nil, err2
	}
	if err := os.WriteFile(path.Join(executePath, mainFile), []byte(codeStr), 0644); err != nil {
		log.Println(err)
		return nil, e.ErrServer
	}
	// 将main文件进行编译即可
	compileFiles = []string{path.Join(executePath, mainFile)}

	return compileFiles, nil
}
