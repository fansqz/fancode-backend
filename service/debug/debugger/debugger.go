package debugger

// Debugger
// 用户的一次调试过程处理
// debugger目前设置为支持多文件的
// 需要保证并发安全
type Debugger interface {
	// Launch 开启debuggee，及gdb启动的命令
	// compileFiles 编译的文件列表
	// workPath 工作目录
	Launch(compileFiles []string, workPath string) (chan interface{}, error)
	// Start
	// 开始调试，及调用start命令
	Start() error
	// SendToConsole 输入数据到控制台
	SendToConsole(input string) error
	// Next 下一步，不会进入函数内部
	Next() error
	// Step 下n一步，会进入函数内部
	Step() error
	// Continue 忽略继续执行
	Continue() error
	// AddBreakpoints 添加断点
	// 返回的是添加成功的断点
	AddBreakpoints(breakpoints []Breakpoint) error
	// RemoveBreakpoints 移除断点
	// 返回的是移除成功的断点
	RemoveBreakpoints(breakpoints []Breakpoint) error
	// Terminate 终止调试
	// 调用完该命令以后可以重新Launch
	Terminate() error
}
