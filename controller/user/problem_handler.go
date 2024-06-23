package user

import (
	"FanCode/constants"
	"FanCode/controller/utils"
	"FanCode/models/po"
	r "FanCode/models/vo"
	"FanCode/service"
	"github.com/gin-gonic/gin"
)

type ProblemController interface {
	// GetProblemList 读取题目列表
	GetProblemList(ctx *gin.Context)
	// GetProblem 读取题目详细信息
	GetProblem(ctx *gin.Context)
	// GetProblemTemplateCode 读取题目的模板代码
	GetProblemTemplateCode(ctx *gin.Context)
	// GetUserCode 获取用户代码
	GetUserCode(ctx *gin.Context)
	// GetUserCodeByProblemID 根据题目id获取用户代码，无语言类型
	GetUserCodeByProblemID(ctx *gin.Context)
	// SaveUserCode 保存用户代码
	SaveUserCode(ctx *gin.Context)
}

type problemController struct {
	problemService  service.ProblemService
	userCodeService service.UserCodeService
}

func NewProblemController(problemService service.ProblemService, userCodeService service.UserCodeService) ProblemController {
	return &problemController{
		problemService:  problemService,
		userCodeService: userCodeService,
	}
}

func (p *problemController) GetProblemList(ctx *gin.Context) {
	result := r.NewResult(ctx)
	pageQuery, err := utils.GetPageQueryByQuery(ctx)
	if err != nil {
		result.Error(err)
		return
	}
	bankIDStr := ctx.Query("bankID")
	if bankIDStr != "" {
		bankID := uint(utils.AtoiOrDefault(bankIDStr, 0))
		pageQuery.Query = &po.Problem{
			BankID: &bankID,
		}
	}
	pageInfo, err := p.problemService.GetUserProblemList(ctx, pageQuery)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(pageInfo)
}

func (p *problemController) GetProblem(ctx *gin.Context) {
	result := r.NewResult(ctx)
	numberStr := ctx.Param("number")
	problem, err := p.problemService.GetProblemByNumber(numberStr)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(problem)
}

func (p *problemController) GetProblemTemplateCode(ctx *gin.Context) {
	result := r.NewResult(ctx)
	problemID := utils.GetIntParamOrDefault(ctx, "problemID", 0)
	language := ctx.Param("language")
	code, err := p.userCodeService.GetProblemTemplateCode(uint(problemID), language)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(code)
}

// GetUserCode 获取用户代码
func (p *problemController) GetUserCode(ctx *gin.Context) {
	result := r.NewResult(ctx)
	problemID := utils.GetIntParamOrDefault(ctx, "problemID", 0)
	language := ctx.Param("language")
	code, err := p.userCodeService.GetUserCode(ctx, uint(problemID), constants.LanguageType(language))
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(code)
}

// SaveUserCode 保存用户代码
func (p *problemController) SaveUserCode(ctx *gin.Context) {
	result := r.NewResult(ctx)
	// 题库id
	code := ctx.PostForm("code")
	language := ctx.PostForm("language")
	problemID := utils.AtoiOrDefault(ctx.PostForm("problemID"), 0)
	userCode := &po.UserCode{
		Code:      code,
		ProblemID: uint(problemID),
		Language:  language,
	}
	if err := p.userCodeService.SaveUserCode(ctx, userCode); err != nil {
		result.Error(err)
		return
	}
	result.SuccessMessage("保存成功")
}

// GetUserCodeByProblemID 根据题目id获取用户代码，无语言类型
func (p *problemController) GetUserCodeByProblemID(ctx *gin.Context) {
	result := r.NewResult(ctx)
	problemID := utils.GetIntParamOrDefault(ctx, "problemID", 0)
	code, err := p.userCodeService.GetUserCodeByProblemID(ctx, uint(problemID))
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(code)
}
