package debug

import "strconv"

// DebugOptions 调试方法可选参数
type DebugOptions struct {
	Breakpoints []Breakpoint // 断点，仅在start中使用
	LimitTime   int64        // 超时时间，用于限制命令的运行时间，如果超时调试程序将终止
	DebugFile   string       // 关注的文件，如果不为空，则调试不回跳出该文件，在文件内部执行
}

// DebugResult 调试统一返回格式
type DebugResult struct {
	IsEnd      bool      // 调试是否结束
	File       string    // 运行到哪个文件
	Function   string    // 运行到那个函数
	Line       int       // 行号
	UserOutput string    // 用户输出
	BackTrace  BackTrace // 调用栈信息
}

// CompileOptions 编译文件可选参数
// 和judger中的CompileOptions类似，但是这里不需要language
type CompileOptions struct {
	LimitTime       int64
	ExcludedPaths   []string // 屏蔽的敏感路径
	ReplacementPath string   // 取代敏感路径的路径
}

// CompileResult 系统编译结果
type CompileResult struct {
	Compiled         bool   // 判题是否编译成功
	ErrorMessage     string // 异常信息
	CompiledFilePath string // 输出文件路径
}

// Breakpoint 表示断点
type Breakpoint struct {
	File string // 编程文件名
	Line int    // 行号
}

func (b *Breakpoint) toString() string {
	return b.File + "," + strconv.Itoa(b.Line)
}

// StackFrame 表示一个栈帧，包含当前栈帧的函数名称，执行的位置等信息
type StackFrame struct {
	Function string // 调用函数名称
	Args     string // 参数
	File     string // 函数所在文件
	Line     int    // 方法返回地址
}

type BackTrace []StackFrame
