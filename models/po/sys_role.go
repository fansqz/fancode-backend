package po

import (
	"github.com/jinzhu/gorm"
)

type SysRole struct {
	gorm.Model
	Name        string   `gorm:"column:name" json:"name"`               // 角色名称
	Alias       string   `gorm:"column:alias" json:"alias"`             // 备注
	Description string   `gorm:"column:description" json:"description"` // 描述
	Apis        []SysApi `gorm:"many2many:role_apis;" json:"apis"`      // 角色关联的api
}
