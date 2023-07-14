package models

// question 结构提
type Question struct {
	ID             int    `gorm:"primary_key;AUTO_INCREMENT;column:id" json:"id"`
	QuestionName   string `gorm:"column:question_name" json:"questionName"`
	QuestionNumber string `gorm:"column:question_Number" json:"questionNumber"`
	Description    string `gorm:"column:description" json:"description"`
	Title          string `gorm:"column:title" json:"title"`
	// 存放题目内容的路径
	Path string `gorm:"column:path" json:"path"`
}
