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

// UpdateUser 更新用户，注：不更新密码
func UpdateUser(db *gorm.DB, user *po.SysUser) error {
	return db.Model(user).Omit("password").Updates(user).Error
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
func GetUserList(db *gorm.DB, username string, page int, pageSize int) ([]*po.SysUser, error) {
	offset := (page - 1) * pageSize
	var users []*po.SysUser
	err := db.Where("username LIKE ?", "%"+username+"%").Limit(pageSize).Offset(offset).Find(&users).Error
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

// DeleteUserRoleByUserID 清除所有与userID关联的userID-roleID数据
func DeleteUserRoleByUserID(db *gorm.DB, userID uint) error {
	user := po.SysUser{}
	user.ID = userID
	if err := db.Model(&user).Association("Roles").Clear().Error; err != nil {
		return err
	}
	return nil
}

// GetRolesByUserID 获取用户关联的所有role的名称
func GetRolesByUserID(db *gorm.DB, userID uint) ([]po.SysRole, error) {
	user := po.SysUser{}
	user.ID = userID
	if err := db.Model(&user).Association("Roles").Find(&user.Roles).Error; err != nil {
		return nil, err
	}
	return user.Roles, nil
}

// GetRoleIDsByUserID 获取用户关联的所有role的id
func GetRoleIDsByUserID(db *gorm.DB, userID uint) ([]uint, error) {
	user := po.SysUser{}
	user.ID = userID
	if err := db.Model(&user).Association("Roles").Find(&user.Roles).Error; err != nil {
		return nil, err
	}
	roleIDs := make([]uint, len(user.Roles))
	for i, role := range user.Roles {
		roleIDs[i] = role.ID
	}
	return roleIDs, nil
}

// InsertRolesToUser 给用户设置角色
func InsertRolesToUser(db *gorm.DB, userID uint, roleIDs []uint) error {
	user := &po.SysUser{}
	user.ID = userID
	for _, apiID := range roleIDs {
		api := &po.SysApi{}
		api.ID = apiID
		err := db.Model(user).Association("Roles").Append(api).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// UpdateUserPassword 更新用户密码
func UpdateUserPassword(db *gorm.DB, userID uint, password string) error {
	user := po.SysUser{}
	user.ID = userID
	return db.Model(&user).Update("password", password).Error
}
