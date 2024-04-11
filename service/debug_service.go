package service

import (
	"FanCode/config"
	"FanCode/constants"
	e "FanCode/error"
	"FanCode/models/dto"
	"FanCode/models/vo"
	"FanCode/service/debug"
	"FanCode/service/debug/debugger"
	"FanCode/service/sse"
	"FanCode/utils"
	"github.com/gin-gonic/gin"
	"log"
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
	// Start 加载并启动用户程序
	Start(ctx *gin.Context, startReq dto.StartDebugRequest) *e.Error
	SendToConsole(key string, input string) *e.Error
	Next(key string) *e.Error
	Step(key string) *e.Error
	Continue(key string) *e.Error
	AddBreakpoints(key string, breakpoints []int) *e.Error
	RemoveBreakpoints(key string, breakpoints []int) *e.Error
	// CloseDebugSession 关闭用户程序并关闭调试session
	CloseDebugSession(key string) *e.Error
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

	// 暂停event的处理
	debugSession.StopProcessEventChan <- struct{}{}

	//启动用户程序
	err := debugge.Launch(compileFiles, executePath, startReq.Language)
	if err != nil {
		return e.ErrUnknown
	}

	go func() {
		for {
			data := <-debugSession.DebuggerChan
			if event, ok := data.(*debugger.LaunchEvent); ok {
				if event.Success {
					d.processDebugEvent(startReq.Key, data)
					break
				} else {
					d.processDebugEvent(startReq.Key, data)
					go d.startProcessDebugEvent(startReq.Key)
					return
				}
			} else {
				d.processDebugEvent(startReq.Key, data)
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
		for {
			data := <-debugSession.DebuggerChan
			if _, ok := data.(*debugger.BreakpointEvent); ok {
				j++
				if j == len(breakpoints) {
					break
				}
			}
			d.processDebugEvent(startReq.Key, data)
		}

		// 启动用户程序
		_ = debugge.Start()
		go d.startProcessDebugEvent(startReq.Key)
	}()
	return nil
}

func (d *debugService) CreateSseConnect(ctx *gin.Context, key string) {
	result := vo.NewResult(ctx)
	sse.Close(key)
	_, y := debug.DebugSessionManage.GetDebugSession(key)
	if !y {
		result.SimpleErrorMessage("key 不存在")
		return
	}
	// gdb调试启动成功，创建管道
	sse.CreateSssConnection(key, ctx.Writer)
	d.startProcessDebugEvent(key)
}

func (d *debugService) startProcessDebugEvent(key string) {
	debugContext, y := debug.DebugSessionManage.GetDebugSession(key)
	if !y {
		return
	}
	// 循环遍历所有输入数据
	for {
		select {
		case data := <-debugContext.DebuggerChan:
			d.processDebugEvent(key, data)
		case _ = <-debugContext.StopProcessEventChan:
			return
		}
	}
}

func (d *debugService) processDebugEvent(key string, data interface{}) {
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
			Event:  constants.StoppedEvent,
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
			Event: constants.CompileEvent,
		}
	}
	if eevent, ok := data.(*debugger.ExitedEvent); ok {
		event = dto.ExitedEvent{
			Event:    constants.ExitedEvent,
			ExitCode: eevent.ExitCode,
		}
	}
	if err := sse.SendData(key, event); err != nil {
		log.Println(err)
	}
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

func (d *debugService) Next(key string) *e.Error {
	// 获取调试上下文
	debugContext, _ := debug.DebugSessionManage.GetDebugSession(key)
	if err := debugContext.Debugger.Next(); err != nil {
		log.Println(err)
		return e.ErrUnknown
	}
	return nil
}

func (d *debugService) Step(key string) *e.Error {
	// 获取调试上下文
	debugContext, _ := debug.DebugSessionManage.GetDebugSession(key)
	if err := debugContext.Debugger.Step(); err != nil {
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

func (d *debugService) CloseDebugSession(key string) *e.Error {
	// 获取调试上下文
	debugContext, _ := debug.DebugSessionManage.GetDebugSession(key)
	if err := debugContext.Debugger.Terminate(); err != nil {
		log.Println(err)
		return e.ErrUnknown
	}
	if err := sse.Close(key); err != nil {
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
