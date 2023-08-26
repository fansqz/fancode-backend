package dao

import (
	"FanCode/models/po"
	"gorm.io/gorm"
)

func InsertProblemAttempt(db *gorm.DB, problemAttempt *po.ProblemAttempt) error {
	return db.Create(problemAttempt).Error
}

func UpdateProblemAttempt(db *gorm.DB, problemAttempt *po.ProblemAttempt) error {
	return db.Model(problemAttempt).UpdateColumns(map[string]interface{}{
		"submission_count": problemAttempt.SubmissionCount,
		"success_count":    problemAttempt.SuccessCount,
		"err_count":        problemAttempt.ErrCount,
		"code":             problemAttempt.Code,
		"state":            problemAttempt.State,
		"updated_at":       problemAttempt.UpdatedAt,
	}).Error
}

func GetProblemAttempt(db *gorm.DB, userId uint, problemId uint) (*po.ProblemAttempt, error) {
	var problemAttempt po.ProblemAttempt
	err := db.Model(&po.ProblemAttempt{}).Where("user_id = ? and problem_id = ?", userId, problemId).
		Find(&problemAttempt).Error
	return &problemAttempt, err
}

func GetProblemAttemptState(db *gorm.DB, userId uint, problemID uint) (int, error) {
	var problemAttempt po.ProblemAttempt
	err := db.Model(&po.ProblemAttempt{}).Select("state").
		Where("user_id = ? and problem_id = ?", userId, problemID).Find(&problemAttempt).Error
	if err != nil {
		return 0, err
	}
	return problemAttempt.State, nil
}
