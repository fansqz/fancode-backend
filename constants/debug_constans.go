package constants

type DebugMessageType string

const (
	RequestMessage  DebugMessageType = "request"
	ResponseMessage DebugMessageType = "response"
	EventMessage    DebugMessageType = "event"
)

// DebugOptionType 调试请求操作类型
type DebugOptionType string

const (
	// StartDebug 开始调试过程，返回可能出现的错误。
	StartDebug DebugOptionType = "start"
	// SendToConsole 输入数据到控制台，返回可能出现的错误。
	SendToConsole DebugOptionType = "sendToConsole"
	// Next 执行下一步操作，但不会进入函数内部，返回可能出现的错误。
	Next DebugOptionType = "next"
	// Step 执行下一步操作，会进入函数内部（如有调用函数，则步入函数），返回可能出现的错误。
	Step DebugOptionType = "step"
	// Continue：继续执行程序，直到遇到下一个断点或程序结束，返回可能出现的错误。
	Continue DebugOptionType = "continue"
	// AddBreakpoints 添加断点，接受文件源和断点列表，返回添加成功的断点和可能出现的错误。
	AddBreakpoints DebugOptionType = "addBreakpoints"
	// RemoveBreakpoints 移除断点，接受文件源和断点列表，返回移除成功的断点和可能出现的错误。
	RemoveBreakpoints DebugOptionType = "removeBreakpoints"
	// Terminate 终止当前的调试会话，之后可以重新调用 Launch 方法开始新的会话，返回可能出现的错误。
	Terminate DebugOptionType = "terminate"
)

type DebugEventType string

const (
	BreakpointEvent DebugEventType = "breakpoint"
	OutputEvent     DebugEventType = "output"
	StoppedEvent    DebugEventType = "stopped"
	ContinuedEvent  DebugEventType = "continued"
	CompileEvent    DebugEventType = "compile"
	ExitedEvent     DebugEventType = "exited"
	LaunchEvent     DebugEventType = "launch"
)

// BreakpointReasonType 断点改变类型
type BreakpointReasonType string

const (
	ChangeType  BreakpointReasonType = "change"
	NewType     BreakpointReasonType = "new"
	RemovedType BreakpointReasonType = "removed"
)

// StoppedReasonType 程序停止类型
type StoppedReasonType string

const (
	BreakpointStopped StoppedReasonType = "breakpoint"
	StepStopped       StoppedReasonType = "step"
)

// StepType 单步调试类型
type StepType string

const (
	StepIn   StepType = "stepIn"
	StepOut  StepType = "stepOut"
	StepOver StepType = "stepOver"
)
