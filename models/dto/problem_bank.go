package dto

import (
	"FanCode/models/po"
	"FanCode/utils"
)

// ProblemBankDtoForList 获取题目列表
type ProblemBankDtoForList struct {
	ID           uint       `json:"id"`
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	CreatedAt    utils.Time `json:"createdAt"`
	UpdatedAt    utils.Time `json:"updatedAt"`
	CreatorName  string     `json:"creatorName"`
	ProblemCount int64      `json:"problemCount"`
}

func NewProblemBankDtoForList(bank *po.ProblemBank) *ProblemBankDtoForList {
	response := &ProblemBankDtoForList{
		ID:          bank.ID,
		CreatedAt:   utils.Time(bank.CreatedAt),
		UpdatedAt:   utils.Time(bank.UpdatedAt),
		Name:        bank.Name,
		Description: bank.Description,
	}
	return response
}
