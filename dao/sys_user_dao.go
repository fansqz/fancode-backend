package dao

import (
	"FanCode/models/dto"
	"FanCode/models/po"
	"errors"
	"gorm.io/gorm"
)

// InsertUser 创建用户
func InsertUser(db *gorm.DB, user *po.SysUser) error {
	return db.Create(user).Error
}

// UpdateUser
func UpdateUser(db *gorm.DB, id uint, m map[string]interface{}) error {
	return db.Model(&po.SysUser{}).Where("id = ?", id).Updates(m).Error
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

// GetUserNameByID 通过用户id获取用户名称
func GetUserNameByID(db *gorm.DB, id uint) (string, error) {
	var user po.SysUser
	err := db.Select("username").First(&user, id).Error
	if err != nil {
		return "", err
	}
	return user.Username, nil
}

// GetUserList 获取用户列表
func GetUserList(db *gorm.DB, pageQuery *dto.PageQuery) ([]*po.SysUser, error) {
	user := pageQuery.Query.(*po.SysUser)
	offset := (pageQuery.Page - 1) * pageQuery.PageSize
	var users []*po.SysUser
	err := db.Where("username LIKE ?", "%"+user.Username+"%").
		Limit(pageQuery.PageSize).Offset(offset).Find(&users).Error
	return users, err
}

// GetUserCount 获取所有用户数量
func GetUserCount(db *gorm.DB) (int64, error) {
	var count int64
	err := db.Model(&po.SysUser{}).Count(&count).Error
	return count, err
}

// GetUserByLoginName 根据用户登录名称获取用户
func GetUserByLoginName(db *gorm.DB, loginName string) (*po.SysUser, error) {
	var user po.SysUser
	err := db.Where("login_name = ?", loginName).Preload("Roles").First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail 根据用户邮箱获取用户信息
func GetUserByEmail(db *gorm.DB, email string) (*po.SysUser, error) {
	var user po.SysUser
	err := db.Where("email = ?", email).Preload("Roles").First(&user).Error
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

func ListLoginName(db *gorm.DB, loginName string) ([]string, error) {
	var users []*po.SysUser
	err := db.Model(&po.SysUser{}).Where("login_name like ?", loginName+"%").
		Select("login_name").Find(&users).Error
	if err != nil {
		return nil, err
	}
	answer := make([]string, len(users))
	for i := 0; i < len(users); i++ {
		answer[i] = users[i].LoginName
	}
	return answer, nil
}

// CheckEmail 检测邮箱是否已经存在
func CheckEmail(db *gorm.DB, email string) (bool, error) {
	var user *po.SysUser
	err := db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return user.ID != 0, nil
}

// DeleteUserRoleByUserID 清除所有与userID关联的userID-roleID数据
func DeleteUserRoleByUserID(db *gorm.DB, userID uint) error {
	user := po.SysUser{}
	user.ID = userID
	if err := db.Model(&user).Association("Roles").Clear(); err != nil {
		return err
	}
	return nil
}

// GetRolesByUserID 获取用户关联的所有role的名称
func GetRolesByUserID(db *gorm.DB, userID uint) ([]po.SysRole, error) {
	user := po.SysUser{}
	user.ID = userID
	if err := db.Model(&user).Association("Roles").Find(&user.Roles); err != nil {
		return nil, err
	}
	return user.Roles, nil
}

// GetRoleIDsByUserID 获取用户关联的所有role的id
func GetRoleIDsByUserID(db *gorm.DB, userID uint) ([]uint, error) {
	user := po.SysUser{}
	user.ID = userID
	if err := db.Model(&user).Association("Roles").Find(&user.Roles); err != nil {
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
	for _, roleID := range roleIDs {
		role := &po.SysRole{}
		role.ID = roleID
		err := db.Model(user).Association("Roles").Append(role)
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
