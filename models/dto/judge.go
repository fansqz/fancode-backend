package dto

import "time"

// 判题请请求需要的
type JudgingRequestDTO struct {
	QuestionID uint
	Code       string
}

type SubmitResultDTO struct {
	QuestionID     uint
	Status         uint
	ErrorMessage   string
	ExpectedOutput string //预期输出
	UserOutput     string //用户输出
	Timestamp      *time.Time
}

type ExecuteResultDto struct {
	QuestionId     uint       `json:"question_id"`
	Status         uint       `json:"status"`
	ErrorMessage   string     `json:"errorMessage"`
	ExpectedOutput string     `json:"expectedOutput"` //预期输出
	UserOutput     string     `json:"userOutput"`     //用户输出
	Timestamp      *time.Time `json:"timestamp"`
}
