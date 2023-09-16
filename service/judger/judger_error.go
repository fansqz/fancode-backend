package judger

var ExecuteTimoutErr error = executeTimeoutErr{}
var ExecuteMemoryExceededErr error = executeMemoryExceededErr{}

type executeTimeoutErr struct{}

func (executeTimeoutErr) Error() string {
	return "运行超时"
}

type executeMemoryExceededErr struct{}

func (executeMemoryExceededErr) Error() string {
	return "内存使用超出限制"
}
