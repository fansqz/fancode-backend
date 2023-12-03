package dto

import "FanCode/models/po"

type UserCodeDto struct {
	ProblemID uint   `json:"problemID"`
	Code      string `json:"code"`
	CodeType  string `json:"codeType"`
	Language  string `json:"language"`
}

func NewUserCodeDto(attempt *po.ProblemAttempt) *UserCodeDto {
	return &UserCodeDto{
		ProblemID: attempt.ProblemID,
		Code:      attempt.Code,
		Language:  attempt.Language,
	}
}
