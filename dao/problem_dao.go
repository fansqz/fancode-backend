package dao

import (
	"FanCode/models/po"
	"github.com/jinzhu/gorm"
)

// InsertProblem 添加题库
func InsertProblem(db *gorm.DB, problem *po.Problem) error {
	return db.Create(problem).Error
}

// GetProblemByProblemCode
func GetProblemByProblemCode(db *gorm.DB, problemCode string) (*po.Problem, error) {
	question := &po.Problem{}
	err := db.Where("code = ?", problemCode).First(&question).Error
	return question, err
}

// GetProblemByProblemID
func GetProblemByProblemID(db *gorm.DB, problemID uint) (*po.Problem, error) {
	question := &po.Problem{}
	err := db.First(&question, problemID).Error
	return question, err
}

// UpdateProblem 更新题目
func UpdateProblem(db *gorm.DB, problem *po.Problem) error {
	return db.Save(&problem).Error
}

// CheckProblemCodeExists 检测用户ID是否存在
func CheckProblemCodeExists(db *gorm.DB, problemCode string) (bool, error) {
	//执行
	row := db.Model(&po.Problem{}).Select("code").Where("code = ?", problemCode)
	if row.Error != nil {
		return false, row.Error
	}
	problem := &po.Problem{}
	row.Scan(&problem)
	return problem.Code != "", nil
}

// SetProblemEnable 让一个题目可用
func SetProblemEnable(db *gorm.DB, id uint, enable bool) error {
	return db.Model(&po.Problem{}).Where("id = ?", id).Update("enable", enable).Error
}

func GetProblemList(db *gorm.DB, page int, pageSize int) ([]*po.Problem, error) {
	offset := (page - 1) * pageSize
	var problems []*po.Problem
	err := db.Limit(pageSize).Offset(offset).Find(&problems).Error
	return problems, err
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

func GetProblemCount(db *gorm.DB) (uint, error) {
	var count uint
	err := db.Model(&po.Problem{}).Count(&count).Error
	return count, err
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
