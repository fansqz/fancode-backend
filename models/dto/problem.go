package dto

import (
	"FanCode/models/po"
	"time"
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
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Name      string    `json:"name"`
	Number    string    `json:"number"`
	Title     string    `json:"title"`
	Path      string    `json:"path"`
}

func NewProblemDtoForList(question *po.Problem) *ProblemDtoForList {
	response := &ProblemDtoForList{
		ID:        question.ID,
		CreatedAt: question.CreatedAt,
		UpdatedAt: question.UpdatedAt,
		Name:      question.Name,
		Number:    question.Number,
		Title:     question.Title,
		Path:      question.Path,
	}
	return response
}
