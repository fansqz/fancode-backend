package controllers

import (
	r "FanCode/api_models/result"
	"FanCode/dao"
	"FanCode/models"
	"FanCode/utils"
	"github.com/gin-gonic/gin"
	"log"
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
}

func NewUserController() UserController {
	return &userController{}
}

func (u *userController) Register(ctx *gin.Context) {
	result := r.NewResult(ctx)
	userNumber := ctx.PostForm("number")
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")
	if username == "" {
		username = "fancoder"
		return
	}
	if dao.CheckUserNumber(userNumber) {
		result.SimpleErrorMessage("用户名称已存在")
		return
	}
	if len(password) < 6 {
		result.SimpleErrorMessage("密码不能小于6位")
	}
	//进行注册操作
	newPassword, err := utils.GetPwd(password)
	if err != nil {
		log.Println(err)
		result.SimpleErrorMessage("注册失败")
		return
	}
	user := &models.User{}
	user.Number = userNumber
	user.Password = string(newPassword)
	user.Username = username
	//插入
	dao.InsertUser(user)

	//注册成功返回数据
	if err != nil {
		result.SimpleErrorMessage("注册失败，未知错误")
	} else {
		result.SuccessMessage("注册成功，请登录")
	}
}

func (u *userController) Login(ctx *gin.Context) {
	result := r.NewResult(ctx)
	//获取并检验用户参数
	userNumber := ctx.PostForm("number")
	password := ctx.PostForm("password")
	if userNumber == "" {
		result.SimpleErrorMessage("用户ID不可为空")
		return
	}
	if password == "" {
		result.SimpleErrorMessage("密码不可为空")
		return
	}
	user, userErr := dao.GetUserByUserNumber(userNumber)
	if userErr != nil {
		result.SimpleErrorMessage("系统错误")
		log.Println(userErr)
		return
	}
	if user == nil || !utils.ComparePwd(user.Password, password) {
		result.SimpleErrorMessage("用户名或密码错误")
		return
	}
	token, err := utils.GenerateToken(user)
	if err != nil {
		log.Println(err)
		result.SimpleErrorMessage("系统错误")
		return
	}
	result.SuccessData(token)
}

func (u *userController) ChangePassword(ctx *gin.Context) {
	result := r.NewResult(ctx)
	userNumber := ctx.PostForm("number")
	oldPassword := ctx.PostForm("oldPassword")
	newPassword := ctx.PostForm("newPassword")
	if userNumber == "" {
		result.SimpleErrorMessage("用户名不可为空")
		return
	}
	if oldPassword == "" {
		result.SimpleErrorMessage("请输入原始密码")
		return
	}
	//检验用户名
	user, err := dao.GetUserByUserNumber(userNumber)
	if err != nil {
		log.Println(err)
		result.SimpleErrorMessage("系统错误")
		return
	}
	if user == nil {
		result.SimpleErrorMessage("用户不存在")
		return
	}
	//检验旧密码
	if !utils.ComparePwd(oldPassword, user.Password) {
		result.SimpleErrorMessage("原始密码输入错误")
		return
	}
	password, getPwdErr := utils.GetPwd(newPassword)
	if getPwdErr != nil {
		result.SimpleErrorMessage("系统错误")
		log.Println(getPwdErr)
		return
	}
	user.Password = string(password)
	_ = dao.UpdateUser(user)
	token, daoErr := utils.GenerateToken(user)
	if daoErr != nil {
		result.SimpleErrorMessage("登录失败")
		return
	}
	result.SuccessData(token)
}

func (u *userController) GetUserInfo(ctx *gin.Context) {
	result := r.NewResult(ctx)
	user := ctx.Keys["user"].(*models.User)
	result.SuccessData(user)
}
