package dao

import (
	"FanCode/models/po"
	"github.com/jinzhu/gorm"
)

func InsertSubmission(db *gorm.DB, submission *po.Submission) error {
	return db.Create(submission).Error
}

func GetSubmissionListByUserIDAndProblemID(db *gorm.DB, userID uint, problemID uint) ([]*po.Submission, error) {
	var submissions []*po.Submission
	err := db.Where(`user_id = ? and problem_id = ?`, userID, problemID).Find(&submissions).Error
	return submissions, err
}
