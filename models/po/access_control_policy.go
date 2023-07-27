package po

import "github.com/jinzhu/gorm"

type AccessControlPolicy struct {
	gorm.Model         // 如果使用 GORM，可以嵌入 gorm.Model
	Role        string `gorm:"column:role"`        // 角色名称
	URL         string `gorm:"column:url"`         // URL路径
	Method      string `gorm:"column:method"`      // 请求方法
	Description string `gorm:"column:description"` //描述
}
