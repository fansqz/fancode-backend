package dao

import (
	"FanCode/db"
	"FanCode/models/po"
)

// InsertProblem 添加题库
func InsertProblem(problem *po.Problem) error {
	return db.DB.Create(problem).Error
}

// GetProblemByProblemCode
func GetProblemByProblemCode(problemCode string) (*po.Problem, error) {
	question := &po.Problem{}
	err := db.DB.Where("code = ?", problemCode).First(&question).Error
	return question, err
}

// GetProblemByProblemID
func GetProblemByProblemID(problemID uint) (*po.Problem, error) {
	question := &po.Problem{}
	err := db.DB.First(&question, problemID).Error
	return question, err
}

// UpdateProblem 更新题目
func UpdateProblem(problem *po.Problem) error {
	return db.DB.Save(&problem).Error
}

// CheckProblemCodeExists 检测用户ID是否存在
func CheckProblemCodeExists(problemCode string) (bool, error) {
	//执行
	row := db.DB.Model(&po.Problem{}).Select("code").Where("code = ?", problemCode)
	if row.Error != nil {
		return false, row.Error
	}
	problem := &po.Problem{}
	row.Scan(&problem)
	return problem.Code != "", nil
}

// SetProblemEnable 让一个题目可用
func SetProblemEnable(id uint, enable bool) error {
	return db.DB.Model(&po.Problem{}).Where("id = ?", id).Update("enable", enable).Error
}

func GetProblemList(page int, pageSize int) ([]*po.Problem, error) {
	offset := (page - 1) * pageSize
	var problems []*po.Problem
	err := db.DB.Limit(pageSize).Offset(offset).Find(&problems).Error
	return problems, err
}

func UpdatePathByCode(path string, problemCode string) error {
	return db.DB.Model(&po.Problem{}).
		Where("code = ?", problemCode).Update("path", path).Error
}

func UpdatePathByID(path string, id uint) error {
	return db.DB.Model(&po.Problem{}).
		Where("id = ?", id).Update("path", path).Error
}

func DeleteProblemByID(id uint) error {
	return db.DB.Delete(&po.Problem{}, id).Error
}

func GetProblemCount() (uint, error) {
	var count uint
	err := db.DB.Model(&po.Problem{}).Count(&count).Error
	return count, err
}

// UpdateProblemField 根据字段进行更新
func UpdateProblemField(id uint, field string, value string) error {
	updateData := map[string]interface{}{
		field: value,
	}
	if err := db.DB.Model(&po.Problem{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return err
	}
	return nil
}

// GetProblemFilePathByID 根据题目id获取题目文件的path
func GetProblemFilePathByID(id uint) (string, error) {
	row := db.DB.Model(&po.Problem{}).Select("path").Where("id = ?", id)
	if row.Error != nil {
		return "", row.Error
	}
	problem := &po.Problem{}
	row.Scan(&problem)
	return problem.Path, nil
}
