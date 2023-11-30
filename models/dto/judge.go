package dto

import (
	"FanCode/models/po"
	"time"
)

// SubmitRequestDto 判题请请求需要的
type SubmitRequestDto struct {
	ProblemID uint
	Code      string
	Language  string
	CodeType  string
}

type SubmitResultDto struct {
	ProblemID    uint   `json:"problemID"`
	Status       int    `json:"status"`
	ErrorMessage string `json:"errorMessage"`
	CaseName     string `json:"caseName"`
	CaseData     string `json:"caseData"`
	// 预期输出
	ExpectedOutput string `json:"expectedOutput"`
	// 用户输出
	UserOutput string `json:"userOutput"`
	// 判题使用时间
	TimeUsed time.Duration `json:"timeUsed"`
	// 内存使用量（以字节为单位）
	MemoryUsed int64 `json:"memoryUsed"`
}

func NewSubmitResultDto(submission *po.Submission) *SubmitResultDto {
	return &SubmitResultDto{
		ProblemID:      submission.ProblemID,
		Status:         submission.Status,
		ErrorMessage:   submission.ErrorMessage,
		CaseName:       submission.CaseName,
		CaseData:       submission.CaseData,
		ExpectedOutput: submission.ExpectedOutput,
		UserOutput:     submission.UserOutput,
		TimeUsed:       submission.TimeUsed,
		MemoryUsed:     submission.MemoryUsed,
	}
}

// ExecuteRequestDto 执行请求需要的dto
type ExecuteRequestDto struct {
	ProblemID uint   // 题目id
	Code      string // 代码
	Input     string // 自测用例
	Language  string // 编程语言
	CodeType  string // acm或核心代码
}

// ExecuteResultDto 执行的响应结果
type ExecuteResultDto struct {
	ProblemID    uint   `json:"problemID"`
	Status       uint   `json:"status"`
	ErrorMessage string `json:"errorMessage"`
	UserOutput   string `json:"userOutput"` //用户输出
}
