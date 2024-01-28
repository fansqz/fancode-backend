package define

// BreakpointEvent 断点事件
// 该event指示有关断点的某些信息已更改。
type BreakpointEvent struct {
	Reason ReasonType
}

type ReasonType string

const (
	ChangeType  ReasonType = "change"
	NewType     ReasonType = "new"
	RemovedType ReasonType = "removed"
)

// OutputEvent
// 该事件表明目标已经产生了一些输出。
type OutputEvent struct {
	Category OutputCategory // 输出类型
	Output   string         // 输出内容
	Line     int            // 产生输出的位置的行。
}

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

// StoppedEvent
// 该event表明，由于某些原因，被调试进程的执行已经停止。
// 这可能是由先前设置的断点、完成的步进请求、执行调试器语句等引起的。
type StoppedEvent struct {
	HitBreakpointIds []int
}

// ContinuedEvent
// 该event表明debug的执行已经继续。
// 请注意:debug adapter不期望发送此事件来响应暗示执行继续的请求，例如启动或继续。
// 它只有在没有先前的request暗示这一点时，才有必要发送一个持续的事件。
type ContinuedEvent struct {
}

// CompileEvent
// 编译事件
type CompileEvent struct {
	Success bool
	Message string // 编译产生的信息
}

// ExitedEvent
// 该event表明被调试对象已经退出并返回exit code。
type ExitedEvent struct {
	ExitCode int
}
