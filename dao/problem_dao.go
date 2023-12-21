package dao

import (
	"FanCode/models/dto"
	"FanCode/models/po"
	"gorm.io/gorm"
)

type ProblemDao interface {
	// GetProblemByNumber 根据题目编码获取题目
	GetProblemByNumber(db *gorm.DB, problemCode string) (*po.Problem, error)
	// GetProblemIDByNumber 根据题目number获取题目id
	GetProblemIDByNumber(db *gorm.DB, problemNumber string) (uint, error)
	// GetProblemNameByID 根据题目id获取题目名称
	GetProblemNameByID(db *gorm.DB, problemID uint) (string, error)
	// GetProblemByID 根据题目id获取题目
	GetProblemByID(db *gorm.DB, problemID uint) (*po.Problem, error)
	// InsertProblem 添加题库
	InsertProblem(db *gorm.DB, problem *po.Problem) error
	// UpdateProblem 更新题目
	// 不修改path
	UpdateProblem(db *gorm.DB, problem *po.Problem) error
	// UpdateProblemField 根据字段进行更新
	UpdateProblemField(db *gorm.DB, id uint, field string, value string) error
	// CheckProblemNumberExists 检测用户ID是否存在
	CheckProblemNumberExists(db *gorm.DB, problemCode string) (bool, error)
	// SetProblemEnable 让一个题目可用
	SetProblemEnable(db *gorm.DB, id uint, enable int) error
	// DeleteProblemByID 删除题目
	DeleteProblemByID(db *gorm.DB, id uint) error
	GetProblemList(db *gorm.DB, pageQuery *dto.PageQuery) ([]*po.Problem, error)
	GetProblemCount(db *gorm.DB, problem *po.Problem) (int64, error)
}

type problemDao struct {
}

func NewProblemDao() ProblemDao {
	return &problemDao{}
}

func (p *problemDao) GetProblemByNumber(db *gorm.DB, problemCode string) (*po.Problem, error) {
	question := &po.Problem{}
	err := db.Where("number = ?", problemCode).First(&question).Error
	return question, err
}

func (p *problemDao) GetProblemIDByNumber(db *gorm.DB, problemNumber string) (uint, error) {
	question := &po.Problem{}
	err := db.Where("number = ?", problemNumber).Select("id").First(question).Error
	return question.ID, err
}

func (p *problemDao) GetProblemNameByID(db *gorm.DB, problemID uint) (string, error) {
	question := &po.Problem{}
	err := db.Where("id = ?", problemID).Select("name").Find(question).Error
	return question.Name, err
}

func (p *problemDao) GetProblemByID(db *gorm.DB, problemID uint) (*po.Problem, error) {
	question := &po.Problem{}
	err := db.First(&question, problemID).Error
	return question, err
}

func (p *problemDao) GetProblemList(db *gorm.DB, pageQuery *dto.PageQuery) ([]*po.Problem, error) {
	var problem *po.Problem
	if pageQuery.Query != nil {
		problem = pageQuery.Query.(*po.Problem)
	}
	if problem != nil && problem.Number != "" {
		db = db.Where("number like ?", "%"+problem.Number+"%")
	}
	if problem != nil && problem.Name != "" {
		db = db.Where("name like ?", "%"+problem.Name+"%")
	}
	if problem != nil && problem.Difficulty != 0 {
		db = db.Where("difficulty = ?", problem.Difficulty)
	}
	if problem != nil && problem.Enable != 0 {
		db = db.Where("enable = ?", problem.Enable)
	}
	if problem != nil && problem.BankID != nil {
		db = db.Where("bank_id = ?", problem.BankID)
	}
	offset := (pageQuery.Page - 1) * pageQuery.PageSize
	var problems []*po.Problem
	db = db.Offset(offset).Limit(pageQuery.PageSize)
	if pageQuery.SortProperty != "" && pageQuery.SortRule != "" {
		order := pageQuery.SortProperty + " " + pageQuery.SortRule
		db = db.Order(order)
	}
	err := db.Find(&problems).Error
	return problems, err
}

func (p *problemDao) GetProblemCount(db *gorm.DB, problem *po.Problem) (int64, error) {
	var count int64
	if problem != nil && problem.Name != "" {
		db = db.Where("name like ?", "%"+problem.Name+"%")
	}
	if problem != nil && problem.Number != "" {
		db = db.Where("number = ?", problem.Number)
	}
	if problem != nil && problem.Difficulty != 0 {
		db = db.Where("difficulty = ?", problem.Difficulty)
	}
	if problem != nil && problem.BankID != nil {
		db = db.Where("bank_id = ?", problem.BankID)
	}
	if problem != nil && problem.Enable != 0 {
		db = db.Where("enable = ?", problem.Enable)
	}
	err := db.Model(&po.Problem{}).Count(&count).Error
	return count, err
}

func (p *problemDao) InsertProblem(db *gorm.DB, problem *po.Problem) error {
	return db.Create(problem).Error
}

func (p *problemDao) UpdateProblem(db *gorm.DB, problem *po.Problem) error {
	return db.Model(&po.Problem{}).Where("id = ?", problem.ID).Updates(map[string]interface{}{
		"updated_at":  problem.UpdatedAt,
		"bank_id":     problem.BankID,
		"number":      problem.Number,
		"name":        problem.Name,
		"description": problem.Description,
		"difficulty":  problem.Difficulty,
		"title":       problem.Title,
		"languages":   problem.Languages,
		"enable":      problem.Enable,
	}).Error
}

func (p *problemDao) UpdateProblemField(db *gorm.DB, id uint, field string, value string) error {
	updateData := map[string]interface{}{
		field: value,
	}
	if err := db.Model(&po.Problem{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return err
	}
	return nil
}

func (p *problemDao) CheckProblemNumberExists(db *gorm.DB, problemCode string) (bool, error) {
	//执行
	row := db.Model(&po.Problem{}).Select("number").Where("number = ?", problemCode)
	if row.Error != nil {
		return false, row.Error
	}
	problem := &po.Problem{}
	row.Scan(&problem)
	return problem.Number != "", nil
}

func (p *problemDao) SetProblemEnable(db *gorm.DB, id uint, enable int) error {
	return db.Model(&po.Problem{}).Where("id = ?", id).Update("enable", enable).Error
}

func (p *problemDao) DeleteProblemByID(db *gorm.DB, id uint) error {
	return db.Delete(&po.Problem{}, id).Error
}
