package debug

// DebugOptions 调试可选参数
type DebugOptions struct {
	CodeFile    string   // 断点文件
	Breakpoints []string // 断点
	LimitTime   int64    // 超时时间
}

// DebugResult 调试统一返回格式
type DebugResult struct {
	IsEnd      bool     // 调试是否结束
	LineNumber string   // 行号
	Output     string   // /标准输出
	BackTrace  []string // 调用栈信息
}
