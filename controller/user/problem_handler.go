package user

import (
	r "FanCode/models/vo"
	"FanCode/service"
	"github.com/gin-gonic/gin"
)

type ProblemController interface {
	// GetProblemList 读取题目列表
	GetProblemList(ctx *gin.Context)
	// GetProblem 读取题目详细信息
	GetProblem(ctx *gin.Context)
	// GetProblemTemplateCode 读取题目编程文件
	GetProblemTemplateCode(ctx *gin.Context)
}

type problemController struct {
	problemService service.ProblemService
}

func NewProblemController(problemService service.ProblemService) ProblemController {
	return &problemController{
		problemService: problemService,
	}
}

func (p *problemController) GetProblemList(ctx *gin.Context) {
	result := r.NewResult(ctx)
	pageQuery, err := GetPageQueryByQuery(ctx)
	if err != nil {
		result.Error(err)
		return
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
	number := ctx.Param("number")
	language := ctx.Param("language")
	codeType := ctx.Param("codeType")
	code, err := p.problemService.GetProblemTemplateCode(ctx, number, language, codeType)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(code)
}
