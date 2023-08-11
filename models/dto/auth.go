package dto

import (
	"FanCode/models/po"
)

type UserInfo struct {
	ID        uint     `json:"id"`
	LoginName string   `json:"loginName"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Phone     string   `json:"phone"`
	Roles     []uint   `json:"roles"`
	Menus     []string `json:"menus"`
}

func NewUserInfo(user *po.SysUser) *UserInfo {
	userInfo := &UserInfo{
		ID:        user.ID,
		LoginName: user.LoginName,
		Username:  user.Username,
		Email:     user.Email,
		Phone:     user.Phone,
	}
	userInfo.Roles = make([]uint, len(user.Roles))
	userInfo.Menus = make([]string, 10)
	index := 0
	for i := 0; i < len(user.Roles); i++ {
		userInfo.Roles[i] = user.Roles[i].ID
		for j := 0; j < len(user.Roles[i].Menus); j++ {
			userInfo.Menus[index] = user.Roles[i].Menus[j].Name
		}
	}
	return userInfo
}
