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
	user.Number = ctx.PostForm("number")
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
	userNumber := ctx.PostForm("number")
	password := ctx.PostForm("password")
	if userNumber == "" || password == "" {
		result.Error(e.ErrBadRequest)
		return
	}
	token, err := u.userService.Login(userNumber, password)
	if err != nil {
		result.Error(err)
	} else {
		result.SuccessData(token)
	}
}

func (u *userController) ChangePassword(ctx *gin.Context) {
	result := r.NewResult(ctx)
	userNumber := ctx.PostForm("number")
	oldPassword := ctx.PostForm("oldPassword")
	newPassword := ctx.PostForm("newPassword")
	if userNumber == "" || oldPassword == "" {
		result.Error(e.ErrBadRequest)
		return
	}
	err := u.userService.ChangePassword(userNumber, oldPassword, newPassword)
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
