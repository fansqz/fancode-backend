package dao

import (
	"FanCode/models/po"
	"github.com/jinzhu/gorm"
)

// InsertRole 创建角色
func InsertRole(db *gorm.DB, role *po.SysRole) error {
	return db.Create(role).Error
}

// UpdateRole 更新角色
func UpdateRole(db *gorm.DB, role *po.SysRole) error {
	return db.Save(role).Error
}

// DeleteRoleByID 删除角色
func DeleteRoleByID(db *gorm.DB, id uint) error {
	return db.Delete(&po.SysRole{}, id).Error
}

// GetRoleByID 通过角色id获取角色
func GetRoleByID(db *gorm.DB, roleID uint) (*po.SysRole, error) {
	var role po.SysRole
	err := db.First(&role, roleID).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// GetRoleList 获取角色列表
func GetRoleList(db *gorm.DB, page int, pageSize int) ([]*po.SysRole, error) {
	offset := (page - 1) * pageSize
	var roles []*po.SysRole
	err := db.Limit(pageSize).Offset(offset).Find(&roles).Error
	return roles, err
}

// GetRoleCount 获取所有角色数量
func GetRoleCount(db *gorm.DB) (uint, error) {
	var count uint
	err := db.Model(&po.SysRole{}).Count(&count).Error
	return count, err
}
