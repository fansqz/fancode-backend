package judger

import "time"

// ExecuteOption 程序请求参数
type ExecuteOption struct {
	ExecFile string
	Language int

	// 用例输入管道
	InputCh <-chan []byte
	// 退出管道
	ExitCh <-chan string
	// 结果输出管道
	OutputCh chan<- ExecuteResult

	LimitTime   time.Duration
	LimitMemory int
}

// ExecuteResult 程序执行结果
type ExecuteResult struct {
	Executed      bool   // 判题是否执行成功
	Error         error  // 异常
	Output        []byte // 输出结果（如果有）
	ExecutionTime int64  // 执行时间（以纳秒为单位）
	MemoryUsed    int64  // 内存使用量（以字节为单位）
}
