package service

import (
	"FanCode/dao"
	"FanCode/global"
	"FanCode/models/dto"
	"FanCode/models/po"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAccountService_GetAccountInfo(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	userDao := dao.NewMockSysUserDao(mockCtl)
	// mock数据
	sex := 1
	birthDay := time.Now()
	user := &po.SysUser{
		Avatar:       "https://avatar/menmei",
		Username:     "niliu",
		LoginName:    "niliu",
		Password:     "sdfwer",
		Email:        "2958556459@qq.com",
		Phone:        "123456789",
		Introduction: "测试用户",
		// 1表示男 0表示女
		Sex:      &sex,
		BirthDay: birthDay,
	}
	user.CreatedAt = time.Now()
	userDao.EXPECT().GetUserByID(global.Mysql, uint(1)).Return(user)

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
