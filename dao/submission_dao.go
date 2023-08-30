package dao

import (
	"FanCode/models/po"
	"gorm.io/gorm"
	"time"
)

func InsertSubmission(db *gorm.DB, submission *po.Submission) error {
	return db.Create(submission).Error
}

func GetSubmissionListByUserIDAndProblemID(db *gorm.DB, userID uint, problemID uint) ([]*po.Submission, error) {
	var submissions []*po.Submission
	err := db.Where(`user_id = ? and problem_id = ?`, userID, problemID).Find(&submissions).Error
	return submissions, err
}

func GetUserSimpleSubmissionsByTime(db *gorm.DB, userID uint, begin time.Time, end time.Time) ([]*po.Submission, error) {
	submission := po.Submission{}
	submission.UserID = userID
	var submissions []*po.Submission
	err := db.Model(&submission).Where("created_at >= ? and created_at <= ?", begin, end).
		Select("created_at").Find(submissions).Error
	if err != nil {
		return nil, err
	}
	return submissions, err
}

func CheckUserIsSubmittedByTime(db *gorm.DB, userID uint, begin time.Time, end time.Time) (bool, error) {
	submission := po.Submission{}
	submission.UserID = userID
	data := &po.Submission{}
	err := db.Model(&submission).Where("created_at >= ? and created_at <= ?", begin, end).Take(data).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}
