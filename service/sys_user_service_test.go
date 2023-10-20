package service

import (
	"FanCode/dao/mock"
	e "FanCode/error"
	"FanCode/global"
	"FanCode/models/dto"
	"FanCode/models/po"
	"FanCode/utils"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
	"time"
)

func TestSysUserService_GetUserByID(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	// mock userDao
	userDao := mock.NewMockSysUserDao(mockCtl)
	user := &po.SysUser{
		Avatar:       "avatar",
		Username:     "username",
		LoginName:    "loginName",
		Password:     "password",
		Email:        "email",
		Phone:        "12345678910",
		Introduction: "introduction",
		Sex:          1,
		BirthDay:     time.Now(),
		Model: gorm.Model{
			ID: 1,
		},
	}
	userDao.EXPECT().GetUserByID(gomock.Any(), uint(1)).Return(user, nil)
	userDao.EXPECT().GetUserByID(gomock.Any(), uint(2)).Return(nil, gorm.ErrRecordNotFound)
	userDao.EXPECT().GetUserByID(gomock.Any(), uint(3)).Return(nil, gorm.ErrInvalidDB)

	// 测试
	userService := NewSysUserService(userDao, nil)
	user2, err := userService.GetUserByID(1)
	assert.Equal(t, user2, user)
	assert.Nil(t, err)

	user2, err = userService.GetUserByID(2)
	assert.Nil(t, user2)
	assert.Equal(t, err, e.ErrUserNotExist)

	user2, err = userService.GetUserByID(3)
	assert.Nil(t, user2)
	assert.Equal(t, err, e.ErrMysql)
}

func TestSysUserService_UpdateSysUser(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	// mock userDao
	userDao := mock.NewMockSysUserDao(mockCtl)
	user := &po.SysUser{
		Avatar:       "avatar",
		Username:     "username",
		LoginName:    "loginName",
		Password:     "password",
		Email:        "email",
		Phone:        "12345678910",
		Introduction: "introduction",
		Sex:          1,
		BirthDay:     time.Now(),
		Model: gorm.Model{
			ID: 1,
		},
	}
	userDao.EXPECT().UpdateUser(gomock.Any(), user).
		DoAndReturn(func(db *gorm.DB, sysUser *po.SysUser) error {
			assert.NotEqual(t, sysUser.UpdatedAt, time.Time{})
			sysUser.UpdatedAt = time.Time{}
			assert.Equal(t, user, sysUser)
			return nil
		})
	userDao.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(gorm.ErrInvalidDB)

	// 测试
	userService := NewSysUserService(userDao, nil)
	err := userService.UpdateSysUser(user)
	assert.Nil(t, err)
	err = userService.UpdateSysUser(&po.SysUser{})
	assert.Equal(t, err, e.ErrMysql)
}

func TestSysUserService_DeleteSysUser(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	// mock userDao
	userDao := mock.NewMockSysUserDao(mockCtl)
	userDao.EXPECT().DeleteUserByID(gomock.Any(), uint(1)).Return(nil)
	userDao.EXPECT().DeleteUserByID(gomock.Any(), uint(2)).Return(gorm.ErrInvalidDB)

	// 测试
	userService := NewSysUserService(userDao, nil)
	err := userService.DeleteSysUser(1)
	assert.Nil(t, err)
	err = userService.DeleteSysUser(2)
	assert.Equal(t, err, e.ErrMysql)
}

