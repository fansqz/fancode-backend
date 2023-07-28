package dao

import (
	"FanCode/models/po"
	"errors"
	"github.com/jinzhu/gorm"
)

// InsertUser 创建用户
func InsertUser(db *gorm.DB, user *po.SysUser) error {
	return db.Create(user).Error
}

// UpdateUser 更新用户
func UpdateUser(db *gorm.DB, user *po.SysUser) error {
	return db.Save(user).Error
}

// DeleteUserByID 删除用户
func DeleteUserByID(db *gorm.DB, id uint) error {
	return db.Delete(&po.SysUser{}, id).Error
}

// GetUserByID 通过用户id获取用户
func GetUserByID(db *gorm.DB, id uint) (*po.SysUser, error) {
	var user po.SysUser
	err := db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserList 获取用户列表
func GetUserList(db *gorm.DB, page int, pageSize int) ([]*po.SysUser, error) {
	offset := (page - 1) * pageSize
	var users []*po.SysUser
	err := db.Limit(pageSize).Offset(offset).Find(&users).Error
	return users, err
}

// GetUserCount 获取所有用户数量
func GetUserCount(db *gorm.DB) (uint, error) {
	var count uint
	err := db.Model(&po.SysUser{}).Count(&count).Error
	return count, err
}

// GetUserByLoginName 根据用户登录名称获取用户
func GetUserByLoginName(db *gorm.DB, loginName string) (*po.SysUser, error) {
	var user po.SysUser
	err := db.Where("login_name = ?", loginName).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CheckLoginName 检测loginname是否存在
func CheckLoginName(db *gorm.DB, loginname string) (bool, error) {
	var user *po.SysUser
	err := db.Where("login_name = ?", loginname).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
