package dao

import (
	"FanCode/models/po"
	"gorm.io/gorm"
	"time"
)

func InsertSubmission(db *gorm.DB, submission *po.Submission) error {
	return db.Create(submission).Error
}

func GetLastSubmission(db *gorm.DB, userID uint, problemID uint) (*po.Submission, error) {
	var submission *po.Submission
	err := db.Where("user_id = ? and problem_id = ?", userID, problemID).Last(submission).Error
	return submission, err
}

func GetSubmissionList(db *gorm.DB, page int, pageSize int, submission *po.Submission) ([]*po.Submission, error) {
	var submissions []*po.Submission
	db2 := db
	if submission.UserID != 0 {
		db2 = db2.Where("user_id = ?", submission.UserID)
	}
	if submission.ProblemID != 0 {
		db2 = db2.Where("problem_id = ?", submission.ProblemID)
	}
	offset := (page - 1) * pageSize
	err := db2.Limit(pageSize).Offset(offset).Find(&submissions).Error
	return submissions, err
}

func GetSubmissionCount(db *gorm.DB, submission *po.Submission) (int64, error) {
	var count int64
	db2 := db
	if submission != nil && submission.UserID != 0 {
		db2 = db2.Where("user_id = ?", submission.UserID)
	}
	if submission != nil && submission.ProblemID != 0 {
		db2 = db2.Where("problem_id = ?", submission.ProblemID)
	}
	err := db2.Model(&po.Submission{}).Count(&count).Error
	return count, err
}

func GetUserSimpleSubmissionsByTime(db *gorm.DB, userID uint, begin time.Time, end time.Time) ([]*po.Submission, error) {
	var submissions []*po.Submission
	err := db.Where("user_id = ? and created_at >= ? and created_at <= ?", userID, begin, end).
		Select("created_at").Find(&submissions).Error
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
