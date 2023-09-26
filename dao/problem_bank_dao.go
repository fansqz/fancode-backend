package dao

import (
	"FanCode/models/dto"
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

// GetProblemBankCount 读取题库数量
func GetProblemBankCount(db *gorm.DB, problemBank *po.ProblemBank) (int64, error) {
	var count int64
	db2 := db
	if problemBank != nil && problemBank.Name != "" {
		db2 = db2.Where("name like ?", "%"+problemBank.Name+"%")
	}
	if problemBank != nil && problemBank.Description != "" {
		db2 = db2.Where("description = ?", problemBank.Description)
	}
	err := db2.Model(&po.ProblemBank{}).Count(&count).Error
	return count, err
}

// GetProblemBankList 获取题库列表
func GetProblemBankList(db *gorm.DB, pageQuery *dto.PageQuery) ([]*po.ProblemBank, error) {
	var problemBank *po.ProblemBank
	if pageQuery.Query != nil {
		problemBank = pageQuery.Query.(*po.ProblemBank)
	}
	if problemBank != nil && problemBank.Name != "" {
		db = db.Where("name like ?", "%"+problemBank.Name+"%")
	}
	if problemBank != nil && problemBank.Description != "" {
		db = db.Where("description like ?", "%"+problemBank.Description+"%")
	}
	offset := (pageQuery.Page - 1) * pageQuery.PageSize
	var banks []*po.ProblemBank
	db = db.Offset(offset).Limit(pageQuery.PageSize)
	if pageQuery.SortProperty != "" && pageQuery.SortRule != "" {
		order := pageQuery.SortProperty + " " + pageQuery.SortRule
		db = db.Order(order)
	}
	err := db.Find(&banks).Error
	return banks, err
}
