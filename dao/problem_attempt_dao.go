package dao

import (
	"FanCode/models/po"
	"gorm.io/gorm"
)

// ProblemAttemptDao
// 保存用户在一道题目中的做题情况。成功几次失败几次等。
type ProblemAttemptDao interface {
	InsertProblemAttempt(db *gorm.DB, problemAttempt *po.ProblemAttempt) error
	UpdateProblemAttempt(db *gorm.DB, problemAttempt *po.ProblemAttempt) error
	GetProblemAttemptByID(db *gorm.DB, userId uint, problemId uint) (*po.ProblemAttempt, error)
	GetProblemAttemptStatus(db *gorm.DB, userId uint, problemID uint) (int, error)
}

type problemAttemptDao struct {
}

func NewProblemAttemptDao() ProblemAttemptDao {
	return &problemAttemptDao{}
}

func (p *problemAttemptDao) InsertProblemAttempt(db *gorm.DB, problemAttempt *po.ProblemAttempt) error {
	return db.Create(problemAttempt).Error
}

func (p *problemAttemptDao) UpdateProblemAttempt(db *gorm.DB, problemAttempt *po.ProblemAttempt) error {
	return db.Model(problemAttempt).UpdateColumns(map[string]interface{}{
		"submission_count": problemAttempt.SubmissionCount,
		"success_count":    problemAttempt.SuccessCount,
		"err_count":        problemAttempt.ErrCount,
		"code":             problemAttempt.Code,
		"status":           problemAttempt.Status,
		"updated_at":       problemAttempt.UpdatedAt,
	}).Error
}

func (p *problemAttemptDao) GetProblemAttemptByID(db *gorm.DB, userId uint, problemId uint) (*po.ProblemAttempt, error) {
	problemAttempt := po.ProblemAttempt{}
	err := db.Model(&po.ProblemAttempt{}).Where("user_id = ? and problem_id = ?", userId, problemId).
		First(&problemAttempt).Error
	return &problemAttempt, err
}

func (p *problemAttemptDao) GetProblemAttemptStatus(db *gorm.DB, userId uint, problemID uint) (int, error) {
	var problemAttempt po.ProblemAttempt
	err := db.Model(&po.ProblemAttempt{}).Select("status", "id").
		Where("user_id = ? and problem_id = ?", userId, problemID).First(&problemAttempt).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, err
	}
	return problemAttempt.Status, nil
}
