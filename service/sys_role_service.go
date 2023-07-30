package service

import (
	"FanCode/dao"
	e "FanCode/error"
	"FanCode/global"
	"FanCode/models/dto"
	"FanCode/models/po"
)

type SysRoleService interface {
	// InsertSysRole 添加角色
	InsertSysRole(sysSysRole *po.SysRole) (uint, *e.Error)
	// UpdateSysRole 更新角色
	UpdateSysRole(SysRole *po.SysRole) *e.Error
	// DeleteSysRole 删除角色
	DeleteSysRole(id uint) *e.Error
	// GetSysRoleList 获取角色列表
	GetSysRoleList(roleName string, page int, pageSize int) (*dto.PageInfo, *e.Error)
	// UpdateRoleApis 更新角色apis
	UpdateRoleApis(roleID uint, apiIDs []uint) *e.Error
	// UpdateRoleMenus 更新角色menu
	UpdateRoleMenus(roleID uint, menuIDs []uint) *e.Error
	// GetApiIDsByRoleID 通过角色id获取该角色拥有的apiID
	GetApiIDsByRoleID(roleID uint) ([]uint, *e.Error)
	// GetMenuIDsByRoleID 通过角色id获取该角色拥有的menuID
	GetMenuIDsByRoleID(roleID uint) ([]uint, *e.Error)
}

type sysRoleService struct {
}

func NewSysRoleService() SysRoleService {
	return &sysRoleService{}
}

func (r *sysRoleService) InsertSysRole(sysSysRole *po.SysRole) (uint, *e.Error) {
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

func (r *sysRoleService) UpdateSysRole(sysSysRole *po.SysRole) *e.Error {
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

func (r *sysRoleService) DeleteSysRole(id uint) *e.Error {
	// 删除删除角色
	err := dao.DeleteRoleByID(global.Mysql, id)
	if err != nil {
		return e.ErrRoleUnknownError
	}
	return nil
}

func (r *sysRoleService) GetSysRoleList(roleName string, page int, pageSize int) (*dto.PageInfo, *e.Error) {
	// 获取角色列表
	sysSysRoles, err := dao.GetRoleList(global.Mysql, roleName, page, pageSize)
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

func (r *sysRoleService) UpdateRoleApis(roleID uint, apiIDs []uint) *e.Error {
	tx := global.Mysql.Begin()
	err := dao.DeleteRoleAPIsByRoleID(tx, roleID)
	if err != nil {
		tx.Rollback()
		return e.ErrApiUnknownError
	}
	err = dao.InsertApisToRole(tx, roleID, apiIDs)
	if err != nil {
		tx.Rollback() // 发生错误时回滚事务
		return e.ErrApiUnknownError
	}
	tx.Commit()
	return nil
}

func (r *sysRoleService) UpdateRoleMenus(roleID uint, menuIDs []uint) *e.Error {
	tx := global.Mysql.Begin()
	err := dao.DeleteRoleMenusByRoleID(tx, roleID)
	if err != nil {
		tx.Rollback()
		return e.ErrApiUnknownError
	}
	err = dao.InsertMenusToRole(tx, roleID, menuIDs)
	if err != nil {
		tx.Rollback()
		return e.ErrApiUnknownError
	}
	tx.Commit()
	return nil
}

func (r *sysRoleService) GetApiIDsByRoleID(roleID uint) ([]uint, *e.Error) {
	apiIDs, err := dao.GetApiIDsByRoleID(global.Mysql, roleID)
	if err != nil {
		return nil, e.ErrRoleUnknownError
	}
	return apiIDs, nil
}

func (r *sysRoleService) GetMenuIDsByRoleID(roleID uint) ([]uint, *e.Error) {
	menuIDs, err := dao.GetMenuIDsByRoleID(global.Mysql, roleID)
	if err != nil {
		return nil, e.ErrRoleUnknownError
	}
	return menuIDs, nil
}
