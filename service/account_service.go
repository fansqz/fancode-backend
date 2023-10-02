package service

import (
	"FanCode/dao"
	e "FanCode/error"
	"FanCode/file_store"
	"FanCode/global"
	"FanCode/models/dto"
	"FanCode/models/po"
	r "FanCode/models/vo"
	"FanCode/utils"
	"github.com/gin-gonic/gin"
	"log"
	"mime/multipart"
	"path"
	"time"
)

const (
	// UserAvatarPath cos中，用户图片存储的位置
	UserAvatarPath = "/avatar/user"
)

type AccountService interface {
	// UploadAvatar 上传头像
	UploadAvatar(file *multipart.FileHeader) (string, *e.Error)
	// ReadAvatar 读取头像
	ReadAvatar(ctx *gin.Context, avatarName string)
	// GetAccountInfo 获取账号信息
	GetAccountInfo(ctx *gin.Context) (*dto.AccountInfo, *e.Error)
	// UpdateAccountInfo 更新账号信息
	UpdateAccountInfo(ctx *gin.Context, user *po.SysUser) *e.Error
	// ChangePassword 修改密码
	ChangePassword(ctx *gin.Context, oldPassword, newPassword string) *e.Error
	// ResetPassword 重置密码
	ResetPassword(ctx *gin.Context) *e.Error
}

func NewAccountService() AccountService {
	return &accountService{}
}

type accountService struct {
}

func (a *accountService) UploadAvatar(file *multipart.FileHeader) (string, *e.Error) {
	cos := file_store.NewImageCOS()
	fileName := file.Filename
	fileName = utils.GetUUID() + "." + path.Base(fileName)
	file2, err := file.Open()
	if err != nil {
		return "", e.ErrBadRequest
	}
	err = cos.SaveFile(UserAvatarPath+"/"+fileName, file2)
	if err != nil {
		log.Println(err)
		return "", e.ErrServer
	}
	return global.Conf.ProUrl + UserAvatarPath + "/" + fileName, nil
}

func (a *accountService) ReadAvatar(ctx *gin.Context, avatarName string) {
	result := r.NewResult(ctx)
	cos := file_store.NewImageCOS()
	bytes, err := cos.ReadFile(UserAvatarPath + "/" + avatarName)
	if err != nil {
		result.Error(e.ErrServer)
		return
	}
	_, _ = ctx.Writer.Write(bytes)
}

func (a *accountService) GetAccountInfo(ctx *gin.Context) (*dto.AccountInfo, *e.Error) {
	user := ctx.Keys["user"].(*dto.UserInfo)
	u, err := dao.GetUserByID(global.Mysql, user.ID)
	if err != nil {
		return nil, e.ErrMysql
	}
	return dto.NewAccountInfo(u), nil
}

func (a *accountService) UpdateAccountInfo(ctx *gin.Context, user *po.SysUser) *e.Error {
	userInfo := ctx.Keys["user"].(*dto.UserInfo)
	user.ID = userInfo.ID
	user.UpdatedAt = time.Now()
	err := dao.UpdateUser(global.Mysql, user)
	if err != nil {
		log.Panicln(err)
		return e.ErrMysql
	}
	return nil
}

func (u *accountService) ChangePassword(ctx *gin.Context, oldPassword, newPassword string) *e.Error {
	userInfo := ctx.Keys["user"].(*dto.UserInfo)
	//检验用户名
	user, err := dao.GetUserByID(global.Mysql, userInfo.ID)
	if err != nil {
		log.Println(err)
		return e.ErrMysql
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
		return e.ErrMysql
	}
	return nil
}

func (u *accountService) ResetPassword(ctx *gin.Context) *e.Error {
	userInfo := ctx.Keys["user"].(*dto.UserInfo)
	password := utils.GetRandomPassword(11)
	password2, err := utils.GetPwd(password)
	if err != nil {
		return e.ErrUserUnknownError
	}
	// 更新密码
	tx := global.Mysql.Begin()
	user := &po.SysUser{}
	user.ID = userInfo.ID
	user.Password = string(password2)
	err = dao.UpdateUser(tx, user)
	if err != nil {
		tx.Rollback()
		return e.ErrMysql
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
