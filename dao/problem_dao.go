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
	//写sql语句
	sqlStr := `select id,name,code,description,title,path
	from problems where code = ?`
	//执行
	row := db.DB.Raw(sqlStr, problemCode)
	problem := &po.Problem{}
	row.Scan(&problem)
	return problem, nil
}

// GetProblemByProblemID
func GetProblemByProblemID(problemID uint) (*po.Problem, error) {
	//写sql语句
	sqlStr := `select id,name,code,description,title,path
	from problems where id = ?`
	//执行
	row := db.DB.Raw(sqlStr, problemID)
	question := &po.Problem{}
	row.Scan(&question)
	return question, nil
}

// UpdateProblem 更新题目
func UpdateProblem(problem *po.Problem) error {
	sqlStr := "update `problems` set name = ?, code = ?, description = ?, title = ? where id = ?"
	//执行
	err := db.DB.Exec(sqlStr, problem.Name, problem.Code, problem.Description, problem.Title, problem.ID).Error
	return err
}

// CheckUserID检测用户ID是否存在
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

func GetProblemList(page int, pageSize int) ([]*po.Problem, error) {
	offset := (page - 1) * pageSize
	var problems []*po.Problem
	err := db.DB.Limit(pageSize).Offset(offset).Find(&problems).Error
	return problems, err
}

func UpdatePathByCode(path string, problemCode string) error {
	sqlStr := "update `problems` set path = ? where code = ?"
	//执行
	err := db.DB.Exec(sqlStr, path, problemCode).Error
	return err
}

func UpdatePathByID(path string, id uint) error {
	sqlStr := "update `problems` set path = ? where id = ?"
	//执行
	err := db.DB.Exec(sqlStr, path, id).Error
	return err
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
