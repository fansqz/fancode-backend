package debug

import (
	"FanCode/constants"
	"FanCode/service/debug/debugger"
)

/**
 * 调试上下文对象，用于存储用户的一次调试的信息
 */
type DebugSession struct {
	// 用于停止循环处理调试器返回的event
	StopProcessDebuggerEventChan chan struct{}
	// 用于停止循环处理service返回的event
	StopProcessDtoEventChan chan struct{}
	// DebuggerEventChan 返回调试信息给service的管道
	DebuggerEventChan chan interface{}
	// DtoEventChan 将event返回给用户的channel
	DtoEventChan chan interface{}
	// Debugger 用户的调试器
	Debugger debugger.Debugger
	// Language 调试的语言类型
	Language constants.LanguageType
}
