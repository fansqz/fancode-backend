package models

// User 结构体
type User struct {
	ID       int    `gorm:"primary_key;AUTO_INCREMENT;column:id" json:"id"`
	Username string `gorm:"column:username" json:"username"`
	Password string `gorm:"column:password" json:"password"`
	Email    string `gorm:"column:email" json:"email"`
	Phone    string `gorm:"column:phone" json:"phone"`
	Sex      int    `gorm:"column:sex" json:"sex"`
	Role     int    `gorm:"column:role" json:"role"`
	UserID   int    `gorm:"column:user_id" json:"userID"`
}
