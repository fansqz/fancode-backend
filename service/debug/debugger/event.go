package debugger

import "FanCode/constants"

// BreakpointEvent 断点事件
// 该event指示有关断点的某些信息已更改。
type BreakpointEvent struct {
	Reason      constants.BreakpointReasonType
	Breakpoints []Breakpoint
}

// OutputEvent
// 该事件表明目标已经产生了一些输出。
type OutputEvent struct {
	Category constants.OutputCategory // 输出类型
	Output   string                   // 输出内容
	Line     int                      // 产生输出的位置的行。
}

// StoppedEvent
// 该event表明，由于某些原因，被调试进程的执行已经停止。
// 这可能是由先前设置的断点、完成的步进请求、执行调试器语句等引起的。
type StoppedEvent struct {
	Reason constants.StoppedReasonType // 停止执行的原因
	File   string                      // 当前停止在哪个文件
	Line   int                         // 停止在某行
}

// ContinuedEvent
// 该event表明debug的执行已经继续。
// 请注意:debug adapter不期望发送此事件来响应暗示执行继续的请求，例如启动或继续。
// 它只有在没有先前的request暗示这一点时，才有必要发送一个持续的事件。
type ContinuedEvent struct {
}

// ExitedEvent
// 该event表明被调试对象已经退出并返回exit code。但是并不意味着调试会话结束
type ExitedEvent struct {
	ExitCode int
}

// TerminatedEvent
// 该event表示调试会话结束
type TerminatedEvent struct {
}

// CompileEvent
// 编译事件
type CompileEvent struct {
	Success bool
	Message string // 编译产生的信息
}

// LaunchEvent
// 启动gdb的事件
type LaunchEvent struct {
	Success bool
	Message string // 启动gdb的消息
}
