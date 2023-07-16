package dao

import (
	"FanCode/db"
	"FanCode/models/po"
)

// InsertQestion 添加题库
func InsertQuestion(question *po.Question) {
	db.DB.Create(question)
}

// GetQuestionByQuestioinNumber
func GetQuestionByQuestionNumber(questionNumber string) (*po.Question, error) {
	//写sql语句
	sqlStr := `select id,name,number,description,title,path
	from questions where number = ?`
	//执行
	row := db.DB.Raw(sqlStr, questionNumber)
	question := &po.Question{}
	row.Scan(&question)
	return question, nil
}

// GetQuestionByQuestioinNumber
func GetQuestionByQuestionID(questionID uint) (*po.Question, error) {
	//写sql语句
	sqlStr := `select id,name,number,description,title,path
	from questions where id = ?`
	//执行
	row := db.DB.Raw(sqlStr, questionID)
	question := &po.Question{}
	row.Scan(&question)
	return question, nil
}

// UpdateQuestion 更新题目
func UpdateQuestion(question *po.Question) error {
	sqlStr := "update `questions` set name = ?, number = ?, discriptioin = ?, title = ?, path = ? where id = ?"
	//执行
	err := db.DB.Exec(sqlStr, question.Name, question.Number, question.Description, question.Title, question.Path, question.ID).Error
	return err
}

// CheckUserID检测用户ID是否存在
func CheckQuestionNumber(questionNumber string) bool {
	//执行
	row := db.DB.Model(&po.User{}).Select("number").Where("number = ?", questionNumber)
	question := &po.Question{}
	row.Scan(&question)
	return question.Number != ""
}

func GetQuestionList(page int, pageSize int) ([]*po.Question, error) {
	offset := (page - 1) * pageSize
	var questions []*po.Question
	err := db.DB.Limit(pageSize).Offset(offset).Find(&questions).Error
	return questions, err
}

func UpdatePathByNumber(path string, questionNumber string) error {
	sqlStr := "update `questions` set path = ? where number = ?"
	//执行
	err := db.DB.Exec(sqlStr, path, questionNumber).Error
	return err
}

func DeleteQuestionByID(id uint) error {
	return db.DB.Delete(&po.Question{}, id).Error
}
