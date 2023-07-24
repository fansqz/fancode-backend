package dao

import (
	"FanCode/db"
	"FanCode/models/po"
)

// InsertProblem 添加题库
func InsertProblem(problem *po.Problem) {
	db.DB.Create(problem)
}

// GetProblemByProblemNumber
func GetProblemByProblemNumber(problemNumber string) (*po.Problem, error) {
	//写sql语句
	sqlStr := `select id,name,number,description,title,path
	from problems where number = ?`
	//执行
	row := db.DB.Raw(sqlStr, problemNumber)
	problem := &po.Problem{}
	row.Scan(&problem)
	return problem, nil
}

// GetProblemByProblemID
func GetProblemByProblemID(problemID uint) (*po.Problem, error) {
	//写sql语句
	sqlStr := `select id,name,number,description,title,path
	from problems where id = ?`
	//执行
	row := db.DB.Raw(sqlStr, problemID)
	question := &po.Problem{}
	row.Scan(&question)
	return question, nil
}

// UpdateProblem 更新题目
func UpdateProblem(problem *po.Problem) error {
	sqlStr := "update `problems` set name = ?, number = ?, discriptioin = ?, title = ?, path = ? where id = ?"
	//执行
	err := db.DB.Exec(sqlStr, problem.Name, problem.Number, problem.Description, problem.Title, problem.Path, problem.ID).Error
	return err
}

// CheckUserID检测用户ID是否存在
func CheckProblemNumber(problemNumber string) bool {
	//执行
	row := db.DB.Model(&po.User{}).Select("number").Where("number = ?", problemNumber)
	problem := &po.Problem{}
	row.Scan(&problem)
	return problem.Number != ""
}

func GetProblemList(page int, pageSize int) ([]*po.Problem, error) {
	offset := (page - 1) * pageSize
	var problems []*po.Problem
	err := db.DB.Limit(pageSize).Offset(offset).Find(&problems).Error
	return problems, err
}

func UpdatePathByNumber(path string, problemNumber string) error {
	sqlStr := "update `problems` set path = ? where number = ?"
	//执行
	err := db.DB.Exec(sqlStr, path, problemNumber).Error
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
