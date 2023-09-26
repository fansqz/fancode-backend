package po

import (
	"gorm.io/gorm"
)

// Problem 结构
type Problem struct {
	gorm.Model
	CreatorID   uint   `gorm:"column:creator_id" json:"creatorID"`
	Name        string `gorm:"column:name" json:"name"`
	Number      string `gorm:"column:number;type:varchar(255);unique_index:idx_number" json:"number"`
	Description string `gorm:"column:description;type:text" json:"description"`
	Title       string `gorm:"column:title" json:"title"`
	// 存放题目内容的路径
	Path       string `gorm:"column:path" json:"path"`
	Difficulty *int   `gorm:"column:difficulty" json:"difficulty"`
	Enable     *bool  `gorm:"column:enable" json:"enable"`
	// 支持的语言用,分割
	Languages string `gorm:"column:languages" json:"languages"`
	// 所属题库id
	BankID string `gorm:"column:bank_id" json:"bankID"`
}
