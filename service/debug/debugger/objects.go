package debugger

import "FanCode/constants"

// Breakpoint 表示断点
type Breakpoint struct {
	File string // 文件名称
	Line int    // 行号
}

// StackFrame 栈帧
type StackFrame struct {
	Id   string `json:"id"`   // 栈帧id
	Name string `json:"name"` // 函数名称
	Path string `json:"path"` // 文件路径
	Line int    `json:"line"`
}

// Scope 作用域
type Scope struct {
	Name               constants.ScopeName
	VariablesReference string // 作用域的引用
}

// Variable 变量
type Variable struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
	// 变量引用
	Reference string `json:"reference"`
}
