package service

import (
	"FanCode/dao"
	e "FanCode/error"
	"FanCode/global"
	"FanCode/models/dto"
	"FanCode/models/po"
	"FanCode/utils"
	"log"
)

type AuthService interface {
	// Login 用户登录
	Login(loginName string, password string) (string, *e.Error)
	// Register 注册
	Register(user *po.SysUser) *e.Error
	// ChangePassword 改密码
	ChangePassword(loginName, oldPassword, newPassword string) *e.Error
}

type authService struct {
}

func NewAuthService() AuthService {
	return &authService{}
}

func (u *authService) Register(user *po.SysUser) *e.Error {
	if user.Username == "" {
		user.Username = "fancoder"
		return nil
	}
	b, err := dao.CheckLoginName(global.Mysql, user.LoginName)
	if err != nil {
		return e.ErrUserUnknownError
	}
	if b {
		return e.ErrUserNameIsExist
	}
	if len(user.Password) < 6 {
		return e.ErrUserPasswordNotEnoughAccuracy
	}
	//进行注册操作
	newPassword, err := utils.GetPwd(user.Password)
	if err != nil {
		return e.ErrPasswordEncodeFailed
	}
	user.Password = string(newPassword)
	// 设置enable
	//插入
	err = dao.InsertUser(global.Mysql, user)
	if err != nil {
		log.Println(err)
		return e.ErrUserCreationFailed
	} else {
		return nil
	}
}

func (u *authService) Login(userLoginName string, password string) (string, *e.Error) {

	user, userErr := dao.GetUserByLoginName(global.Mysql, userLoginName)
	if userErr != nil {
		log.Println(userErr)
		return "", e.ErrUserUnknownError
	}
	if user == nil || user.LoginName == "" {
		return "", e.ErrUserNotExist
	}
	if user == nil || !utils.ComparePwd(user.Password, password) {
		return "", e.ErrUserNameOrPasswordWrong
	}
	token, err := utils.GenerateToken(dto.NewUserInfo(user))
	if err != nil {
		log.Println(err)
		return "", e.ErrUserUnknownError
	}
	return token, nil
}

func (u *authService) ChangePassword(userLoginName, oldPassword, newPassword string) *e.Error {
	//检验用户名
	user, err := dao.GetUserByLoginName(global.Mysql, userLoginName)
	if err != nil {
		log.Println(err)
		return e.ErrUserUnknownError
	}
	if user == nil || user.LoginName == "" {
		return e.ErrUserNotExist
	}
	//检验旧密码
	if !utils.ComparePwd(oldPassword, user.Password) {
		return e.ErrUserNameOrPasswordWrong
	}
	password, getPwdErr := utils.GetPwd(newPassword)
	if getPwdErr != nil {
		log.Println(getPwdErr)
		return e.ErrPasswordEncodeFailed
	}
	user.Password = string(password)
	err = dao.UpdateUser(global.Mysql, user)
	if err != nil {
		return e.ErrUserUnknownError
	}
	return nil
}
