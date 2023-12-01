package debug

import "FanCode/constants"

// DebugHandler
// 用户的一次调试过程处理
// 需要保证并发安全
type DebugHandler interface {
	// Compile 编译文件
	Compile(compileFiles []string, outFilePath string, options *CompileOptions) (*CompileResult, error)
	// Reset 重置DebugHandler对象
	Reset() error
	// StartDebug 开启调试
	// execFile:执行文件， workPath工作目录，调试过程中产生的文件可能会在该目录内
	// workPath最好是execFile所在目录，调试过程中会将该目录屏蔽，用/代替
	StartDebug(execFile string, workPath string, options *DebugOptions) (*DebugResult, error)
	// Next 下num步，不会进入函数内部
	Next(num int, options *DebugOptions) (*DebugResult, error)
	// Step 下num步，会进入函数内部
	Step(num int, options *DebugOptions) (*DebugResult, error)
	// Continue 忽略n个断点继续执行
	Continue(num int, options *DebugOptions) (*DebugResult, error)
	// AddBreakpoints 添加断点
	AddBreakpoints(breakpoints []Breakpoint) error
	// RemoveBreakpoints 移除断点
	RemoveBreakpoints(breakpoints []Breakpoint) error
}

func NewDebugHandler(language string) DebugHandler {
	switch language {
	case constants.ProgramC:
		return NewDebugHandlerC()
	}
	return nil
}
