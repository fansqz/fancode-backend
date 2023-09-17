package po

import (
	"gorm.io/gorm"
	"time"
)

type Submission struct {
	gorm.Model
	// 用户id
	UserID uint `gorm:"column:user_id"`
	// 题目id
	ProblemID uint `gorm:"column:problem_id"`
	// 使用的编程语言
	Language string `gorm:"column:language"`
	// 用户代码
	Code string `gorm:"column:code"`
	// 状态
	Status int `gorm:"column:status"`
	// 异常信息
	ErrorMessage string `gorm:"column:error_message"`
	// 期望输出
	ExpectedOutput string `gorm:"column:expected_output"`
	// 用户输出
	UserOutput string        `gorm:"user_output"`
	TimeUsed   time.Duration // 判题使用时间
	MemoryUsed int64         // 内存使用量（以字节为单位）

}
