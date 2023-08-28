package service

import (
	"FanCode/dao"
	e "FanCode/error"
	"FanCode/global"
	"FanCode/models/dto"
	"FanCode/models/po"
	"FanCode/utils"
	"fmt"
	"log"
	"math/rand"
	"time"
)

const (
	Email_Pro_Key = "emailcode-"
)

type AuthService interface {
	// Login 用户登录
	Login(loginName string, password string) (string, *e.Error)
	// Register 注册
	Register(user *po.SysUser) *e.Error
	// SendRegisterCode 获取邮件的验证码
	SendRegisterCode(email string) (string, *e.Error)
	// ChangePassword 改密码
	ChangePassword(loginName, oldPassword, newPassword string) *e.Error
}

type authService struct {
}

func NewAuthService() AuthService {
	return &authService{}
}

func (u *authService) Login(userLoginName string, password string) (string, *e.Error) {

	user, userErr := dao.GetUserByLoginName(global.Mysql, userLoginName)
	if userErr != nil {
		log.Println(userErr)
		return "", e.ErrUserUnknownError
	}
	// 读取菜单
	for i := 0; i < len(user.Roles); i++ {
		var err error
		user.Roles[i].Menus, err = dao.GetMenusByRoleID(global.Mysql, user.Roles[i].ID)
		if err != nil {
			return "", e.ErrUserUnknownError
		}
	}
	if user == nil || user.LoginName == "" {
		return "", e.ErrUserNotExist
	}
	if user == nil || !utils.ComparePwd(user.Password, password) {
		return "", e.ErrUserNameOrPasswordWrong
	}
	userInfo := dto.NewUserInfo(user)
	token, err := utils.GenerateToken(utils.Claims{
		ID:        userInfo.ID,
		Username:  userInfo.Username,
		LoginName: userInfo.LoginName,
		Phone:     userInfo.Phone,
		Email:     userInfo.Email,
		Roles:     userInfo.Roles,
		Menus:     userInfo.Menus,
	})
	if err != nil {
		log.Println(err)
		return "", e.ErrUserUnknownError
	}
	return token, nil
}

func (u *authService) SendRegisterCode(email string) (string, *e.Error) {

	f, err := dao.CheckEmail(global.Mysql, email)
	if err != nil {
		return "", e.ErrUserUnknownError
	}
	if f {
		return "", e.ErrUserEmailIsExist
	}
	// 发送code
	code := u.getCode()
	message := utils.EmailMessage{
		To:      []string{email},
		Subject: "fancode注册验证码",
		Body:    "验证码：" + code,
	}
	err = utils.SendMail(global.Conf.EmailConfig, message)
	if err != nil {
		log.Println(err)
		return "", e.ErrUserUnknownError
	}
	// 存储到redis
	_, err2 := global.Redis.Set(Email_Pro_Key+email, code, 10*time.Minute).Result()
	if err2 != nil {
		log.Println(err2)
		return "", e.ErrUserUnknownError
	}
	return code, nil
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
	user.UpdatedAt = time.Now()
	err = dao.UpdateUser(global.Mysql, user)
	if err != nil {
		return e.ErrUserUnknownError
	}
	return nil
}

// getCode 生成6位验证码
func (u *authService) getCode() string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	vcode := fmt.Sprintf("%06v", rnd.Int31n(1000000))
	return vcode
}
