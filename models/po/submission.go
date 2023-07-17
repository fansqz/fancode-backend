package po

import "github.com/jinzhu/gorm"

type Submission struct {
	gorm.DB
	UserID uint
	QuestionID uint
	code: 用户提交的代码
	submission_time: 提交时间戳
	status: 运行结果状态，如Accepted（通过）、Wrong Answer（答案错误）、Time Limit Exceeded（超时）等
	score: 提交获得的得分
	language: 使用的编程语言
	error_message: 错误消息（如果有错误）
}
