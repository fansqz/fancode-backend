package dao

import (
	"FanCode/models/dto"
	"FanCode/models/po"
	"gorm.io/gorm"
)

type ProblemBankDao interface {
	// InsertProblemBank 添加题库
	InsertProblemBank(db *gorm.DB, bank *po.ProblemBank) error
	// GetProblemBankByID 根据题库id获取题库
	GetProblemBankByID(db *gorm.DB, bankID uint) (*po.ProblemBank, error)
	// UpdateProblemBank 更新题库
	UpdateProblemBank(db *gorm.DB, bank *po.ProblemBank) error
	// DeleteProblemBankByID 删除题库
	DeleteProblemBankByID(db *gorm.DB, id uint) error
	// GetProblemBankCount 读取题库数量
	GetProblemBankCount(db *gorm.DB, problemBank *po.ProblemBank) (int64, error)
	// GetProblemBankList 获取题库列表
	GetProblemBankList(db *gorm.DB, pageQuery *dto.PageQuery) ([]*po.ProblemBank, error)
	// GetAllProblemBank 获取所有的题目数据
	GetAllProblemBank(db *gorm.DB) ([]*po.ProblemBank, error)
	// GetSimpleProblemBankList 获取题库列表，只包含id和名称
	GetSimpleProblemBankList(db *gorm.DB) ([]*po.ProblemBank, error)
}

type problemBankDao struct {
}

func NewProblemBankDao() ProblemBankDao {
	return &problemBankDao{}
}

func (p *problemBankDao) InsertProblemBank(db *gorm.DB, bank *po.ProblemBank) error {
	return db.Create(bank).Error
}

func (p *problemBankDao) GetProblemBankByID(db *gorm.DB, bankID uint) (*po.ProblemBank, error) {
	bank := &po.ProblemBank{}
	err := db.First(&bank, bankID).Error
	return bank, err
}

func (p *problemBankDao) UpdateProblemBank(db *gorm.DB, bank *po.ProblemBank) error {
	return db.Model(bank).Updates(bank).Error
}

func (p *problemBankDao) DeleteProblemBankByID(db *gorm.DB, id uint) error {
	return db.Delete(&po.ProblemBank{}, id).Error
}

func (p *problemBankDao) GetProblemBankCount(db *gorm.DB, problemBank *po.ProblemBank) (int64, error) {
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

func (p *problemBankDao) GetProblemBankList(db *gorm.DB, pageQuery *dto.PageQuery) ([]*po.ProblemBank, error) {
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

func (p *problemBankDao) GetAllProblemBank(db *gorm.DB) ([]*po.ProblemBank, error) {
	var banks []*po.ProblemBank
	err := db.Find(&banks).Error
	return banks, err
}

func (p *problemBankDao) GetSimpleProblemBankList(db *gorm.DB) ([]*po.ProblemBank, error) {
	var banks []*po.ProblemBank
	err := db.Select("id", "name").Find(&banks).Error
	return banks, err
}
