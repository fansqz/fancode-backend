package dao

import (
	"FanCode/db"
	"FanCode/models/po"
)

func InsertSubmission(submission *po.Submission) error {
	return db.DB.Create(submission).Error
}

func GetSubmissionListByUserIDAndProblemID(userID uint, problemID uint) ([]*po.Submission, error) {
	var submissions []*po.Submission
	err := db.DB.Where(`user_id = ? and problem_id = ?`, userID, problemID).Find(&submissions).Error
	return submissions, err
}
