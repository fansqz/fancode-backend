package controllers

import (
	e "FanCode/error"
	"FanCode/models/po"
	r "FanCode/models/vo"
	"FanCode/service"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

// AccountController 关于一些账号信息的handler
type AccountController interface {
	// UploadAvatar 上传头像
	UploadAvatar(ctx *gin.Context)
	// ReadAvatar 读取头像
	ReadAvatar(ctx *gin.Context)
	// GetAccountInfo 获取账号信息
	GetAccountInfo(ctx *gin.Context)
	// UpdateAccountInfo 更新账号信息
	UpdateAccountInfo(ctx *gin.Context)
	// ChangePassword 修改密码
	ChangePassword(ctx *gin.Context)
	// ResetPassword 重置密码
	ResetPassword(ctx *gin.Context)
}

type accountController struct {
	accountService service.AccountService
}

func NewAccountController() AccountController {
	return &accountController{
		accountService: service.NewAccountService(),
	}
}

func (a *accountController) UploadAvatar(ctx *gin.Context) {
	result := r.NewResult(ctx)
	file, err := ctx.FormFile("avatar")
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	if file.Size > 2<<20 {
		result.SimpleErrorMessage("文件大小不能超过2m")
		return
	}
	path, err2 := a.accountService.UploadAvatar(file)
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessData(path)
}

func (a *accountController) ReadAvatar(ctx *gin.Context) {
	avatarName := ctx.Param("avatarName")
	a.accountService.ReadAvatar(ctx, avatarName)
}

func (a *accountController) GetAccountInfo(ctx *gin.Context) {
	result := r.NewResult(ctx)
	accountInfo, err := a.accountService.GetAccountInfo(ctx)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(accountInfo)
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
	t, err2 := time.ParseInLocation("2006-01-02 15:04:05", birthDay, time.Local)
	if err2 != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	user.BirthDay = t
	err3 := a.accountService.UpdateAccountInfo(ctx, user)
	if err3 != nil {
		result.Error(err3)
		return
	}
	result.SuccessMessage("提交成功，重新登录可更新数据")
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
