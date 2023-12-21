package dao

import (
	"FanCode/models/dto"
	"FanCode/models/po"
	"gorm.io/gorm"
	"time"
)

type SubmissionDao interface {
	GetLastSubmission(db *gorm.DB, userID uint, problemID uint) (*po.Submission, error)
	GetSubmissionList(db *gorm.DB, pageQuery *dto.PageQuery) ([]*po.Submission, error)
	GetSubmissionCount(db *gorm.DB, submission *po.Submission) (int64, error)
	GetUserSimpleSubmissionsByTime(db *gorm.DB, userID uint, begin time.Time, end time.Time) ([]*po.Submission, error)
	CheckUserIsSubmittedByTime(db *gorm.DB, userID uint, begin time.Time, end time.Time) (bool, error)
	InsertSubmission(db *gorm.DB, submission *po.Submission) error
}

type submissionDao struct {
}

func NewSubmissionDao() SubmissionDao {
	return &submissionDao{}
}

func (s *submissionDao) GetLastSubmission(db *gorm.DB, userID uint, problemID uint) (*po.Submission, error) {
	var submission *po.Submission
	err := db.Where("user_id = ? and problem_id = ?", userID, problemID).Last(submission).Error
	return submission, err
}

func (s *submissionDao) GetSubmissionList(db *gorm.DB, pageQuery *dto.PageQuery) ([]*po.Submission, error) {
	submission := pageQuery.Query.(*po.Submission)
	var submissions []*po.Submission
	if submission.UserID != 0 {
		db = db.Where("user_id = ?", submission.UserID)
	}
	if submission.ProblemID != 0 {
		db = db.Where("problem_id = ?", submission.ProblemID)
	}
	offset := (pageQuery.Page - 1) * pageQuery.PageSize
	err := db.Limit(pageQuery.PageSize).Offset(offset).Find(&submissions).Error
	return submissions, err
}

func (s *submissionDao) GetSubmissionCount(db *gorm.DB, submission *po.Submission) (int64, error) {
	var count int64
	if submission != nil && submission.UserID != 0 {
		db = db.Where("user_id = ?", submission.UserID)
	}
	if submission != nil && submission.ProblemID != 0 {
		db = db.Where("problem_id = ?", submission.ProblemID)
	}
	err := db.Model(&po.Submission{}).Count(&count).Error
	return count, err
}

func (s *submissionDao) GetUserSimpleSubmissionsByTime(db *gorm.DB, userID uint, begin time.Time, end time.Time) ([]*po.Submission, error) {
	var submissions []*po.Submission
	err := db.Where("user_id = ? and created_at >= ? and created_at <= ?", userID, begin, end).
		Select("created_at").Find(&submissions).Error
	if err != nil {
		return nil, err
	}
	return submissions, err
}

func (s *submissionDao) CheckUserIsSubmittedByTime(db *gorm.DB, userID uint, begin time.Time, end time.Time) (bool, error) {
	submission := po.Submission{}
	submission.UserID = userID
	data := &po.Submission{}
	if err := db.Model(&submission).Where("created_at >= ? and created_at <= ?", begin, end).Take(data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func (s *submissionDao) InsertSubmission(db *gorm.DB, submission *po.Submission) error {
	return db.Create(submission).Error
}
