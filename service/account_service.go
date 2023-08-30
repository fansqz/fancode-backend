package service

import (
	"FanCode/dao"
	e "FanCode/error"
	"FanCode/global"
	"FanCode/models/dto"
	"FanCode/models/po"
	"FanCode/utils"
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

type AccountService interface {
	UpdateAccountInfo(user *po.SysUser) *e.Error
	ChangePassword(ctx *gin.Context, oldPassword, newPassword string) *e.Error
}

func NewAccountService() AccountService {
	return &accountService{}
}

type accountService struct {
}

func (a *accountService) UpdateAccountInfo(user *po.SysUser) *e.Error {
	err := dao.UpdateUser(global.Mysql, user.ID, map[string]interface{}{
		"avatar":       user.Avatar,
		"username":     user.Username,
		"introduction": user.Introduction,
		"sex":          user.Sex,
		"birth_day":    user.BirthDay,
	})
	if err != nil {
		log.Panicln(err)
		return e.ErrUserUnknownError
	}
	return nil
}

func (u *accountService) ChangePassword(ctx *gin.Context, oldPassword, newPassword string) *e.Error {
	userInfo := ctx.Keys["user"].(*dto.UserInfo)
	//检验用户名
	user, err := dao.GetUserByID(global.Mysql, userInfo.ID)
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
	err = dao.UpdateUser(global.Mysql, user.ID, map[string]interface{}{
		"updated_at": user.UpdatedAt,
		"password":   user.Password,
	})
	if err != nil {
		return e.ErrUserUnknownError
	}
	return nil
}
