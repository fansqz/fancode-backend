package dao

import (
	"FanCode/models/po"
	"github.com/jinzhu/gorm"
)

// InsertMenu 创建角色
func InsertMenu(db *gorm.DB, menu *po.SysMenu) error {
	return db.Create(menu).Error
}

// GetMenuByID 通过menu的id获取menu
func GetMenuByID(db *gorm.DB, id uint) (*po.SysMenu, error) {
	var menu po.SysMenu
	err := db.First(&menu, id).Error
	return &menu, err
}

// GetMenuListByParentID 通过父id找到其所有子menu
func GetMenuListByParentID(db *gorm.DB, parentID int32) ([]*po.SysMenu, error) {
	var sysMenus []*po.SysMenu
	err := db.Where("parent_menu_id = ?", parentID).Find(&sysMenus).Error
	if err != nil {
		return nil, err
	}
	return sysMenus, nil
}

// GetMenuCount 获取menu总数
func GetMenuCount(db *gorm.DB) (uint, error) {
	var count uint
	err := db.Model(&po.SysMenu{}).Count(&count).Error
	return count, err
}

// GetMenuListByPathKeyword 模糊查询menu
func GetMenuListByPathKeyword(db *gorm.DB, keyword string, page int, pageSize int) ([]*po.SysMenu, error) {
	var sysMenus []*po.SysMenu
	err := db.Where("path LIKE ?", "%"+keyword+"%").Offset((page - 1) * pageSize).Limit(pageSize).Find(&sysMenus).Error
	if err != nil {
		return nil, err
	}
	return sysMenus, nil
}

// DeleteMenuByID 根据menu的id进行删除
func DeleteMenuByID(db *gorm.DB, id uint) error {
	return db.Delete(&po.SysMenu{}, id).Error
}

// UpdateMenu 修改menu
func UpdateMenu(db *gorm.DB, menu *po.SysMenu) error {
	return db.Save(menu).Error
}

// GetChildMenusByParentID 根据父API的ID获取所有子API
func GetChildMenusByParentID(db *gorm.DB, parentID uint) ([]*po.SysMenu, error) {
	var childMenus []*po.SysMenu
	if err := db.Where("parent_menu_id = ?", parentID).Find(&childMenus).Error; err != nil {
		return nil, err
	}
	return childMenus, nil
}

// GetAllMenu 获取所有menu
func GetAllMenu(db *gorm.DB) ([]*po.SysMenu, error) {
	var menuList []*po.SysMenu
	err := db.Find(&menuList).Error
	return menuList, err
}
