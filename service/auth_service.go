package service

import (
	conf "FanCode/config"
	"FanCode/constants"
	"FanCode/dao"
	e "FanCode/error"
	"FanCode/global"
	"FanCode/models/dto"
	"FanCode/models/po"
	"FanCode/utils"
	"github.com/Chain-Zhang/pinyin"
	"gorm.io/gorm"
	"log"
	"time"
)

const (
	RegisterEmailProKey = "emailcode-register-"
	LoginEmailProKey    = "emailcode-login-"
)

type AuthService interface {

	// PasswordLogin 密码登录 account可能是邮箱可能是用户id
	PasswordLogin(account string, password string) (string, *e.Error)
	// EmailLogin 邮箱验证登录
	EmailLogin(email string, code string) (string, *e.Error)
	// SendAuthCode 获取邮件的验证码
	SendAuthCode(email string, kind string) (string, *e.Error)
	// UserRegister 用户注册
	UserRegister(user *po.SysUser, code string) *e.Error
}

type authService struct {
	config     *conf.AppConfig
	sysUserDao dao.SysUserDao
	sysMenuDao dao.SysMenuDao
	sysRoleDao dao.SysRoleDao
}

func NewAuthService(config *conf.AppConfig, userDao dao.SysUserDao, menuDao dao.SysMenuDao, roleDao dao.SysRoleDao) AuthService {
	return &authService{
		config:     config,
		sysUserDao: userDao,
		sysMenuDao: menuDao,
		sysRoleDao: roleDao,
	}
}

func (u *authService) PasswordLogin(account string, password string) (string, *e.Error) {
	var user *po.SysUser
	var userErr error
	if utils.VerifyEmailFormat(account) {
		user, userErr = u.sysUserDao.GetUserByEmail(global.Mysql, account)
	} else {
		user, userErr = u.sysUserDao.GetUserByLoginName(global.Mysql, account)
	}
	if user == nil || userErr == gorm.ErrRecordNotFound {
		return "", e.ErrUserNotExist
	}
	if userErr != nil {
		log.Println(userErr)
		return "", e.ErrUserUnknownError
	}
	// 比较密码
	if !utils.ComparePwd(user.Password, password) {
		return "", e.ErrUserNameOrPasswordWrong
	}
	// 读取菜单
	for i := 0; i < len(user.Roles); i++ {
		var err error
		user.Roles[i].Menus, err = u.sysRoleDao.GetMenusByRoleID(global.Mysql, user.Roles[i].ID)
		if err != nil {
			return "", e.ErrUserUnknownError
		}
	}
	userInfo := dto.NewUserInfo(user)
	token, err := utils.GenerateToken(utils.Claims{
		ID:        userInfo.ID,
		Avatar:    userInfo.Avatar,
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

func (u *authService) EmailLogin(email string, code string) (string, *e.Error) {
	if !utils.VerifyEmailFormat(email) {
		return "", e.ErrUserUnknownError
	}
	// 获取用户
	user, err := u.sysUserDao.GetUserByEmail(global.Mysql, email)
	if err != nil {
		log.Println(err)
		return "", e.ErrUserUnknownError
	}
	if user == nil || user.LoginName == "" {
		return "", e.ErrUserNotExist
	}
	// 检测验证码
	key := LoginEmailProKey + email
	result, err2 := global.Redis.Get(key).Result()
	if err2 != nil {
		log.Println(err)
		return "", e.ErrUserUnknownError
	}
	if result != code {
		return "", e.ErrLoginCodeWrong
	}
	// 获取菜单
	for i := 0; i < len(user.Roles); i++ {
		var err error
		user.Roles[i].Menus, err = u.sysRoleDao.GetMenusByRoleID(global.Mysql, user.Roles[i].ID)
		if err != nil {
			return "", e.ErrUserUnknownError
		}
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

func (u *authService) SendAuthCode(email string, kind string) (string, *e.Error) {
	if kind == "register" {
		f, err := u.sysUserDao.CheckEmail(global.Mysql, email)
		if err != nil {
			return "", e.ErrUserUnknownError
		}
		if f {
			return "", e.ErrUserEmailIsExist
		}
	}

	var subject string
	if kind == "register" {
		subject = "fancode注册验证码"
	} else if kind == "login" {
		subject = "fancode登录验证码"
	}
	// 发送code
	code := utils.GetRandomNumber(6)
	message := utils.EmailMessage{
		To:      []string{email},
		Subject: subject,
		Body:    "验证码：" + code,
	}
	err := utils.SendMail(u.config.EmailConfig, message)
	if err != nil {
		log.Println(err)
		return "", e.ErrUserUnknownError
	}
	// 存储到redis
	var key string
	if kind == "register" {
		key = RegisterEmailProKey + email
	} else {
		key = LoginEmailProKey + email
	}
	_, err2 := global.Redis.Set(key, code, 10*time.Minute).Result()
	if err2 != nil {
		log.Println(err2)
		return "", e.ErrUserUnknownError
	}
	return code, nil
}

func (u *authService) UserRegister(user *po.SysUser, code string) *e.Error {
	// 检测是否已注册过
	f, err := u.sysUserDao.CheckEmail(global.Mysql, user.Email)
	if f {
		return e.ErrUserEmailIsExist
	}
	// 检测code
	result := global.Redis.Get(RegisterEmailProKey + user.Email)
	if result.Err() != nil {
		return e.ErrUserUnknownError
	}
	if result.Val() != code {
		return e.ErrRoleUnknownError
	}
	// 设置用户名
	if user.Username == "" {
		user.Username = "fancoder"
		return nil
	}
	// 生成用户名称，唯一
	loginName, err := pinyin.New(user.Username).Split("").Convert()
	if err != nil {
		return e.ErrUserUnknownError
	}
	loginName = loginName + utils.GetRandomNumber(3)
	for i := 0; i < 5; i++ {
		b, err := u.sysUserDao.CheckLoginName(global.Mysql, user.LoginName)
		if err != nil {
			log.Println(err)
			return e.ErrUserUnknownError
		}
		if b {
			loginName = loginName + utils.GetRandomNumber(1)
		} else {
			break
		}
	}
	user.LoginName = loginName
	if len(user.Password) < 6 {
		return e.ErrUserPasswordNotEnoughAccuracy
	}
	//进行注册操作
	newPassword, err := utils.GetPwd(user.Password)
	if err != nil {
		return e.ErrPasswordEncodeFailed
	}
	user.Password = string(newPassword)

	err = global.Mysql.Transaction(func(tx *gorm.DB) error {
		err2 := u.sysUserDao.InsertUser(tx, user)
		if err2 != nil {
			return err
		}
		err2 = u.sysUserDao.InsertRolesToUser(tx, user.ID, []uint{constants.UserID})
		return err2
	})
	if err != nil {
		return e.ErrMysql
	}
	return nil
}
