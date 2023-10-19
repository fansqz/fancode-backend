package service

import (
	"FanCode/dao"
	e "FanCode/error"
	"FanCode/global"
	"FanCode/models/dto"
	"FanCode/models/po"
	"gorm.io/gorm"
	"time"
)

type SysRoleService interface {

	// GetRoleByID 根角色户id获取角色信息
	GetRoleByID(roleID uint) (*po.SysRole, *e.Error)
	// InsertSysRole 添加角色
	InsertSysRole(sysSysRole *po.SysRole) (uint, *e.Error)
	// UpdateSysRole 更新角色
	UpdateSysRole(SysRole *po.SysRole) *e.Error
	// DeleteSysRole 删除角色
	DeleteSysRole(id uint) *e.Error
	// GetSysRoleList 获取角色列表
	GetSysRoleList(pageQuery *dto.PageQuery) (*dto.PageInfo, *e.Error)
	// UpdateRoleApis 更新角色apis
	UpdateRoleApis(roleID uint, apiIDs []uint) *e.Error
	// UpdateRoleMenus 更新角色menu
	UpdateRoleMenus(roleID uint, menuIDs []uint) *e.Error
	// GetApiIDsByRoleID 通过角色id获取该角色拥有的apiID
	GetApiIDsByRoleID(roleID uint) ([]uint, *e.Error)
	// GetMenuIDsByRoleID 通过角色id获取该角色拥有的menuID
	GetMenuIDsByRoleID(roleID uint) ([]uint, *e.Error)
	// GetApisByRoleID 通过角色id获取该角色的所有api
	GetApisByRoleID(roleID uint) ([]*po.SysApi, *e.Error)
}

type sysRoleService struct {
	sysRoleDao dao.SysRoleDao
}

func NewSysRoleService(roleDao dao.SysRoleDao) SysRoleService {
	return &sysRoleService{
		sysRoleDao: roleDao,
	}
}

func (r *sysRoleService) GetRoleByID(roleID uint) (*po.SysRole, *e.Error) {
	role, err := r.sysRoleDao.GetRoleByID(global.Mysql, roleID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, e.ErrMysql
	}
	if err == gorm.ErrRecordNotFound {
		return nil, e.ErrRoleNotExist
	}
	return role, nil
}

func (r *sysRoleService) InsertSysRole(sysRole *po.SysRole) (uint, *e.Error) {
	// 添加
	err := r.sysRoleDao.InsertRole(global.Mysql, sysRole)
	if err != nil {
		return 0, e.ErrMysql
	}
	return sysRole.ID, nil
}

func (r *sysRoleService) UpdateSysRole(sysRole *po.SysRole) *e.Error {
	sysRole.UpdatedAt = time.Now()
	err := r.sysRoleDao.UpdateRole(global.Mysql, sysRole)
	if err != nil {
		return e.ErrMysql
	}
	return nil
}

func (r *sysRoleService) DeleteSysRole(id uint) *e.Error {
	// 删除删除角色
	err := r.sysRoleDao.DeleteRoleByID(global.Mysql, id)
	if err != nil {
		return e.ErrMysql
	}
	return nil
}

func (r *sysRoleService) GetSysRoleList(query *dto.PageQuery) (*dto.PageInfo, *e.Error) {
	var roleQuery *po.SysRole
	if query.Query != nil {
		roleQuery = query.Query.(*po.SysRole)
	}
	// 获取角色列表
	sysSysRoles, err := r.sysRoleDao.GetRoleList(global.Mysql, query)
	if err != nil {
		return nil, e.ErrMysql
	}
	newSysRoles := make([]*dto.SysRoleDtoForList, len(sysSysRoles))
	for i := 0; i < len(sysSysRoles); i++ {
		newSysRoles[i] = dto.NewSysRoleDtoForList(sysSysRoles[i])
	}
	// 获取所有角色总数目
	var count int64
	count, err = r.sysRoleDao.GetRoleCount(global.Mysql, roleQuery)
	if err != nil {
		return nil, e.ErrMysql
	}
	pageInfo := &dto.PageInfo{
		Total: count,
		Size:  int64(len(newSysRoles)),
		List:  newSysRoles,
	}
	return pageInfo, nil
}

func (r *sysRoleService) UpdateRoleApis(roleID uint, apiIDs []uint) *e.Error {
	tx := global.Mysql.Begin()
	err := r.sysRoleDao.DeleteRoleAPIsByRoleID(tx, roleID)
	if err != nil {
		tx.Rollback()
		return e.ErrMysql
	}
	err = r.sysRoleDao.InsertApisToRole(tx, roleID, apiIDs)
	if err != nil {
		tx.Rollback() // 发生错误时回滚事务
		return e.ErrMysql
	}
	tx.Commit()
	return nil
}

func (r *sysRoleService) UpdateRoleMenus(roleID uint, menuIDs []uint) *e.Error {
	tx := global.Mysql.Begin()
	err := r.sysRoleDao.DeleteRoleMenusByRoleID(tx, roleID)
	if err != nil {
		tx.Rollback()
		return e.ErrMysql
	}
	err = r.sysRoleDao.InsertMenusToRole(tx, roleID, menuIDs)
	if err != nil {
		tx.Rollback()
		return e.ErrMysql
	}
	tx.Commit()
	return nil
}

func (r *sysRoleService) GetApiIDsByRoleID(roleID uint) ([]uint, *e.Error) {
	apiIDs, err := r.sysRoleDao.GetApiIDsByRoleID(global.Mysql, roleID)
	if err != nil {
		return nil, e.ErrMysql
	}
	return apiIDs, nil
}

func (r *sysRoleService) GetMenuIDsByRoleID(roleID uint) ([]uint, *e.Error) {
	menuIDs, err := r.sysRoleDao.GetMenuIDsByRoleID(global.Mysql, roleID)
	if err != nil {
		return nil, e.ErrMysql
	}
	return menuIDs, nil
}

func (r *sysRoleService) GetApisByRoleID(roleID uint) ([]*po.SysApi, *e.Error) {
	apis, err := r.sysRoleDao.GetApisByRoleID(global.Mysql, roleID)
	if err != nil {
		return nil, e.ErrMysql
	}
	return apis, nil
}
