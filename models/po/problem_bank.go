package po

import "gorm.io/gorm"

// ProblemBank  题库
type ProblemBank struct {
	gorm.Model
	Name        string    `gorm:"column:name"`
	Icon        string    `gorm:"column:icon"`
	Description string    `gorm:"column:description"`
	CreatorID   uint      `gorm:"column:creator_id"`
	Problems    []Problem `gorm:"foreignKey:bank_id"`
}
