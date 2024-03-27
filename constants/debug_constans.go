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
	StartEvent      DebugEventType = "start"
	BreakpointEvent DebugEventType = "breakpoint"
	OutputEvent     DebugEventType = "output"
	StoppedEvent    DebugEventType = "stopped"
	ContinuedEvent  DebugEventType = "continued"
	CompileEvent    DebugEventType = "compile"
	ExitedEvent     DebugEventType = "exited"
)

// BreakpointReasonType 断点改变类型
type BreakpointReasonType string

const (
	ChangeType  BreakpointReasonType = "change"
	NewType     BreakpointReasonType = "new"
	RemovedType BreakpointReasonType = "removed"
)

// OutputCategory 输出类型定义
type OutputCategory string

const (
	// Console 在客户端的默认消息UI中显示输出
	Console OutputCategory = "console"
	// Important 提示客户端在客户端UI中显示输出
	// 用于重要和高度可见的信息，例如作为弹出窗口通知。此类别应仅用于重要邮件
	// 来自调试器(而不是debuggee)。这个类别值是一个提示，客户端可能会忽略这个提示
	Important OutputCategory = "important"
	// Stdout 将输出显示为被调试程序的正常程序输出。
	Stdout OutputCategory = "stdout"
	// Stderr 将输出显示为错误程序从被调试程序输出。
	Stderr OutputCategory = "stderr"
	// Telemetry 将输出发送到Telemetry，而不是显示给用户。
	Telemetry OutputCategory = "telemetry"
)
