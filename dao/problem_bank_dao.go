package dao

import (
	"FanCode/models/po"
	"gorm.io/gorm"
)

// InsertProblemBank 添加题库
func InsertProblemBank(db *gorm.DB, bank *po.ProblemBank) error {
	return db.Create(bank).Error
}

// GetProblemBankByID 根据题库id获取题库
func GetProblemBankByID(db *gorm.DB, bankID uint) (*po.ProblemBank, error) {
	bank := &po.ProblemBank{}
	err := db.First(&bank, bankID).Error
	return bank, err
}

// UpdateProblemBank 更新题库
func UpdateProblemBank(db *gorm.DB, bank *po.ProblemBank) error {
	return db.Model(&po.ProblemBank{}).Where("id = ?", bank.ID).Updates(map[string]interface{}{
		"name":        bank.Name,
		"icon":        bank.Icon,
		"description": bank.Description,
		"creator_id":  bank.CreatorID,
	}).Error
}

// DeleteProblemBankByID 删除题库
func DeleteProblemBankByID(db *gorm.DB, id uint) error {
	return db.Delete(&po.ProblemBank{}, id).Error
}
