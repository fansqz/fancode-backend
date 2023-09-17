package dao

import (
	"FanCode/models/po"
	"gorm.io/gorm"
)

// InsertProblem 添加题库
func InsertProblem(db *gorm.DB, problem *po.Problem) error {
	return db.Create(problem).Error
}

// GetProblemByNumber 根据题目编码获取题目
func GetProblemByNumber(db *gorm.DB, problemCode string) (*po.Problem, error) {
	question := &po.Problem{}
	err := db.Where("number = ?", problemCode).First(&question).Error
	return question, err
}

// GetProblemIDByNumber 根据题目number获取题目id
func GetProblemIDByNumber(db *gorm.DB, problemNumber string) (uint, error) {
	question := &po.Problem{}
	err := db.Where("number = ?", problemNumber).Select("id").Find(question).Error
	return question.ID, err
}

// GetProblemNameByID 根据题目id获取题目名称
func GetProblemNameByID(db *gorm.DB, problemID uint) (string, error) {
	question := &po.Problem{}
	err := db.Where("id = ?", problemID).Select("name").Find(question).Error
	return question.Name, err
}

// GetProblemByID 根据题目id获取题目
func GetProblemByID(db *gorm.DB, problemID uint) (*po.Problem, error) {
	question := &po.Problem{}
	err := db.First(&question, problemID).Error
	return question, err
}

// UpdateProblem 更新题目
// 不修改path
func UpdateProblem(db *gorm.DB, problem *po.Problem) error {
	return db.Model(&po.Problem{}).Where("id = ?", problem.ID).Updates(map[string]interface{}{
		"number":      problem.Number,
		"name":        problem.Name,
		"description": problem.Description,
		"difficulty":  problem.Difficulty,
		"title":       problem.Title,
		"languages":   problem.Languages,
		"enable":      problem.Enable,
	}).Error
}

// CheckProblemNumberExists 检测用户ID是否存在
func CheckProblemNumberExists(db *gorm.DB, problemCode string) (bool, error) {
	//执行
	row := db.Model(&po.Problem{}).Select("number").Where("number = ?", problemCode)
	if row.Error != nil {
		return false, row.Error
	}
	problem := &po.Problem{}
	row.Scan(&problem)
	return problem.Number != "", nil
}

// SetProblemEnable 让一个题目可用
func SetProblemEnable(db *gorm.DB, id uint, enable bool) error {
	return db.Model(&po.Problem{}).Where("id = ?", id).Update("enable", enable).Error
}

func GetProblemList(db *gorm.DB, page int, pageSize int, problem *po.Problem) ([]*po.Problem, error) {
	db2 := db
	if problem != nil && problem.Number != "" {
		db2 = db2.Where("number like ?", "%"+problem.Number+"%")
	}
	if problem != nil && problem.Name != "" {
		db2 = db2.Where("name like ?", "%"+problem.Name+"%")
	}
	if problem != nil && problem.Difficulty != nil {
		db2 = db2.Where("difficulty = ?", *problem.Difficulty)
	}
	if problem != nil && problem.Enable != nil {
		db2 = db2.Where("enable = ?", *problem.Enable)
	}
	offset := (page - 1) * pageSize
	var problems []*po.Problem
	err := db2.Limit(pageSize).Offset(offset).Find(&problems).Error
	return problems, err
}

func GetProblemCount(db *gorm.DB, problem *po.Problem) (int64, error) {
	var count int64
	db2 := db
	if problem != nil && problem.Name != "" {
		db2 = db2.Where("name like ?", "%"+problem.Name+"%")
	}
	if problem != nil && problem.Number != "" {
		db2 = db2.Where("number = ?", problem.Number)
	}
	if problem != nil && problem.Difficulty != nil {
		db2 = db2.Where("difficulty = ?", *problem.Difficulty)
	}
	if problem != nil && problem.Enable != nil {
		db2 = db2.Where("enable = ?", *problem.Enable)
	}
	err := db2.Model(&po.Problem{}).Count(&count).Error
	return count, err
}

func UpdatePathByCode(db *gorm.DB, path string, problemCode string) error {
	return db.Model(&po.Problem{}).
		Where("code = ?", problemCode).Update("path", path).Error
}

func UpdatePathByID(db *gorm.DB, path string, id uint) error {
	return db.Model(&po.Problem{}).
		Where("id = ?", id).Update("path", path).Error
}

func DeleteProblemByID(db *gorm.DB, id uint) error {
	return db.Delete(&po.Problem{}, id).Error
}

// UpdateProblemField 根据字段进行更新
func UpdateProblemField(db *gorm.DB, id uint, field string, value string) error {
	updateData := map[string]interface{}{
		field: value,
	}
	if err := db.Model(&po.Problem{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return err
	}
	return nil
}

// GetProblemFilePathByID 根据题目id获取题目文件的path
func GetProblemFilePathByID(db *gorm.DB, id uint) (string, error) {
	row := db.Model(&po.Problem{}).Select("path").Where("id = ?", id)
	if row.Error != nil {
		return "", row.Error
	}
	problem := &po.Problem{}
	row.Scan(&problem)
	return problem.Path, nil
}
