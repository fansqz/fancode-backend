package judger

// ExecuteOption 程序请求参数
type ExecuteOption struct {
	ExecFile string
	Language string

	// 用例输入管道
	InputCh <-chan []byte
	// 退出管道
	ExitCh <-chan string
	// 结果输出管道
	OutputCh chan<- ExecuteResult

	// 资源限制
	LimitTime   int64
	MemoryLimit int64
	CPUQuota    int64
}

// ExecuteResult 程序执行结果
type ExecuteResult struct {
	Executed     bool   // 判题是否执行成功
	ErrorMessage string // 异常信息
	Output       []byte // 输出结果（正常输出结果，如果有）
	UsedTime     int64  // 执行时间（以纳秒为单位）
	UsedMemory   int64  // 内存使用量（以字节为单位）
	UsedCpuTime  int64  // cpu使用
}
