package dto

// DebugStartRequestDto debug的请求
type DebugStartRequestDto struct {
	ProblemID  uint
	Code       string
	BreakPoint string // 断点
}
