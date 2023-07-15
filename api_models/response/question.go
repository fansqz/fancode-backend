package response

import (
	"FanCode/models"
	"time"
)

type QuestionResponseForList struct {
	ID          uint       `json:"id"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt"`
	Name        string     `json:"name"`
	Number      string     `json:"number"`
	Description string     `json:"description"`
	Title       string     `json:"title"`
	Path        string     `json:"path"`
}

func NewQuestionResponseForList(question *models.Question) *QuestionResponseForList {
	response := &QuestionResponseForList{
		ID:          question.ID,
		CreatedAt:   question.CreatedAt,
		UpdatedAt:   question.UpdatedAt,
		DeletedAt:   question.DeletedAt,
		Name:        question.Name,
		Number:      question.Number,
		Description: question.Description,
		Title:       question.Title,
		Path:        question.Path,
	}
	return response
}
