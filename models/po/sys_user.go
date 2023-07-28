package po

import (
	"github.com/jinzhu/gorm"
)

type SysUser struct {
	gorm.Model
	LoginName string    `gorm:"column:login_name" json:"loginName"`
	Password  string    `gorm:"column:password" json:"password"`
	Username  string    `gorm:"column:username" json:"username"`
	Email     string    `gorm:"column:email" json:"email"`
	Phone     string    `gorm:"column:phone" json:"phone"`
	Avatar    string    `gorm:"column:avatar" json:"avatar"`
	Remark    string    `gorm:"column:remark" json:"remark"`
	Enable    int       `gorm:"column:enable" json:"enable"`
	Roles     []SysRole `gorm:"many2many:user_roles;" json:"roles"`
}
