package service

import (
	conf "FanCode/config"
	"FanCode/constants"
	"FanCode/dao"
	e "FanCode/error"
	"FanCode/global"
	"FanCode/models/dto"
	"FanCode/models/po"
	"github.com/gin-gonic/gin"
	"log"
)

type UserCodeService interface {
	// SaveUserCode 保存用户代码
	SaveUserCode(ctx *gin.Context, userCode *po.UserCode) *e.Error
	// GetUserCode 读取用户代码
	GetUserCode(ctx *gin.Context, problemID uint, language constants.LanguageType) (string, *e.Error)
	// GetUserCodeByProblemID 根据题目id获取用户代码，无语言类型
	GetUserCodeByProblemID(ctx *gin.Context, problemID uint) (*po.UserCode, *e.Error)
	// GetProblemTemplateCode 获取题目的模板代码
	GetProblemTemplateCode(problemID uint, language string) (string, *e.Error)
}

type userCodeService struct {
	codeDao dao.UserCodeDao
}

func NewUserCodeService(config *conf.AppConfig, userCodeDao dao.UserCodeDao) UserCodeService {
	return &userCodeService{
		codeDao: userCodeDao,
	}
}

// SaveUserCode 保存用户代码
func (u *userCodeService) SaveUserCode(ctx *gin.Context, userCode *po.UserCode) *e.Error {
	userInfo := ctx.Keys["user"].(*dto.UserInfo)
	userCode.UserID = userInfo.ID
	exist, err := u.codeDao.CheckUserCode(global.Mysql, userInfo.ID, userCode.ProblemID, constants.LanguageType(userCode.Language))
	if err != nil {
		log.Println(err)
		return e.ErrUnknown
	}
	if !exist {
		err = u.codeDao.InsertUserCode(global.Mysql, userCode)
	} else {
		var code *po.UserCode
		code, err = u.codeDao.GetUserCode(global.Mysql, userInfo.ID, userCode.ProblemID, constants.LanguageType(userCode.Language))
		if err != nil {
			log.Println(err)
			return e.ErrUnknown
		}
		code.Code = userCode.Code
		err = u.codeDao.UpdateUserCode(global.Mysql, code)
	}
	if err != nil {
		log.Println(err)
		return e.ErrUnknown
	}
	return nil
}

// GetUserCode 读取用户代码
func (u *userCodeService) GetUserCode(ctx *gin.Context, problemId uint, language constants.LanguageType) (string, *e.Error) {
	userInfo := ctx.Keys["user"].(*dto.UserInfo)
	exist, err := u.codeDao.CheckUserCode(global.Mysql, userInfo.ID, problemId, language)
	if err != nil {
		log.Println(err)
		return "", e.ErrUnknown
	}
	// 如果用户代码不存在，那么读取模板
	if !exist {
		// 读取acm模板
		code, err := getAcmCodeTemplate(language)
		if err != nil {
			return "", e.ErrProblemGetFailed
		}
		return code, nil
	}
	var code *po.UserCode
	code, err = u.codeDao.GetUserCode(global.Mysql, userInfo.ID, problemId, language)
	if err != nil {
		log.Println(err)
		return "", e.ErrUnknown
	}
	return code.Code, nil
}

// GetUserCodeByProblemID 根据题目id获取用户代码，无语言类型
func (u *userCodeService) GetUserCodeByProblemID(ctx *gin.Context, problemId uint) (*po.UserCode, *e.Error) {
	userInfo := ctx.Keys["user"].(*dto.UserInfo)
	codeList, err := u.codeDao.GetUserCodeListByProblemID(global.Mysql, userInfo.ID, problemId)
	if err != nil {
		log.Println(err)
		return nil, e.ErrUnknown
	}

	if len(codeList) != 0 {
		return codeList[0], nil
	}
	// 如果用户代码不存在，那么读取模板
	language := constants.LanguageC
	// 读取acm模板
	code, err := getAcmCodeTemplate(language)
	if err != nil {
		return nil, e.ErrProblemGetFailed
	}
	return &po.UserCode{
		Code:     code,
		Language: string(language),
	}, nil
}

func (u *userCodeService) GetProblemTemplateCode(problemID uint, language string) (string, *e.Error) {
	// 读取acm模板
	code, err := getAcmCodeTemplate(constants.LanguageType(language))
	if err != nil {
		return "", e.ErrProblemGetFailed
	}
	return code, nil
}
