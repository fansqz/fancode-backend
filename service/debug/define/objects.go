package define

// Breakpoint 表示断点
type Breakpoint struct {
	Line int // 行号
}

// StackFrame 表示一个栈帧，包含当前栈帧的函数名称，执行的位置等信息
type StackFrame struct {
	Function string // 调用函数名称
	Args     string // 参数
	File     string // 函数所在文件
	Line     int    // 方法返回地址
}

type BackTrace []StackFrame
