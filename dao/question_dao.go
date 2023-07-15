package dao

import (
	"FanCode/db"
	"FanCode/models"
)

// InsertQestion 添加题库
func InsertQuestion(question *models.Question) {
	db.DB.Create(question)
}

// GetQuestionByQuestioinNumber
func GetQuestionByQuestionNumber(questionNumber string) (*models.Question, error) {
	//写sql语句
	sqlStr := `select id,name,number,description,title,path
	from questions where number = ?`
	//执行
	row := db.DB.Raw(sqlStr, questionNumber)
	question := &models.Question{}
	row.Scan(&question)
	return question, nil
}

// UpdateQuestion 更新题目
func UpdateQuestion(question *models.Question) error {
	sqlStr := "update `questions` set name = ?, number = ?, discriptioin = ?, title = ?, path = ? where id = ?"
	//执行
	db.DB.Exec(sqlStr, question.Name, question.Number, question.Description, question.Title, question.Path, question.ID)
	return nil
}

// CheckUserID检测用户ID是否存在
func CheckQuestionNumber(questionNumber string) bool {
	//执行
	row := db.DB.Model(&models.User{}).Select("number").Where("number = ?", questionNumber)
	question := &models.Question{}
	row.Scan(&question)
	return question.Number != ""
}
