package po

import (
	"github.com/jinzhu/gorm"
	"time"
)

type Submission struct {
	gorm.DB
	UserID         uint
	QuestionID     uint
	Code           string
	SubmissionTime *time.Time
	Status         int
	ErrorMessage   string
}
