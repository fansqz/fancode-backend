package dto

import "time"

// 判题请请求需要的
type SubmitRequestDTO struct {
	ProblemID uint
	Code      string
}

type SubmitResultDTO struct {
	ProblemID      uint
	Status         uint
	ErrorMessage   string
	ExpectedOutput string //预期输出
	UserOutput     string //用户输出
	Timestamp      *time.Time
}

// ExecuteRequestDto 执行请求需要的dto
type ExecuteRequestDto struct {
	ProblemID uint   // 题目id
	Code      string // 代码
	Input     string // 自测用例
}

// ExecuteResultDto 执行的响应结果
type ExecuteResultDto struct {
	ProblemID    uint       `json:"question_id"`
	Status       uint       `json:"status"`
	ErrorMessage string     `json:"errorMessage"`
	UserOutput   string     `json:"userOutput"` //用户输出
	Timestamp    *time.Time `json:"timestamp"`
}
