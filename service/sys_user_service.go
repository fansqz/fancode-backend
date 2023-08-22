package service

import (
	"FanCode/dao"
	e "FanCode/error"
	"FanCode/global"
	"FanCode/models/dto"
	"FanCode/models/po"
	"FanCode/utils"
	"gorm.io/gorm"
	"log"
	"time"
)

type SysUserService interface {
	// InsertSysUser 添加用户
	InsertSysUser(sysSysUser *po.SysUser) (uint, *e.Error)
	// UpdateSysUser 更新用户，但是不更新密码
	UpdateSysUser(SysUser *po.SysUser) *e.Error
	// DeleteSysUser 删除用户
	DeleteSysUser(id uint) *e.Error
	// GetSysUserList 获取用户列表
	GetSysUserList(userName string, page int, pageSize int) (*dto.PageInfo, *e.Error)
	// UpdateUserRoles 更新角色roleIDs
	UpdateUserRoles(userID uint, roleIDs []uint) *e.Error
	// GetRoleIDsByUserID 通过用户id获取所有角色id
	GetRoleIDsByUserID(userID uint) ([]uint, *e.Error)
	// GetAllSimpleRole
	GetAllSimpleRole() ([]*dto.SimpleRoleDto, *e.Error)
}

type sysUserService struct {
}

func NewSysUserService() SysUserService {
	return &sysUserService{}
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
		log.Println(err)
		return 0, e.ErrSysUserUnknownError
	}
	sysUser.Password = string(p)
	err = dao.InsertUser(global.Mysql, sysUser)
	if err != nil {
		return 0, e.ErrSysUserUnknownError
	}
	return sysUser.ID, nil
}

func (s *sysUserService) UpdateSysUser(sysUser *po.SysUser) *e.Error {
	sysUser.UpdatedAt = time.Now()
	err := dao.UpdateUser(global.Mysql, sysUser)
	if err != nil {
		log.Println(err)
		return e.ErrSysUserUnknownError
	}
	return nil
}

func (s *sysUserService) DeleteSysUser(id uint) *e.Error {
	err := dao.DeleteUserByID(global.Mysql, id)
	if err != nil {
		log.Println(err)
		return e.ErrSysUserUnknownError
	}
	return nil
}

func (s *sysUserService) GetSysUserList(userName string, page int, pageSize int) (*dto.PageInfo, *e.Error) {
	var pageInfo *dto.PageInfo
	err := global.Mysql.Transaction(func(tx *gorm.DB) error {
		userList, err := dao.GetUserList(global.Mysql, userName, page, pageSize)
		if err != nil {
			return err
		}
		userDtoList := make([]*dto.SysUserDtoForList, len(userList))
		for i, user := range userList {
			user.Roles, err = dao.GetRolesByUserID(tx, user.ID)
			if err != nil {
				return err
			}
			userDtoList[i] = dto.NewSysUserDtoForList(user)
		}
		var count int64
		count, err = dao.GetUserCount(global.Mysql)
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
		return nil, e.ErrSysUserUnknownError
	}
	return pageInfo, nil
}

func (s *sysUserService) UpdateUserRoles(userID uint, roleIDs []uint) *e.Error {
	tx := global.Mysql.Begin()
	err := dao.DeleteUserRoleByUserID(tx, userID)
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return e.ErrSysUserUnknownError
	}
	err = dao.InsertRolesToUser(tx, userID, roleIDs)
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return e.ErrSysUserUnknownError
	}
	tx.Commit()
	return nil
}

func (s *sysUserService) GetRoleIDsByUserID(userID uint) ([]uint, *e.Error) {
	roleIDs, err := dao.GetRoleIDsByUserID(global.Mysql, userID)
	if err != nil {
		return nil, e.ErrSysUserUnknownError
	}
	return roleIDs, nil
}

func (s *sysUserService) GetAllSimpleRole() ([]*dto.SimpleRoleDto, *e.Error) {
	roles, err := dao.GetAllSimpleRoleList(global.Mysql)
	if err != nil {
		log.Println(err.Error())
		return nil, e.ErrSysUserUnknownError
	}
	simpleRoles := make([]*dto.SimpleRoleDto, len(roles))
	for i, role := range roles {
		simpleRoles[i] = dto.NewSimpleRoleDto(role)
	}
	return simpleRoles, nil
}
