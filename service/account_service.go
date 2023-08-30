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

const (
	// UserAvatarPath cos中，用户图片存储的位置
	UserAvatarPath = "/avatar/user"
	// 图片的网站前缀
	SysURL = "https://code.fansqz.com"
)

type AccountService interface {
	UpdateAccountInfo(user *po.SysUser) *e.Error
	ChangePassword(ctx *gin.Context, oldPassword, newPassword string) *e.Error
	// ResetPassword 重置密码
	ResetPassword(ctx *gin.Context) *e.Error
	// GetActivityMap 获取活动图
	GetActivityMap(ctx *gin.Context, year int) ([]*dto.ActivityItem, *e.Error)
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

func (u *accountService) ResetPassword(ctx *gin.Context) *e.Error {
	user := ctx.Keys["user"].(*dto.UserInfo)
	password := utils.GetRandomPassword(11)
	password2, err := utils.GetPwd(password)
	if err != nil {
		return e.ErrUserUnknownError
	}
	// 更新密码
	tx := global.Mysql.Begin()
	err = dao.UpdateUser(tx, user.ID, map[string]interface{}{
		"password": password2,
	})
	if err != nil {
		tx.Rollback()
		return e.ErrUserUnknownError
	}
	// 发送密码
	message := utils.EmailMessage{
		To:      []string{user.Email},
		Subject: "fancode-重置密码",
		Body:    "新密码：" + password,
	}
	err = utils.SendMail(global.Conf.EmailConfig, message)
	if err != nil {
		tx.Rollback()
		return e.ErrUserUnknownError
	}
	return nil
}

func (u *accountService) GetActivityMap(ctx *gin.Context, year int) ([]*dto.ActivityItem, *e.Error) {
	user := ctx.Keys["user"].(*dto.UserInfo)
	var startDate time.Time
	var endDate time.Time
	// 如果year == 0，获取以今天截至的一年的数据
	if year == 0 {
		startDate = time.Now()
		endDate = time.Date(startDate.Year()-1, startDate.Month(), startDate.Day(),
			0, 0, 0, 0, time.Local)
	} else {
		startDate, endDate = getYearRange(year)
	}
	submissions, err := dao.GetUserSimpleSubmissionsByTime(global.Mysql, user.ID, startDate, endDate)
	if err != nil {
		return nil, e.ErrUserUnknownError
	}
	// 构建活动数据
	m := make(map[string]int, 366)
	for i := 0; i < len(submissions); i++ {
		date := submissions[i].CreatedAt.Format("2006-01-02")
		m[date]++
	}
	answer := make([]*dto.ActivityItem, 366)
	i := 0
	for k, v := range m {
		answer[i] = &dto.ActivityItem{
			Date:  k,
			Count: v,
		}
	}
	return answer, nil
}

func getYearRange(year int) (time.Time, time.Time) {
	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	endDate := time.Date(year, 12, 31, 23, 59, 59, 999999999, time.Local)
	return startDate, endDate
}
