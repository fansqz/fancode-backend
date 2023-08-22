package po

import "gorm.io/gorm"

// problem 结构
type Problem struct {
	gorm.Model
	Name        string `gorm:"column:name" json:"name"`
	Number      string `gorm:"column:number;unique_index" json:"number"`
	Description string `gorm:"column:description;type:text" json:"description"`
	Title       string `gorm:"column:title" json:"title"`
	// 存放题目内容的路径
	Path       string `gorm:"column:path" json:"path"`
	Difficulty *int   `gorm:"column:difficulty" json:"difficulty"`
	Enable     *bool  `gorm:"column:enable" json:"enable"`
}
