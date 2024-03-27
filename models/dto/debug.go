package dto

import "FanCode/constants"

// =======================以下是request=============================

// StartDebugRequest 启动调试请求
type StartDebugRequest struct {
	// Code 需要进行debug的用户代码
	Code string `json:"code"`
	// Language 调试语言
	Language constants.LanguageType `json:"language"`
	// 初始断点
	Breakpoints []int `json:"breakpoints"`
}

type AddBreakpointRequest struct {
	Key         string `json:"key"`
	Breakpoints []int  `json:"breakpoints"`
}

type RemoveBreakpointRequest struct {
	Key         string `json:"key"`
	Breakpoints []int  `json:"breakpoints"`
}

type StartDebugEvent struct {
	Event   constants.DebugEventType
	Success bool   `json:"success"`
	Key     string `json:"key"`
}

type DebugEventBase struct {
	Seq   int32
	Event constants.DebugEventType
}

// CompileEvent
// 编译事件
type CompileEvent struct {
	DebugEventBase
	Success bool
	Message string // 编译产生的信息
}

// BreakpointEvent 断点事件
// 该event指示有关断点的某些信息已更改。
type BreakpointEvent struct {
	DebugEventBase
	Reason constants.BreakpointReasonType
}

// OutputEvent
// 该事件表明目标已经产生了一些输出。
type OutputEvent struct {
	DebugEventBase
	Category OutputCategory // 输出类型
	Output   string         // 输出内容
	Line     int            // 产生输出的位置的行。
}

// OutputCategory 输出类型定义
type OutputCategory string

// StoppedEvent
// 该event表明，由于某些原因，被调试进程的执行已经停止。
// 这可能是由先前设置的断点、完成的步进请求、执行调试器语句等引起的。
type StoppedEvent struct {
	DebugEventBase
	HitBreakpointIds []int
}

// ContinuedEvent
// 该event表明debug的执行已经继续。
// 请注意:debug adapter不期望发送此事件来响应暗示执行继续的请求，例如启动或继续。
// 它只有在没有先前的request暗示这一点时，才有必要发送一个持续的事件。
type ContinuedEvent struct {
	DebugEventBase
}

// ExitedEvent
// 该event表明被调试对象已经退出并返回exit code。
type ExitedEvent struct {
	DebugEventBase
	ExitCode int
}

// TerminatedEvent
// 程序退出事件
type TerminatedEvent struct {
	DebugEventBase
}
