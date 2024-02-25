package dto

import "FanCode/constants"

// ======================DebugBase调试信息的基本结构==============================
type DebugBase struct {
	Type constants.DebugMessageType
}

// =======================以下是request=============================
// DebugRequestBase 调试相关请求数据
type DebugRequestBase struct {
	DebugBase
	Seq    int32
	Option constants.DebugOptionType
}

// StartDebugRequest 启动调试请求
type StartDebugRequest struct {
	DebugRequestBase
	// Code 需要进行debug的用户代码
	Code string
	// Language 调试语言
	Language constants.LanguageType
}

// =======================以下是resp=============================
// DebugResponseBase 调试相关请求数据
type DebugResponse struct {
	DebugBase
	// RequestSeq 请求对应的seq
	RequestSeq int32
	// Success 请求是否成功
	Success bool
	// RequestOption 请求对应的option
	RequestOption constants.DebugOptionType
	// Message 携带信息
	Message string
}

func NewSuccessDebugResponseByRequest(req *DebugRequestBase, message string) *DebugResponse {
	return &DebugResponse{
		DebugBase: DebugBase{
			Type: req.Type,
		},
		Success:       true,
		RequestSeq:    req.Seq,
		RequestOption: req.Option,
		Message:       message,
	}
}

func NewFailDebugResponseByRequest(req *DebugRequestBase, message string) *DebugResponse {
	return &DebugResponse{
		DebugBase: DebugBase{
			Type: req.Type,
		},
		Success:       false,
		RequestSeq:    req.Seq,
		RequestOption: req.Option,
		Message:       message,
	}
}

//=======================以下是event=============================

type DebugEventBase struct {
	DebugBase
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
