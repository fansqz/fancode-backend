package po

import "github.com/jinzhu/gorm"

// problem 结构
type Problem struct {
	gorm.Model
	Name        string `gorm:"column:name" json:"name"`
	Code        string `gorm:"column:code;unique_index" json:"code"`
	Description string `gorm:"column:description;type:text" json:"description"`
	Title       string `gorm:"column:title" json:"title"`
	// 存放题目内容的路径
	Path       string `gorm:"column:path" json:"path"`
	Difficulty int    `gorm:"column:difficultly" json:"difficulty"`
	Enable     bool   `gorm:"column:enable" json:"enable"`
}
