package controllers

import (
	e "FanCode/error"
	"FanCode/models/dto"
	"FanCode/models/po"
	r "FanCode/models/vo"
	"FanCode/service"
	"FanCode/utils"
	"github.com/gin-gonic/gin"
)

type AuthController interface {
	// Login 用户登录
	Login(ctx *gin.Context)
	// SendAuthCode 发送验证码
	SendAuthCode(ctx *gin.Context)
	// UserRegister 用户注册
	UserRegister(ctx *gin.Context)
	// GetUserInfo 根据token获取用户信息
	GetUserInfo(ctx *gin.Context)
	// ChangePassword 改密码
	ChangePassword(ctx *gin.Context)
}

type authController struct {
	authService service.AuthService
}

func NewAuthController() AuthController {
	return &authController{
		authService: service.NewAuthService(),
	}
}

func (u *authController) SendAuthCode(ctx *gin.Context) {
	result := r.NewResult(ctx)
	email := ctx.PostForm("email")
	kind := ctx.PostForm("type")
	if email != "" && !utils.VerifyEmailFormat(email) {
		result.SimpleErrorMessage("邮箱格式错误")
		return
	}
	// 生成code
	_, err := u.authService.SendAuthCode(email, kind)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessMessage("验证码发送成功")
}

func (u *authController) UserRegister(ctx *gin.Context) {
	result := r.NewResult(ctx)
	user := &po.SysUser{}
	user.Email = ctx.PostForm("email")
	code := ctx.PostForm("code")
	user.Username = ctx.PostForm("username")
	user.Password = ctx.PostForm("password")
	err := u.authService.UserRegister(user, code)
	if err != nil {
		result.Error(err)
	} else {
		result.SuccessMessage("注册成功")
	}
}

func (u *authController) Login(ctx *gin.Context) {
	result := r.NewResult(ctx)
	//获取并检验用户参数
	kind := ctx.PostForm("type")
	account := ctx.PostForm("account")
	email := ctx.PostForm("email")
	password := ctx.PostForm("password")
	code := ctx.PostForm("code")
	if kind != "password" && kind != "email" {
		result.Error(e.ErrBadRequest)
		return
	} else if kind == "password" && (account == "" || password == "") {
		result.Error(e.ErrBadRequest)
		return
	} else if kind == "email" && (email == "" || code == "") {
		result.Error(e.ErrBadRequest)
		return
	}
	// 登录
	var token string
	var err *e.Error
	if kind == "password" {
		token, err = u.authService.PasswordLogin(account, password)
	} else if kind == "email" {
		token, err = u.authService.EmailLogin(email, code)
	} else {
		result.Error(e.ErrLoginType)
		return
	}
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(token)

}

func (u *authController) ChangePassword(ctx *gin.Context) {
	result := r.NewResult(ctx)
	loginName := ctx.PostForm("loginName")
	oldPassword := ctx.PostForm("oldPassword")
	newPassword := ctx.PostForm("newPassword")
	if loginName == "" || oldPassword == "" {
		result.Error(e.ErrBadRequest)
		return
	}
	err := u.authService.ChangePassword(loginName, oldPassword, newPassword)
	if err != nil {
		result.Error(err)
	} else {
		result.SuccessMessage("Password changed")
	}
}

func (u *authController) GetUserInfo(ctx *gin.Context) {
	result := r.NewResult(ctx)
	user := ctx.Keys["user"].(*dto.UserInfo)
	result.SuccessData(user)
}
