package dto

import (
	"FanCode/models/po"
	"FanCode/utils"
)

// 获取题目详细信息
type ProblemDtoForGet struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Number      string `json:"number"`
	Description string `json:"description"`
	Title       string `json:"title"`
	Path        string `json:"path"`
}

func NewProblemDtoForGet(problem *po.Problem) *ProblemDtoForGet {
	response := &ProblemDtoForGet{
		ID:          problem.ID,
		Name:        problem.Name,
		Number:      problem.Number,
		Description: problem.Description,
		Title:       problem.Title,
		Path:        problem.Path,
	}
	return response
}

// 获取题目列表
type ProblemDtoForList struct {
	ID        uint       `json:"id"`
	CreatedAt utils.Time `json:"createdAt"`
	UpdatedAt utils.Time `json:"updatedAt"`
	Name      string     `json:"name"`
	Number    string     `json:"number"`
	Title     string     `json:"title"`
	Path      string     `json:"path"`
}

func NewProblemDtoForList(question *po.Problem) *ProblemDtoForList {
	response := &ProblemDtoForList{
		ID:        question.ID,
		CreatedAt: utils.Time(question.CreatedAt),
		UpdatedAt: utils.Time(question.UpdatedAt),
		Name:      question.Name,
		Number:    question.Number,
		Title:     question.Title,
		Path:      question.Path,
	}
	return response
}