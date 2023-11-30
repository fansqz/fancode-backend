package dto

import (
	"FanCode/models/po"
	"FanCode/utils"
)

type SubmissionDtoForList struct {
	ID           uint       `json:"id"`
	ProblemName  string     `json:"problemName"`
	Status       int        `json:"status"`
	ErrorMessage string     `json:"errorMessage"`
	CreatedAt    utils.Time `json:"createdAt"`
}

func NewSubmissionDtoForList(submission *po.Submission) *SubmissionDtoForList {
	return &SubmissionDtoForList{
		ID:           submission.ID,
		Status:       submission.Status,
		ErrorMessage: submission.ErrorMessage,
		CreatedAt:    utils.Time(submission.CreatedAt),
	}
}
