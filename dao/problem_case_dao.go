package dao

import (
	"FanCode/models/dto"
	"FanCode/models/po"
	"gorm.io/gorm"
)

type ProblemCaseDao interface {
	// GetProblemCaseList 获取用例列表
	GetProblemCaseList(db *gorm.DB, query *dto.PageQuery) ([]*po.ProblemCase, error)
	GetProblemCaseList2(db *gorm.DB, problemID uint) ([]*po.ProblemCase, error)
	// GetProblemCaseCount 获取用例数量
	GetProblemCaseCount(db *gorm.DB, problemCase *po.ProblemCase) (int64, error)
	// GetProblemCaseByID 通过id获取题目用例
	GetProblemCaseByID(db *gorm.DB, id uint) (*po.ProblemCase, error)
	// DeleteProblemCaseByID 通过id删除题目用例
	DeleteProblemCaseByID(db *gorm.DB, id uint) error
	// DeleteProblemCaseByProblemID 通过题目id删除题目用例
	DeleteProblemCaseByProblemID(db *gorm.DB, problemID uint) error
	// InsertProblemCase 添加题目用例
	InsertProblemCase(db *gorm.DB, problemCase *po.ProblemCase) error
	// UpdateProblemCase 更新题目用例
	UpdateProblemCase(db *gorm.DB, problemCase *po.ProblemCase) error
}

type problemCaseDao struct {
}

func NewProblemCaseDao() ProblemCaseDao {
	return &problemCaseDao{}
}

func (p *problemCaseDao) GetProblemCaseList(db *gorm.DB, query *dto.PageQuery) ([]*po.ProblemCase, error) {
	var problemCase *po.ProblemCase
	if query.Query != nil {
		problemCase = query.Query.(*po.ProblemCase)
	}
	if problemCase != nil && problemCase.ProblemID != 0 {
		db = db.Where("problem_id = ?", problemCase.ProblemID)
	}
	if problemCase != nil && problemCase.Name != "" {
		db = db.Where("name like ?", "%"+problemCase.Name+"%")
	}
	offset := (query.Page - 1) * query.PageSize
	var cases []*po.ProblemCase
	db = db.Offset(offset).Limit(query.PageSize)
	if query.SortProperty != "" && query.SortRule != "" {
		order := query.SortProperty + " " + query.SortRule
		db = db.Order(order)
	}
	err := db.Find(&cases).Error
	return cases, err
}

func (p *problemCaseDao) GetProblemCaseList2(db *gorm.DB, problemID uint) ([]*po.ProblemCase, error) {
	db = db.Where("problem_id = ?", problemID)
	var cases []*po.ProblemCase
	err := db.Find(&cases).Error
	return cases, err
}

func (p *problemCaseDao) GetProblemCaseCount(db *gorm.DB, problemCase *po.ProblemCase) (int64, error) {
	var count int64
	if problemCase != nil && problemCase.ProblemID != 0 {
		db = db.Where("problem_id = ?", problemCase.ProblemID)
	}
	if problemCase != nil && problemCase.Name != "" {
		db = db.Where("name like ?", "%"+problemCase.Name+"%")
	}
	err := db.Model(&po.ProblemCase{}).Count(&count).Error
	return count, err
}

func (p *problemCaseDao) GetProblemCaseByID(db *gorm.DB, id uint) (*po.ProblemCase, error) {
	problemCase := &po.ProblemCase{}
	err := db.First(&problemCase, id).Error
	return problemCase, err
}

func (p *problemCaseDao) DeleteProblemCaseByID(db *gorm.DB, id uint) error {
	return db.Delete(&po.ProblemCase{}, id).Error
}

func (p *problemCaseDao) DeleteProblemCaseByProblemID(db *gorm.DB, problemID uint) error {
	return db.Where("problem_id = ?", problemID).Delete(&po.ProblemCase{}).Error
}

func (p *problemCaseDao) InsertProblemCase(db *gorm.DB, problemCase *po.ProblemCase) error {
	return db.Create(problemCase).Error
}

func (p *problemCaseDao) UpdateProblemCase(db *gorm.DB, problemCase *po.ProblemCase) error {
	return db.Model(problemCase).Updates(problemCase).Error
}
