package dto

import (
	"FanCode/models/po"
	"time"
)

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

type ProblemDtoForList struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Name        string    `json:"name"`
	Number      string    `json:"number"`
	Description string    `json:"description"`
	Title       string    `json:"title"`
	Path        string    `json:"path"`
}

func NewProblemDtoForList(question *po.Problem) *ProblemDtoForList {
	response := &ProblemDtoForList{
		ID:          question.ID,
		CreatedAt:   question.CreatedAt,
		UpdatedAt:   question.UpdatedAt,
		Name:        question.Name,
		Number:      question.Number,
		Description: question.Description,
		Title:       question.Title,
		Path:        question.Path,
	}
	return response
}
