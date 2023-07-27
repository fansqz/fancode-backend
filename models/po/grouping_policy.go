package po

import "github.com/jinzhu/gorm"

type GroupingPolicy struct {
	gorm.Model        // 如果使用 GORM，可以嵌入 gorm.Model
	UserID     string `gorm:"column:user_id"` // 用户
	Role       string `gorm:"column:role"`    // 角色字段
}
