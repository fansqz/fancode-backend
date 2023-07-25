package po

import "github.com/jinzhu/gorm"

// User 结构体
type User struct {
	gorm.Model
	Username string `gorm:"column:username" json:"username"`
	Password string `gorm:"column:password" json:"password"`
	Email    string `gorm:"column:email" json:"email"`
	Phone    string `gorm:"column:phone" json:"phone"`
	Sex      int    `gorm:"column:sex" json:"sex"`
	Role     int    `gorm:"column:role" json:"role"`
	Code     string `gorm:"column:code" json:"code"`
}
