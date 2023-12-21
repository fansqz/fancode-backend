package po

import "gorm.io/gorm"

// ProblemCase
// 表示一道题目的一个用例
type ProblemCase struct {
	gorm.Model
	ProblemID uint   `gorm:"column:problem_id" json:"problemID"`
	Name      string `gorm:"column:case_name"`
	Input     string `gorm:"column:input"`
	Output    string `gorm:"column:output"`
}
