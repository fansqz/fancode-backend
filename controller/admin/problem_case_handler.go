package admin

import (
	"FanCode/controller/utils"
	"FanCode/models/po"
	r "FanCode/models/vo"
	"FanCode/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProblemCaseManagementController interface {
	// GetProblemCaseList 读取题目用例
	GetProblemCaseList(ctx *gin.Context)
	// GetProblemCaseByID 获取用例
	GetProblemCaseByID(ctx *gin.Context)
	// InsertProblemCase 给题目添加用例
	InsertProblemCase(ctx *gin.Context)
	// UpdateProblemCase 更新用例
	UpdateProblemCase(ctx *gin.Context)
	// DeleteProblemCase 删除用例
	DeleteProblemCase(ctx *gin.Context)
}

type problemCaseManagementController struct {
	problemCaseService service.ProblemCaseService
}

func NewProblemCaseManagementController(pcs service.ProblemCaseService) ProblemCaseManagementController {
	return &problemCaseManagementController{
		problemCaseService: pcs,
	}
}

func (p *problemCaseManagementController) GetProblemCaseList(ctx *gin.Context) {
	result := r.NewResult(ctx)
	pageQuery, err := utils.GetPageQueryByQuery(ctx)
	if err != nil {
		result.Error(err)
		return
	}
	pcase := &po.ProblemCase{}
	pcase.Name = ctx.Query("name")
	pcase.ProblemID = uint(utils.GetIntQueryOrDefault(ctx, "problemID", 0))
	pageQuery.Query = pcase
	pageInfo, err := p.problemCaseService.GetProblemCaseList(pageQuery)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(pageInfo)
}

func (p *problemCaseManagementController) GetProblemCaseByID(ctx *gin.Context) {
	result := r.NewResult(ctx)
	caseID := uint(utils.GetIntQueryOrDefault(ctx, "id", 0))
	answer, err := p.problemCaseService.GetProblemCaseByID(caseID)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(answer)
}

func (p *problemCaseManagementController) InsertProblemCase(ctx *gin.Context) {
	result := r.NewResult(ctx)
	pcase := &po.ProblemCase{
		ProblemID: uint(utils.AtoiOrDefault(ctx.PostForm("problemID"), 0)),
		Name:      ctx.PostForm("name"),
		Input:     ctx.PostForm("input"),
		Output:    ctx.PostForm("output"),
	}
	id, err := p.problemCaseService.InsertProblemCase(pcase)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(id)
}

func (p *problemCaseManagementController) UpdateProblemCase(ctx *gin.Context) {
	result := r.NewResult(ctx)
	pcase := &po.ProblemCase{
		Model: gorm.Model{
			ID: uint(utils.AtoiOrDefault(ctx.PostForm("id"), 0)),
		},
		Name:   ctx.PostForm("name"),
		Input:  ctx.PostForm("input"),
		Output: ctx.PostForm("output"),
	}
	if err := p.problemCaseService.UpdateProblemCase(pcase); err != nil {
		result.Error(err)
		return
	}
	result.SuccessMessage("用例添加成功")
}

func (p *problemCaseManagementController) DeleteProblemCase(ctx *gin.Context) {
	result := r.NewResult(ctx)
	id := utils.GetIntParamOrDefault(ctx, "id", 0)
	if err := p.problemCaseService.DeleteProblemCaseByID(uint(id)); err != nil {
		result.Error(err)
		return
	}
	result.SuccessMessage("用例删除成功")
}
