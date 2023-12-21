package dto

import (
	"FanCode/models/po"
	"FanCode/utils"
)

type ProblemCaseDtoForList struct {
	ID        uint       `json:"id"`
	Name      string     `json:"name"`
	Input     string     `json:"input"`
	Output    string     `json:"output"`
	CreatedAt utils.Time `json:"createdAt"`
}

func NewProblemCaseDtoForList(problemCase *po.ProblemCase) *ProblemCaseDtoForList {
	return &ProblemCaseDtoForList{
		ID:        problemCase.ID,
		Name:      problemCase.Name,
		Input:     problemCase.Input,
		Output:    problemCase.Output,
		CreatedAt: utils.Time(problemCase.CreatedAt),
	}
}

type ProblemCaseDtoForGet struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Input  string `json:"input"`
	Output string `json:"output"`
}

func NewProblemCaseDtoForGet(problemCase *po.ProblemCase) *ProblemCaseDtoForGet {
	return &ProblemCaseDtoForGet{
		ID:     problemCase.ID,
		Name:   problemCase.Name,
		Input:  problemCase.Input,
		Output: problemCase.Output,
	}
}
