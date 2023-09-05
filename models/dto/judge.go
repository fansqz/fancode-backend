package dto

import (
	"FanCode/models/po"
	"time"
)

// SubmitRequestDto 判题请请求需要的
type SubmitRequestDto struct {
	ProblemID uint
	Code      string
}

type SubmitResultDto struct {
	ProblemID    uint   `json:"problemID"`
	Status       int    `json:"status"`
	ErrorMessage string `json:"errorMessage"`
	// 预期输出
	ExpectedOutput string `json:"expectedOutput"`
	// 用户输出
	UserOutput string     `json:"userOutput"`
	Timestamp  *time.Time `json:"timestamp"`
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
	ProblemID    uint       `json:"problemID"`
	Status       uint       `json:"status"`
	ErrorMessage string     `json:"errorMessage"`
	UserOutput   string     `json:"userOutput"` //用户输出
	Timestamp    *time.Time `json:"timestamp"`
}