func TestSysUserService_GetSysUserList(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	userDao := mock.NewMockSysUserDao(mockCtl)

	db, mock, err := sqlmock.New()
	defer db.Close()
	assert.Nil(t, err)
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		SkipInitializeWithVersion: true,
		Conn:                      db,
	}), &gorm.Config{})
	assert.Nil(t, err)
	global.Mysql = gormDB

	// 用例1
	testUserList := []*po.SysUser{
		{
			Avatar: "头像1", Username: "user1", LoginName: "loginName1", Introduction: "介绍1",
			Email: "123456@qq.com", Phone: "12345678910", Sex: 1, BirthDay: time.Now(),
			Model: gorm.Model{ID: 1, UpdatedAt: time.Now()},
		},
		{
			Avatar: "头像2", Username: "user2", LoginName: "loginName2", Introduction: "介绍2",
			Email: "123457@qq.com", Phone: "12345678911", Sex: 2, BirthDay: time.Now(),
			Model: gorm.Model{ID: 2, UpdatedAt: time.Now()},
		},
		{
			Avatar: "头像3", Username: "user3", LoginName: "loginName3", Introduction: "介绍3",
			Email: "123458@qq.com", Phone: "12345678912", Sex: 0, BirthDay: time.Now(),
			Model: gorm.Model{ID: 3, UpdatedAt: time.Now()},
		},
	}
	userDao.EXPECT().GetUserList(gomock.Any(), &dto.PageQuery{
		Page: 1, PageSize: 3, SortProperty: "name", SortRule: "desc", Query: &po.SysUser{Username: "user"},
	}).Return(testUserList, nil)
	userDao.EXPECT().GetUserCount(gomock.Any(), &po.SysUser{Username: "user"}).Return(int64(10), nil)
	userDao.EXPECT().GetRolesByUserID(gomock.Any(), uint(1)).
		Return([]*po.SysRole{{Name: "1"}, {Name: "2"}, {Name: "3"}}, nil)
	userDao.EXPECT().GetRolesByUserID(gomock.Any(), uint(2)).
		Return([]*po.SysRole{{Name: "1"}, {Name: "3"}, {Name: "7"}}, nil)
	userDao.EXPECT().GetRolesByUserID(gomock.Any(), uint(3)).
		Return([]*po.SysRole{{Name: "1"}, {Name: "2"}, {Name: "8"}}, nil)

	mock.ExpectBegin()
	mock.ExpectCommit()
	global.Mysql = gormDB
	userService := NewSysUserService(userDao, nil)
	pageInfo, err2 := userService.GetSysUserList(&dto.PageQuery{
		Page: 1, PageSize: 3, SortProperty: "name", SortRule: "desc", Query: &po.SysUser{Username: "user"},
	})
	assert.Nil(t, err2)
	assert.Equal(t, &dto.PageInfo{
		Size:  3,
		Total: 10,
		List: []*dto.SysUserDtoForList{
			{
				ID: 1, Username: "user1", LoginName: "loginName1", Email: "123456@qq.com", Phone: "12345678910",
				UpdateAt: utils.Time(testUserList[0].UpdatedAt),
				Roles:    []string{"1", "2", "3"},
			},
			{
				ID: 2, Username: "user2", LoginName: "loginName2", Email: "123457@qq.com", Phone: "12345678911",
				UpdateAt: utils.Time(testUserList[1].UpdatedAt),
				Roles:    []string{"1", "3", "7"},
			},
			{
				ID: 3, Username: "user3", LoginName: "loginName3", Email: "123458@qq.com", Phone: "12345678912",
				UpdateAt: utils.Time(testUserList[2].UpdatedAt),
				Roles:    []string{"1", "2", "8"},
			},
		},
	}, pageInfo)

}

func TestSysUserService_GetAllSimpleRole(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	roleDao := mock.NewMockSysRoleDao(mockCtl)
	roleDao.EXPECT().GetAllSimpleRoleList(gomock.Any()).Return([]*po.SysRole{
		{Model: gorm.Model{ID: 1}, Name: "role1"},
		{Model: gorm.Model{ID: 2}, Name: "role2"},
		{Model: gorm.Model{ID: 3}, Name: "role3"},
	}, nil)
	roleDao.EXPECT().GetAllSimpleRoleList(gomock.Any()).Return(nil, gorm.ErrInvalidDB)

	userService := NewSysUserService(nil, roleDao)
	roleList, err := userService.GetAllSimpleRole()
	assert.Equal(t, []*dto.SimpleRoleDto{
		{ID: 1, Name: "role1"}, {ID: 2, Name: "role2"}, {ID: 3, Name: "role3"},
	}, roleList)
	assert.Nil(t, err)
	roleList, err = userService.GetAllSimpleRole()
	assert.Nil(t, roleList)
	assert.Equal(t, e.ErrMysql, err)
}

func TestSysUserService_UpdateUserRoles(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	userDao := mock.NewMockSysUserDao(mockCtl)
	db, mock, err := sqlmock.New()
	defer db.Close()
	assert.Nil(t, err)
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		SkipInitializeWithVersion: true,
		Conn:                      db,
	}), &gorm.Config{})
	assert.Nil(t, err)
	global.Mysql = gormDB

	// 测试1
	userService := NewSysUserService(userDao, nil)
	userDao.EXPECT().DeleteUserRoleByUserID(gomock.Any(), uint(1)).Return(nil)
	userDao.EXPECT().InsertRolesToUser(gomock.Any(), uint(1), []uint{1, 2, 3})
	mock.ExpectBegin()
	mock.ExpectCommit()
	err2 := userService.UpdateUserRoles(1, []uint{1, 2, 3})
	assert.Nil(t, err2)

	userDao.EXPECT().DeleteUserRoleByUserID(gomock.Any(), uint(2)).Return(gorm.ErrInvalidDB)
	mock.ExpectBegin()
	mock.ExpectRollback()
	err2 = userService.UpdateUserRoles(2, []uint{1, 2, 3})
	assert.Equal(t, e.ErrMysql, err2)

	userDao.EXPECT().DeleteUserRoleByUserID(gomock.Any(), uint(3)).Return(nil)
	userDao.EXPECT().InsertRolesToUser(gomock.Any(), uint(3), []uint{1, 2, 3}).Return(gorm.ErrInvalidDB)
	mock.ExpectBegin()
	mock.ExpectRollback()
	err2 = userService.UpdateUserRoles(3, []uint{1, 2, 3})
	assert.Equal(t, e.ErrMysql, err2)
}
