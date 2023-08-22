package dao

import (
	"FanCode/models/po"
	"gorm.io/gorm"
)

// InsertRole 创建角色
func InsertRole(db *gorm.DB, role *po.SysRole) error {
	return db.Create(role).Error
}

// UpdateRole 更新角色
func UpdateRole(db *gorm.DB, role *po.SysRole) error {
	return db.Model(role).UpdateColumns(map[string]interface{}{
		"name":        role.Name,
		"description": role.Description,
		"updated_at":  role.UpdatedAt,
	}).Error
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
func GetRoleCount(db *gorm.DB) (int64, error) {
	var count int64
	err := db.Model(&po.SysRole{}).Count(&count).Error
	return count, err
}

// InsertMenusToRole 给角色添加menu
func InsertMenusToRole(db *gorm.DB, roleID uint, menuIDs []uint) error {
	role := &po.SysRole{}
	role.ID = roleID
	var menus []po.SysMenu
	for _, menuID := range menuIDs {
		menu := po.SysMenu{}
		menu.ID = menuID
		menus = append(menus, menu)
	}
	err := db.Model(role).Association("Menus").Append(menus)
	return err
}

// DeleteRoleMenusByRoleID 清除所有与roleID关联的roleID-menuID数据
func DeleteRoleMenusByRoleID(db *gorm.DB, roleID uint) error {
	role := po.SysRole{}
	role.ID = roleID
	if err := db.Model(&role).Association("Menus").Clear(); err != nil {
		return err
	}

	return nil
}

// GetMenuIDsByRoleID 获取用户关联的所有menu的id
func GetMenuIDsByRoleID(db *gorm.DB, roleID uint) ([]uint, error) {
	var role po.SysRole
	role.ID = roleID
	if err := db.Model(&role).Association("Menus").Find(&role.Menus); err != nil {
		return nil, err
	}
	menuIDs := make([]uint, len(role.Menus))
	for i, menu := range role.Menus {
		menuIDs[i] = menu.ID
	}
	return menuIDs, nil
}

// GetMenusByRoleID 通过用户角色获取菜单列表
func GetMenusByRoleID(db *gorm.DB, roleID uint) ([]po.SysMenu, error) {
	var role po.SysRole
	role.ID = roleID
	if err := db.Model(&role).Association("Menus").Find(&role.Menus); err != nil {
		return nil, err
	}
	return role.Menus, nil
}

// InsertApisToRole 给角色添加api
func InsertApisToRole(db *gorm.DB, roleID uint, apiIDs []uint) error {
	role := &po.SysRole{}
	role.ID = roleID
	var apis []po.SysApi
	for _, apiID := range apiIDs {
		api := po.SysApi{}
		api.ID = apiID
		apis = append(apis, api)
	}
	err := db.Model(role).Association("Apis").Append(apis)
	return err
}

// DeleteRoleAPIsByRoleID 清除所有与roleID关联的roleID-apiID数据
func DeleteRoleAPIsByRoleID(db *gorm.DB, roleID uint) error {
	role := po.SysRole{}
	role.ID = roleID
	if err := db.Model(&role).Association("Apis").Clear(); err != nil {
		return err
	}

	return nil
}

// GetApiIDsByRoleID 获取用户关联的所有api的id
func GetApiIDsByRoleID(db *gorm.DB, roleID uint) ([]uint, error) {
	var role po.SysRole
	role.ID = roleID
	if err := db.Model(&role).Association("Apis").Find(&role.Apis); err != nil {
		return nil, err
	}
	apiIDs := make([]uint, len(role.Apis))
	for i, api := range role.Apis {
		apiIDs[i] = api.ID
	}
	return apiIDs, nil
}

// GetAllSimpleRoleList 获取所有角色列表，只含有id和name
func GetAllSimpleRoleList(db *gorm.DB) ([]*po.SysRole, error) {
	var roles []*po.SysRole
	err := db.Select([]string{"id", "name"}).Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}
