package service

import (
	"FanCode/dao/mock"
	"FanCode/global"
	"FanCode/models/dto"
	"FanCode/models/po"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
	"time"
)

func TestAccountService_GetAccountInfo(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	userDao := mock.NewMockSysUserDao(mockCtl)
	// mock数据
	birthDay := time.Now()
	user := &po.SysUser{
		Avatar:       "https://avatar/menmei",
		Username:     "niliu",
		LoginName:    "niliu",
		Password:     "sdfwer",
		Email:        "2958556459@qq.com",
		Phone:        "123456789",
		Introduction: "测试用户",
		Sex:          1,
		BirthDay:     birthDay,
	}
	user.CreatedAt = time.Now()
	userDao.EXPECT().GetUserByID(global.Mysql, uint(1)).Return(user, nil)

	accountService := NewAccountService(userDao)
	ctx := &gin.Context{}
	ctx.Keys = make(map[string]interface{})
	ctx.Keys["user"] = &dto.UserInfo{
		ID: 1,
	}
	accountInfo, _ := accountService.GetAccountInfo(ctx)
	assert.Equal(t, &dto.AccountInfo{
		Avatar:       "https://avatar/menmei",
		LoginName:    "niliu",
		UserName:     "niliu",
		Email:        "2958556459@qq.com",
		Phone:        "123456789",
		Introduction: "测试用户",
		Sex:          1,
		BirthDay:     birthDay.Format("2006-01-02"),
		CodingAge:    0,
	}, accountInfo)
}

func TestAccountService_UpdateAccountInfo(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	userDao := mock.NewMockSysUserDao(mockCtl)
	// mock数据
	birthDay := time.Now()
	user := &po.SysUser{
		Avatar:       "https://avatar/menmei",
		Username:     "niliu",
		LoginName:    "loginName",
		Password:     "password",
		Email:        "2958556459@qq.com",
		Phone:        "123456789",
		Introduction: "测试用户",
		// 1表示男 2表示女
		Sex:      1,
		BirthDay: birthDay,
	}
	userDao.EXPECT().UpdateUser(global.Mysql, gomock.Any()).Return(nil).Do(
		func(db *gorm.DB, user *po.SysUser) {
			assert.NotEqual(t, user.UpdatedAt, time.Time{})
			user.UpdatedAt = time.Time{}
			user2 := &po.SysUser{
				Avatar:       "https://avatar/menmei",
				Username:     "niliu",
				LoginName:    "",
				Password:     "",
				Email:        "2958556459@qq.com",
				Phone:        "123456789",
				Introduction: "测试用户",
				Sex:          1,
				BirthDay:     birthDay,
			}
			user2.ID = 1
			assert.Equal(t, user, user2)
		})
	ctx := &gin.Context{}
	ctx.Keys = make(map[string]interface{})
	ctx.Keys["user"] = &dto.UserInfo{
		ID: 1,
	}
	accountService := NewAccountService(userDao)
	err := accountService.UpdateAccountInfo(ctx, user)
	assert.Nil(t, err)
}
