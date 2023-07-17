package constants

const (
	Accepted = 200 + iota
	WrongAnswer
	TimeLimitExceeded
	// 编译出错
	CompileError
	// 运行出错
	RuntimeError
)
