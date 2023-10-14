package service

import (
	"FanCode/dao/mock"
	e "FanCode/error"
	"FanCode/models/po"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
	"time"
)

func TestSysRoleService_GetRoleByID(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	// mock apiDao
	roleDao := mock.NewMockSysRoleDao(mockCtl)
	role := &po.SysRole{
		Name:        "role名",
		Description: "role描述",
	}
	role.ID = 1
	roleDao.EXPECT().GetRoleByID(gomock.Any(), uint(1)).Return(role, nil)
	roleDao.EXPECT().GetRoleByID(gomock.Any(), uint(2)).Return(nil, gorm.ErrRecordNotFound)
	roleDao.EXPECT().GetRoleByID(gomock.Any(), uint(3)).Return(nil, gorm.ErrInvalidDB)

	// 测试
	roleService := NewSysRoleService(roleDao)
	role2, err := roleService.GetRoleByID(1)
	assert.Equal(t, role, role2)
	assert.Nil(t, err)

	role3, err := roleService.GetRoleByID(2)
	assert.Nil(t, role3)
	assert.Equal(t, err, e.ErrRoleNotExist)

	role4, err := roleService.GetRoleByID(3)
	assert.Nil(t, role4)
	assert.Equal(t, err, e.ErrMysql)
}

func TestSysRoleService_InsertSysRole(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	// mock apiDao
	roleDao := mock.NewMockSysRoleDao(mockCtl)
	role := &po.SysRole{
		Name:        "menu名",
		Description: "menu描述",
	}
	roleDao.EXPECT().InsertRole(gomock.Any(), role).
		DoAndReturn(func(db *gorm.DB, sysRole *po.SysRole) error {
			assert.Equal(t, role, sysRole)
			sysRole.ID = 1
			return nil
		})
	roleDao.EXPECT().InsertRole(gomock.Any(), gomock.Any()).Return(gorm.ErrInvalidDB)

	// 测试
	roleService := NewSysRoleService(roleDao)
	id, err := roleService.InsertSysRole(role)
	assert.Equal(t, id, uint(1))
	assert.Nil(t, err)
	id, err = roleService.InsertSysRole(&po.SysRole{})
	assert.Equal(t, id, uint(0))
	assert.Equal(t, err, e.ErrMysql)
}

func TestSysRoleService_UpdateSysRole(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	// mock apiDao
	roleDao := mock.NewMockSysRoleDao(mockCtl)
	role := &po.SysRole{
		Name:        "role名",
		Description: "role描述",
	}
	role.ID = 1
	roleDao.EXPECT().UpdateRole(gomock.Any(), role).
		DoAndReturn(func(db *gorm.DB, sysRole *po.SysRole) error {
			assert.NotEqual(t, sysRole.UpdatedAt, time.Time{})
			sysRole.UpdatedAt = time.Time{}
			assert.Equal(t, role, sysRole)
			return nil
		})
	roleDao.EXPECT().UpdateRole(gomock.Any(), gomock.Any()).Return(gorm.ErrInvalidDB)

	// 测试
	roleService := NewSysRoleService(roleDao)
	err := roleService.UpdateSysRole(role)
	assert.Nil(t, err)
	err = roleService.UpdateSysRole(&po.SysRole{})
	assert.Equal(t, err, e.ErrMysql)
}

func TestSysRoleService_DeleteSysRole(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	// mock apiDao
	roleDao := mock.NewMockSysRoleDao(mockCtl)
	roleDao.EXPECT().DeleteRoleByID(gomock.Any(), uint(1)).Return(nil)
	roleDao.EXPECT().DeleteRoleByID(gomock.Any(), uint(2)).Return(gorm.ErrInvalidDB)

	// 测试
	roleService := NewSysRoleService(roleDao)
	err := roleService.DeleteSysRole(1)
	assert.Nil(t, err)
	err = roleService.DeleteSysRole(2)
	assert.Equal(t, err, e.ErrMysql)
}
