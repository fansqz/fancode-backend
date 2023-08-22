package user

import (
	e "FanCode/error"
	r "FanCode/models/vo"
	"FanCode/service"
	"github.com/gin-gonic/gin"
	"strconv"
)

type ProblemController interface {
	// 读取题目列表
	GetProblemList(ctx *gin.Context)
	// 读取题目详细信息
	GetProblemByID(ctx *gin.Context)
}

type problemController struct {
	problemService service.ProblemService
}

func NewProblemController() ProblemController {
	return &problemController{
		problemService: service.NewProblemService(),
	}
}

func (p *problemController) GetProblemList(ctx *gin.Context) {
	result := r.NewResult(ctx)
	pageStr := ctx.Param("page")
	pageSizeStr := ctx.Param("pageSize")
	var page int
	var pageSize int
	var convertErr error
	page, convertErr = strconv.Atoi(pageStr)
	if convertErr != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	pageSize, convertErr = strconv.Atoi(pageSizeStr)
	if convertErr != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	pageInfo, err := p.problemService.GetProblemList(page, pageSize)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(pageInfo)
}

func (p *problemController) GetProblemByID(ctx *gin.Context) {
	result := r.NewResult(ctx)
	ids := ctx.Param("id")
	id, convertErr := strconv.Atoi(ids)
	if convertErr != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	problem, err2 := p.problemService.GetProblemByID(uint(id))
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessData(problem)
}

func (p *problemController) GetProblemCode(ctx *gin.Context) {
}
