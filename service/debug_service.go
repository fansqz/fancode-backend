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
	Start(ctx *gin.Context, startReq dto.StartDebugRequest) (string, *e.Error)
	CreateSseConnect(ctx *gin.Context, key string)
	SendToConsole(key string, input string) *e.Error
	Next(key string) *e.Error
	Step(key string) *e.Error
	Continue(key string) *e.Error
	AddBreakpoints(key string, breakpoints []int) *e.Error
	RemoveBreakpoints(key string, breakpoints []int) *e.Error
	Terminate(key string) *e.Error
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

func (d *debugService) Start(ctx *gin.Context, startReq dto.StartDebugRequest) (string, *e.Error) {
	result := vo.NewResult(ctx)
	// 创建工作目录, 用户的临时文件
	executePath := getExecutePath(d.config)
	if err := os.MkdirAll(executePath, os.ModePerm); err != nil {
		log.Printf("MkdirAll error: %v\n", err)
		result.Error(e.ErrBadRequest)
		return "", e.ErrUnknown
	}
	// 保存用户代码到用户的执行路径，并获取编译文件列表
	var compileFiles []string
	var err2 *e.Error
	if compileFiles, err2 = d.saveUserCode(startReq.Language,
		startReq.Code, executePath); err2 != nil {
		result.SimpleErrorMessage("保存用户代码失败")
		return "", e.ErrUnknown
	}

	// 启动debugging
	key := utils.GetUUID()
	if err := debug.StartDebugging(key, startReq.Language, compileFiles, executePath); err != nil {
		// Add logging for error
		result.SimpleErrorMessage("启动调试失败")
		return "", e.ErrUnknown
	}

	// 读取输入数据的管道并创建协程处理数据
	debugContext, _ := debug.GetDebugContext(key)

	// 等待gdb启动成功
	//ans := <-debugContext.DebuggerChan
	//if launchEvent, ok := ans.(*debugger.LaunchEvent); ok {
	//	if !launchEvent.Success {
	//		result.SimpleErrorMessage("调试任务启动失败")
	//		debug.DestroyDebugContext(key)
	//		return "", e.ErrUnknown
	//	}
	//} else {
	//	// 启动失败
	//	result.SimpleErrorMessage("调试任务启动失败")
	//	return "", e.ErrUnknown
	//}

	// 设置断点
	breakpoints := make([]debugger.Breakpoint, len(startReq.Breakpoints))
	mainFile, _ := getMainFileNameByLanguage(debugContext.Language)
	for i, bp := range startReq.Breakpoints {
		breakpoints[i] = debugger.Breakpoint{
			File: mainFile,
			Line: bp,
		}
	}
	debugContext.Debugger.AddBreakpoints(breakpoints)

	// 开启调试
	_ = debugContext.Debugger.Start()
	return key, nil
}

func (d *debugService) CreateSseConnect(ctx *gin.Context, key string) {
	result := vo.NewResult(ctx)
	sse.Close(key)
	debugContext, y := debug.GetDebugContext(key)
	if !y {
		result.SimpleErrorMessage("key 不存在")
		return
	}
	// gdb调试启动成功，创建管道
	sse.CreateSssConnection(key, ctx.Writer)
	// 循环遍历所有输入数据
	for {
		data := <-debugContext.DebuggerChan
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
				Event:    constants.StoppedEvent,
				Category: oevent.Category,
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
}

func (d *debugService) SendToConsole(key string, input string) *e.Error {
	// 获取调试上下文
	debugContext, _ := debug.GetDebugContext(key)
	if err := debugContext.Debugger.SendToConsole(input); err != nil {
		log.Println(err)
		return e.ErrUnknown
	}
	return nil
}

func (d *debugService) Next(key string) *e.Error {
	// 获取调试上下文
	debugContext, _ := debug.GetDebugContext(key)
	if err := debugContext.Debugger.Next(); err != nil {
		log.Println(err)
		return e.ErrUnknown
	}
	return nil
}

func (d *debugService) Step(key string) *e.Error {
	// 获取调试上下文
	debugContext, _ := debug.GetDebugContext(key)
	if err := debugContext.Debugger.Step(); err != nil {
		log.Println(err)
		return e.ErrUnknown
	}
	return nil
}

func (d *debugService) Continue(key string) *e.Error {
	// 获取调试上下文
	debugContext, _ := debug.GetDebugContext(key)
	if err := debugContext.Debugger.Continue(); err != nil {
		log.Println(err)
		return e.ErrUnknown
	}
	return nil
}

func (d *debugService) AddBreakpoints(key string, breakpoints []int) *e.Error {
	// 获取调试上下文
	debugContext, _ := debug.GetDebugContext(key)
	bps := make([]debugger.Breakpoint, len(breakpoints))
	mainFile, err := getMainFileNameByLanguage(debugContext.Language)
	if err != nil {
		return err
	}
	for i, breakpoint := range breakpoints {
		bps[i] = debugger.Breakpoint{
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
	debugContext, _ := debug.GetDebugContext(key)
	bps := make([]debugger.Breakpoint, len(breakpoints))
	mainFile, err := getMainFileNameByLanguage(debugContext.Language)
	if err != nil {
		return err
	}
	for i, breakpoint := range breakpoints {
		bps[i] = debugger.Breakpoint{
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

func (d *debugService) Terminate(key string) *e.Error {
	// 获取调试上下文
	debugContext, _ := debug.GetDebugContext(key)
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

// listenAndHandleDebugEvents 循环监控调试事件，并生成event响应给用户
func (d *debugService) listenAndHandleDebugEvents(ctx *gin.Context, channel chan interface{}) {

}
