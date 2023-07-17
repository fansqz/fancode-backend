package dto

import "time"

// 判题请请求需要的
type JudgingRequestDTO struct {
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

type ExecuteResultDto struct {
	ProblemID      uint       `json:"question_id"`
	Status         uint       `json:"status"`
	ErrorMessage   string     `json:"errorMessage"`
	ExpectedOutput string     `json:"expectedOutput"` //预期输出
	UserOutput     string     `json:"userOutput"`     //用户输出
	Timestamp      *time.Time `json:"timestamp"`
}
