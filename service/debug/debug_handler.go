package debug

// DebugHandler
// 用户的一次调试过程处理
// 需要保证并发安全
type DebugHandler interface {
	// Reset 重置DebugHandler对象
	Reset() error
	// StartDebug 开启调试
	StartDebug(execFile string, workPath string, options *DebugOptions) (*DebugResult, error)
	// Next 下num步，不会进入函数内部
	Next(num int) (*DebugResult, error)
	// Step 下num步，会进入函数内部
	Step(num int) (*DebugResult, error)
	// Continue 忽略n个断点继续执行
	Continue(num int) (*DebugResult, error)
	// AddBreakpoints 添加断点
	AddBreakpoints(breakpoints []string) error
	// RemoveBreakpoints 移除断点
	RemoveBreakpoints(breakpoints []string) error
}
