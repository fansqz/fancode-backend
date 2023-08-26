package po

import (
	"gorm.io/gorm"
	"time"
)

type Submission struct {
	gorm.Model
	UserID         uint       `gorm:"column:user_id"`
	ProblemID      uint       `gorm:"column:problem_id"`
	Code           string     `gorm:"column:code"`
	SubmitTime     *time.Time `gorm:"column:submit_time"`
	Status         int        `gorm:"column:status"`
	ErrorMessage   string     `gorm:"column:error_message"`
	ExpectedOutput string     `gorm:"column:expected_output"`
	UserOutput     string     `gorm:"user_output"`
}
