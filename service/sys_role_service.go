package service

import (
	"FanCode/dao"
	e "FanCode/error"
	"FanCode/global"
	"FanCode/models/dto"
	"FanCode/models/po"
)

type SysSysRoleService interface {
	// InsertSysRole 添加角色
	InsertSysRole(sysSysRole *po.SysRole) (uint, *e.Error)
	// UpdateSysRole 更新角色
	UpdateSysRole(SysRole *po.SysRole) *e.Error
	// DeleteSysRole 删除角色
	DeleteSysRole(id uint) *e.Error
	// GetSysRoleList 获取角色列表
	GetSysRoleList(page int, pageSize int) (*dto.PageInfo, *e.Error)
}

type sysSysRoleService struct {
}

func NewSysSysRoleService() SysSysRoleService {
	return &sysSysRoleService{}
}

func (r *sysSysRoleService) InsertSysRole(sysSysRole *po.SysRole) (uint, *e.Error) {
	// 对设置值的数据设置默认值
	if sysSysRole.Name == "" {
		sysSysRole.Name = "未命名角色"
	}
	// 添加
	err := dao.InsertRole(global.Mysql, sysSysRole)
	if err != nil {
		return 0, e.ErrRoleUnknownError
	}
	return sysSysRole.ID, nil
}

func (q *sysSysRoleService) UpdateSysRole(sysSysRole *po.SysRole) *e.Error {
	// 对设置值的数据设置默认值
	if sysSysRole.Name == "" {
		sysSysRole.Name = "未命名角色"
	}

	err := dao.UpdateRole(global.Mysql, sysSysRole)
	if err != nil {
		return e.ErrRoleUnknownError
	}
	return nil
}

func (q *sysSysRoleService) DeleteSysRole(id uint) *e.Error {
	// 删除删除角色
	err := dao.DeleteRoleByID(global.Mysql, id)
	if err != nil {
		return e.ErrRoleUnknownError
	}
	return nil
}

func (q *sysSysRoleService) GetSysRoleList(page int, pageSize int) (*dto.PageInfo, *e.Error) {
	// 获取角色列表
	sysSysRoles, err := dao.GetRoleList(global.Mysql, page, pageSize)
	if err != nil {
		return nil, e.ErrRoleUnknownError
	}
	newSysRoles := make([]*dto.SysRoleDtoForList, len(sysSysRoles))
	for i := 0; i < len(sysSysRoles); i++ {
		newSysRoles[i] = dto.NewSysRoleDtoForList(sysSysRoles[i])
	}
	// 获取所有角色总数目
	var count uint
	count, err = dao.GetRoleCount(global.Mysql)
	if err != nil {
		return nil, e.ErrRoleUnknownError
	}
	pageInfo := &dto.PageInfo{
		Total: count,
		Size:  uint(len(newSysRoles)),
		List:  newSysRoles,
	}
	return pageInfo, nil
}
