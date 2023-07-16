package dto

import "time"

// 判题请请求需要的
type JudgingRequestDTO struct {
	QuestionID uint
	UserID     uint
	Code       string
}

type SubmitResultDTO struct {
	ProblemId    uint
	Status       uint
	ErrorMessage string
	Timestamp    *time.Time
}

type ExecuteResultDto struct {
	ProblemId    uint
	Status       uint
	ErrorMessage string
	Timestamp    *time.Time
}
