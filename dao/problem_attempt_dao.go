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
