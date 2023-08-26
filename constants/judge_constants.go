package constants

const (
	// Accepted 成功
	Accepted = 200 + iota
	// WrongAnswer 答案错误
	WrongAnswer
	TimeLimitExceeded
	// CompileError 编译出错
	CompileError
	// RuntimeError 运行出错
	RuntimeError
)
