package controllers

import (
	e "FanCode/error"
	"FanCode/models/po"
	r "FanCode/models/vo"
	"FanCode/service"
	"FanCode/utils"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

// 关于一些账号信息的handler
type AccountController interface {
	// UpdateAccountInfo 更新账号信息
	UpdateAccountInfo(ctx *gin.Context)
	// ChangePassword 修改密码
	ChangePassword(ctx *gin.Context)
	// GetUserActivity 获取用户活动图
	GetUserActivity(ctx *gin.Context)
	// ResetPassword 重置密码
	ResetPassword(ctx *gin.Context)
}

type accountController struct {
	accountService service.AccountService
}

func NewAccountController() AccountController {
	return &accountController{}
}

func (a *accountController) UpdateAccountInfo(ctx *gin.Context) {
	result := r.NewResult(ctx)
	user := &po.SysUser{}
	user.Avatar = ctx.PostForm("avatar")
	user.Username = ctx.PostForm("username")
	user.Introduction = ctx.PostForm("introduction")
	sex := ctx.PostForm("sex")
	sex2, err := strconv.Atoi(sex)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	user.Sex = &sex2
	birthDay := ctx.PostForm("birthDay")
	t := utils.Time{}
	err = t.UnmarshalJSON([]byte(birthDay))
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	user.BirthDay = time.Time(t)

}

func (a *accountController) ChangePassword(ctx *gin.Context) {
	result := r.NewResult(ctx)
	oldPassword := ctx.PostForm("oldPassword")
	newPassword := ctx.PostForm("newPassword")
	if oldPassword == "" {
		result.Error(e.ErrBadRequest)
		return
	}
	if newPassword == "" {
		result.Error(e.ErrBadRequest)
	}
	err := a.accountService.ChangePassword(ctx, oldPassword, newPassword)
	if err != nil {
		result.Error(err)
	}
	result.SuccessMessage("修改成功，请重新登录")
}

func (a *accountController) ResetPassword(ctx *gin.Context) {
	result := r.NewResult(ctx)
	err := a.accountService.ResetPassword(ctx)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessMessage("重置成功，请留意邮箱")
}

func (a *accountController) GetUserActivity(ctx *gin.Context) {
	result := r.NewResult(ctx)
	yearStr := ctx.PostForm("year")
	var year int
	if yearStr == "" {
		year = 0
	} else {
		var b bool
		year, b = checkYear(yearStr)
		if !b {
			result.Error(e.ErrBadRequest)
			return
		}
	}
	activityMap, err := a.accountService.GetActivityMap(ctx, year)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(activityMap)
}

func checkYear(str string) (int, bool) {
	year, err := strconv.Atoi(str)
	if err != nil {
		return 0, false
	}

	currentYear := time.Now().Year()
	b := year > 2022 && year <= currentYear
	return year, b
}
