package service

import (
	"FanCode/dao"
	e "FanCode/error"
	"FanCode/global"
	"FanCode/models/po"
	"FanCode/utils"
	"log"
)

type UserService interface {
	// Login 用户登录
	Login(userCode string, password string) (string, *e.Error)
	// Register 注册
	Register(user *po.User) *e.Error
	// ChangePassword 改密码
	ChangePassword(userCode, oldPassword, newPassword string) *e.Error
}

type userService struct {
}

func NewUserService() UserService {
	return &userService{}
}

func (u *userService) Register(user *po.User) *e.Error {
	if user.Username == "" {
		user.Username = "fancoder"
		return nil
	}
	if dao.CheckUserCode(global.Mysql, user.Code) {
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
	//插入
	dao.InsertUser(global.Mysql, user)

	if err != nil {
		log.Println(err)
		return e.ErrUserCreationFailed
	} else {
		return nil
	}
}

func (u *userService) Login(userCode string, password string) (string, *e.Error) {

	user, userErr := dao.GetUserByUserCode(global.Mysql, userCode)
	if userErr != nil {
		log.Println(userErr)
		return "", e.ErrUserUnknownError
	}
	if user == nil || user.Code == "" {
		return "", e.ErrUserNotExist
	}
	if user == nil || !utils.ComparePwd(user.Password, password) {
		return "", e.ErrUserNameOrPasswordWrong
	}
	token, err := utils.GenerateToken(user)
	if err != nil {
		log.Println(err)
		return "", e.ErrUserUnknownError
	}
	return token, nil
}

func (u *userService) ChangePassword(userCode, oldPassword, newPassword string) *e.Error {
	//检验用户名
	user, err := dao.GetUserByUserCode(global.Mysql, userCode)
	if err != nil {
		log.Println(err)
		return e.ErrUserUnknownError
	}
	if user == nil || user.Code == "" {
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
