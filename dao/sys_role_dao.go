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
func GetRoleList(db *gorm.DB, roleName string, page int, pageSize int) ([]*po.SysRole, error) {
	offset := (page - 1) * pageSize
	var roles []*po.SysRole
	err := db.Where("name LIKE ?", "%"+roleName+"%").Limit(pageSize).Offset(offset).Find(&roles).Error
	return roles, err
}

// GetRoleCount 获取所有角色数量
func GetRoleCount(db *gorm.DB) (uint, error) {
	var count uint
	err := db.Model(&po.SysRole{}).Count(&count).Error
	return count, err
}

// InsertMenusToRole 给角色添加menu
func InsertMenusToRole(db *gorm.DB, roleID uint, menus []uint) error {
	role := &po.SysRole{}
	role.ID = roleID
	for _, menuID := range menus {
		menu := &po.SysMenu{}
		menu.ID = menuID
		err := db.Model(role).Association("Menus").Append(menu).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// GetRoleMenuIDs 获取用户关联的所有menu的id
func GetRoleMenuIDs(db *gorm.DB, roleID uint) ([]uint, error) {
	var role *po.SysRole
	if err := db.Model(&po.SysRole{}).Select("id").Preload("Menus", "id").
		First(&role, roleID).Error; err != nil {
		return nil, err
	}
	menuIDs := make([]uint, len(role.Menus))
	for i, api := range role.Menus {
		menuIDs[i] = api.ID
	}
	return menuIDs, nil
}

// InsertApisToRole 给角色添加api
func InsertApisToRole(db *gorm.DB, roleID uint, apis []uint) error {
	role := &po.SysRole{}
	role.ID = roleID
	for _, apiID := range apis {
		api := &po.SysApi{}
		api.ID = apiID
		err := db.Model(role).Association("Apis").Append(api).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// GetRoleApiIDs 获取用户关联的所有api的id
func GetRoleApiIDs(db *gorm.DB, roleID uint) ([]uint, error) {
	var role *po.SysRole
	if err := db.Model(&po.SysRole{}).Select("id").Preload("Apis", "id").
		First(&role, roleID).Error; err != nil {
		return nil, err
	}
	apiIDs := make([]uint, len(role.Apis))
	for i, api := range role.Apis {
		apiIDs[i] = api.ID
	}
	return apiIDs, nil
}
