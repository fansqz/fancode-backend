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
	StopProcessEventChan chan struct{}
	// Debugger 用户的调试器
	Debugger debugger.Debugger
	// DebuggerChan 返回调试信息给用户的管道
	DebuggerChan chan interface{}
	// Language 调试的语言类型
	Language constants.LanguageType
}
