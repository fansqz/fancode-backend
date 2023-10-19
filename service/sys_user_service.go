package service

import (
	"FanCode/dao"
	e "FanCode/error"
	"FanCode/global"
	"FanCode/models/dto"
	"FanCode/models/po"
	"FanCode/utils"
	"gorm.io/gorm"
	"time"
)

type SysUserService interface {
	// GetUserByID 根据用户id获取用户信息
	GetUserByID(userID uint) (*po.SysUser, *e.Error)
	// InsertSysUser 添加用户
	InsertSysUser(sysSysUser *po.SysUser) (uint, *e.Error)
	// UpdateSysUser 更新用户，但是不更新密码
	UpdateSysUser(SysUser *po.SysUser) *e.Error
	// DeleteSysUser 删除用户
	DeleteSysUser(id uint) *e.Error
	// GetSysUserList 获取用户列表
	GetSysUserList(pageQuery *dto.PageQuery) (*dto.PageInfo, *e.Error)
	// UpdateUserRoles 更新角色roleIDs
	UpdateUserRoles(userID uint, roleIDs []uint) *e.Error
	// GetRoleIDsByUserID 通过用户id获取所有角色id
	GetRoleIDsByUserID(userID uint) ([]uint, *e.Error)
	// GetAllSimpleRole
	GetAllSimpleRole() ([]*dto.SimpleRoleDto, *e.Error)
}

type sysUserService struct {
	sysUserDao dao.SysUserDao
	sysRoleDao dao.SysRoleDao
}

func NewSysUserService(userDao dao.SysUserDao, roleDao dao.SysRoleDao) SysUserService {
	return &sysUserService{
		sysUserDao: userDao,
		sysRoleDao: roleDao,
	}
}

func (s *sysUserService) GetUserByID(userID uint) (*po.SysUser, *e.Error) {
	user, err := s.sysUserDao.GetUserByID(global.Mysql, userID)
	if err != nil {
		return nil, e.ErrMysql
	}
	return user, nil
}

func (s *sysUserService) InsertSysUser(sysUser *po.SysUser) (uint, *e.Error) {
	if sysUser.Username == "" {
		sysUser.Username = "fancode"
	}
	if sysUser.LoginName == "" {
		sysUser.LoginName = sysUser.LoginName + utils.GetUUID()
	}
	if sysUser.Password == "" {
		sysUser.Password = global.Conf.DefaultPassword
	}
	p, err := utils.GetPwd(sysUser.Password)
	if err != nil {
		return 0, e.ErrMysql
	}
	sysUser.Password = string(p)
	err = s.sysUserDao.InsertUser(global.Mysql, sysUser)
	if err != nil {
		return 0, e.ErrMysql
	}
	return sysUser.ID, nil
}

func (s *sysUserService) UpdateSysUser(sysUser *po.SysUser) *e.Error {
	sysUser.UpdatedAt = time.Now()
	err := s.sysUserDao.UpdateUser(global.Mysql, sysUser)
	if err != nil {
		return e.ErrMysql
	}
	return nil
}

func (s *sysUserService) DeleteSysUser(id uint) *e.Error {
	err := s.sysUserDao.DeleteUserByID(global.Mysql, id)
	if err != nil {
		return e.ErrMysql
	}
	return nil
}

func (s *sysUserService) GetSysUserList(pageQuery *dto.PageQuery) (*dto.PageInfo, *e.Error) {
	var pageInfo *dto.PageInfo
	err := global.Mysql.Transaction(func(tx *gorm.DB) error {
		userList, err := s.sysUserDao.GetUserList(global.Mysql, pageQuery)
		if err != nil {
			return err
		}
		userDtoList := make([]*dto.SysUserDtoForList, len(userList))
		for i, user := range userList {
			user.Roles, err = s.sysUserDao.GetRolesByUserID(tx, user.ID)
			if err != nil {
				return err
			}
			userDtoList[i] = dto.NewSysUserDtoForList(user)
		}
		var count int64
		count, err = s.sysUserDao.GetUserCount(global.Mysql)
		if err != nil {
			return err
		}
		pageInfo = &dto.PageInfo{
			Total: count,
			List:  userDtoList,
		}
		return nil
	})
	if err != nil {
		return nil, e.ErrMysql
	}
	return pageInfo, nil
}

func (s *sysUserService) UpdateUserRoles(userID uint, roleIDs []uint) *e.Error {
	tx := global.Mysql.Begin()
	err := s.sysUserDao.DeleteUserRoleByUserID(tx, userID)
	if err != nil {
		tx.Rollback()
		return e.ErrMysql
	}
	err = s.sysUserDao.InsertRolesToUser(tx, userID, roleIDs)
	if err != nil {
		tx.Rollback()
		return e.ErrMysql
	}
	tx.Commit()
	return nil
}

func (s *sysUserService) GetRoleIDsByUserID(userID uint) ([]uint, *e.Error) {
	roleIDs, err := s.sysUserDao.GetRoleIDsByUserID(global.Mysql, userID)
	if err != nil {
		return nil, e.ErrMysql
	}
	return roleIDs, nil
}

func (s *sysUserService) GetAllSimpleRole() ([]*dto.SimpleRoleDto, *e.Error) {
	roles, err := s.sysRoleDao.GetAllSimpleRoleList(global.Mysql)
	if err != nil {
		return nil, e.ErrMysql
	}
	simpleRoles := make([]*dto.SimpleRoleDto, len(roles))
	for i, role := range roles {
		simpleRoles[i] = dto.NewSimpleRoleDto(role)
	}
	return simpleRoles, nil
}
