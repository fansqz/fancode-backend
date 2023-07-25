package po

import "github.com/jinzhu/gorm"

// problem 结构
type Problem struct {
	gorm.Model
	Name        string `gorm:"column:name" json:"name"`
	Number      string `gorm:"column:number;unique_index" json:"number"`
	Description string `gorm:"column:description" json:"description"`
	Title       string `gorm:"column:title" json:"title"`
	// 存放题目内容的路径
	Path string `gorm:"column:path" json:"path"`
}
