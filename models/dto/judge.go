package dto

import (
	"FanCode/models/po"
	"time"
)

// 判题请请求需要的
type SubmitRequestDto struct {
	ProblemID uint
	Code      string
}

type SubmitResultDto struct {
	ProblemID      uint
	Status         int
	ErrorMessage   string
	ExpectedOutput string //预期输出
	UserOutput     string //用户输出
	Timestamp      *time.Time
}

func NewSubmitResultDto(submission *po.Submission) *SubmitResultDto {
	return &SubmitResultDto{
		ProblemID:      submission.ProblemID,
		Status:         submission.Status,
		ErrorMessage:   submission.ErrorMessage,
		ExpectedOutput: submission.ExpectedOutput,
		UserOutput:     submission.UserOutput,
	}
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
