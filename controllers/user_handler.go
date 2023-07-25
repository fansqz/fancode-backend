package controllers

import (
	e "FanCode/error"
	"FanCode/models/po"
	r "FanCode/models/vo"
	"FanCode/service"
	"github.com/gin-gonic/gin"
)

// UserController
// @Description: 用户账号相关功能
type UserController interface {
	// Login 用户登录
	Login(ctx *gin.Context)
	// Register 注册
	Register(ctx *gin.Context)
	// 根据token获取用户信息
	GetUserInfo(ctx *gin.Context)
	// ChangePassword 改密码
	ChangePassword(ctx *gin.Context)
}

type userController struct {
	userService service.UserService
}

func NewUserController() UserController {
	return &userController{
		userService: service.NewUserService(),
	}
}

func (u *userController) Register(ctx *gin.Context) {
	result := r.NewResult(ctx)
	user := &po.User{}
	user.Code = ctx.PostForm("code")
	user.Password = ctx.PostForm("password")
	user.Username = ctx.PostForm("username")
	err := u.userService.Register(user)
	if err == nil {
		result.Error(err)
	} else {
		result.SuccessMessage("注册成功")
	}
}

func (u *userController) Login(ctx *gin.Context) {
	result := r.NewResult(ctx)
	//获取并检验用户参数
	userCode := ctx.PostForm("code")
	password := ctx.PostForm("password")
	if userCode == "" || password == "" {
		result.Error(e.ErrBadRequest)
		return
	}
	token, err := u.userService.Login(userCode, password)
	if err != nil {
		result.Error(err)
	} else {
		result.SuccessData(token)
	}
}

func (u *userController) ChangePassword(ctx *gin.Context) {
	result := r.NewResult(ctx)
	userCode := ctx.PostForm("code")
	oldPassword := ctx.PostForm("oldPassword")
	newPassword := ctx.PostForm("newPassword")
	if userCode == "" || oldPassword == "" {
		result.Error(e.ErrBadRequest)
		return
	}
	err := u.userService.ChangePassword(userCode, oldPassword, newPassword)
	if err != nil {
		result.Error(err)
	} else {
		result.SuccessMessage("Password changed")
	}
}

func (u *userController) GetUserInfo(ctx *gin.Context) {
	result := r.NewResult(ctx)
	user := ctx.Keys["user"].(*po.User)
	result.SuccessData(user)
}
