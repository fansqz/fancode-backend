package service

import (
	"FanCode/dao"
	e "FanCode/error"
	"FanCode/global"
	"FanCode/models/dto"
	"FanCode/models/po"
	"FanCode/utils"
	"github.com/Chain-Zhang/pinyin"
	"log"
	"math/rand"
	"strconv"
	"time"
)

const (
	Email_Pro_Key = "emailcode-"
	// cos中，用户图片存储的位置
	User_Avatar_Path = "/avatar/user"
	// 图片的网站前缀
	Sys_URL = "http://code.fansqz.com"
)

type AuthService interface {
	// Login 用户登录
	Login(loginName string, password string) (string, *e.Error)
	// SendRegisterCode 获取邮件的验证码
	SendRegisterCode(email string) (string, *e.Error)
	// UserRegister 用户注册
	UserRegister(user *po.SysUser, code string) *e.Error
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
	code := u.getCode(6)
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

func (u *authService) UserRegister(user *po.SysUser, code string) *e.Error {
	if user.Username == "" {
		user.Username = "fancoder"
		return nil
	}
	// 生成用户名称，唯一
	loginName, err := pinyin.New(user.Username).Split("").Convert()
	if err != nil {
		return e.ErrUserUnknownError
	}
	loginName = loginName + u.getCode(3)
	for i := 0; i < 5; i++ {
		b, err2 := dao.CheckLoginName(global.Mysql, user.LoginName)
		if err2 != nil {
			log.Println(err2)
			return e.ErrUserUnknownError
		}
		if b {
			loginName = loginName + u.getCode(1)
		} else {
			break
		}
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
func (u *authService) getCode(number int) string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	a := rnd.Int31n(1000000)
	s := strconv.Itoa(int(a))
	return s[0:number]
}
