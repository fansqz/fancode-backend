package dto

import (
	"FanCode/models/po"
	"FanCode/utils"
)

// SysRoleDtoForList 获取角色列表
type SysRoleDtoForList struct {
	ID          uint       `json:"id"`
	UpdatedAt   utils.Time `json:"updatedAt"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
}

func NewSysRoleDtoForList(role *po.SysRole) *SysRoleDtoForList {
	response := &SysRoleDtoForList{
		ID:          role.ID,
		UpdatedAt:   utils.Time(role.UpdatedAt),
		Name:        role.Name,
		Description: role.Description,
	}
	return response
}
