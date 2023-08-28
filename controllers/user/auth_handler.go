package user

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
	// SendRegisterCode 读取注册时的验证码
	SendRegisterCode(ctx *gin.Context)
	// Register 注册
	Register(ctx *gin.Context)
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

func (u *authController) SendRegisterCode(ctx *gin.Context) {
	result := r.NewResult(ctx)
	email := ctx.PostForm("email")
	if email != "" && utils.VerifyEmailFormat(email) {
		result.SimpleErrorMessage("邮箱格式错误")
		return
	}
	// 生成code
	_, err := u.authService.SendRegisterCode(email)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessMessage("验证码发送成功")
}

func (u *authController) Register(ctx *gin.Context) {
	result := r.NewResult(ctx)
	user := &po.SysUser{}
	user.Phone = ctx.PostForm("phone")
	user.Email = ctx.PostForm("email")
	user.Password = ctx.PostForm("password")
	err := u.authService.Register(user)
	if err == nil {
		result.Error(err)
	} else {
		result.SuccessMessage("注册成功")
	}
}

func (u *authController) Login(ctx *gin.Context) {
	result := r.NewResult(ctx)
	//获取并检验用户参数
	userCode := ctx.PostForm("loginName")
	password := ctx.PostForm("password")
	if userCode == "" || password == "" {
		result.Error(e.ErrBadRequest)
		return
	}
	token, err := u.authService.Login(userCode, password)
	if err != nil {
		result.Error(err)
	} else {
		result.SuccessData(token)
	}
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
