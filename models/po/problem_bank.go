package po

import "gorm.io/gorm"

// ProblemBank  题库
type ProblemBank struct {
	gorm.Model
	Name        string    `gorm:"column:name"`
	Icon        string    `gorm:"column:icon"`
	Description string    `gorm:"column:description"`
	CreatorID   string    `gorm:"column:creator_id"`
	Problems    []Problem `gorm:"foreignKey:id"`
}
