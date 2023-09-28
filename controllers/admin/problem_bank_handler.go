package admin

import (
	"FanCode/controllers"
	e "FanCode/error"
	"FanCode/models/po"
	r "FanCode/models/vo"
	"FanCode/service"
	"github.com/gin-gonic/gin"
	"strconv"
)

// ProblemBankManagementController
// @Description: 题库管理相关功能
type ProblemBankManagementController interface {
	// UploadProblemBankIcon 上传题库图标
	UploadProblemBankIcon(ctx *gin.Context)
	// InsertProblemBank 添加题库
	InsertProblemBank(ctx *gin.Context)
	// UpdateProblemBank 更新题库
	UpdateProblemBank(ctx *gin.Context)
	// DeleteProblemBank 删除题库
	DeleteProblemBank(ctx *gin.Context)
	// GetProblemBankList 读取题库列表
	GetProblemBankList(ctx *gin.Context)
	// GetProblemBankByID 读取题库信息
	GetProblemBankByID(ctx *gin.Context)
}

type problemBankManagementController struct {
	problemBankService service.ProblemBankService
}

func NewProblemBankManagementController() ProblemBankManagementController {
	return &problemBankManagementController{
		problemBankService: service.NewProblemBankService(),
	}
}

func (p *problemBankManagementController) UploadProblemBankIcon(ctx *gin.Context) {
	result := r.NewResult(ctx)
	file, err := ctx.FormFile("icon")
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	if file.Size > 2<<20 {
		result.SimpleErrorMessage("文件大小不能超过2m")
		return
	}
	path, err2 := p.problemBankService..UploadAvatar(file)
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessData(path)
}

func (p *problemBankManagementController) InsertProblemBank(ctx *gin.Context) {
	result := r.NewResult(ctx)
	bank := p.getBank(ctx)
	pID, err := p.problemBankService.InsertProblemBank(bank, ctx)
	if err != nil {
		result.Error(err)
		return
	}
	result.Success("题库添加成功", pID)
}

func (p *problemBankManagementController) UpdateProblemBank(ctx *gin.Context) {
	result := r.NewResult(ctx)
	bank := p.getBank(ctx)
	// 读取id
	bankIDString := ctx.PostForm("id")
	bankID, err := strconv.Atoi(bankIDString)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	bank.ID = uint(bankID)
	// 读取文件
	err2 := p.problemBankService.UpdateProblemBank(bank)
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessData("题库修改成功")
}

func (p *problemBankManagementController) DeleteProblemBank(ctx *gin.Context) {
	result := r.NewResult(ctx)
	// 读取id
	bankIDString := ctx.Param("id")
	bankID, err := strconv.Atoi(bankIDString)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	// 判断是否强制删除
	forceDeleteStr := ctx.Param("forceDelete")
	forceDelete := false
	if forceDeleteStr == "true" {
		forceDelete = true
	}
	// 删除题库
	err2 := p.problemBankService.DeleteProblemBank(uint(bankID), forceDelete)
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessData("题库删除成功")
}

func (p *problemBankManagementController) GetProblemBankList(ctx *gin.Context) {
	result := r.NewResult(ctx)
	pageQuery, err := controllers.GetPageQueryByQuery(ctx)
	if err != nil {
		result.Error(err)
		return
	}
	// 读取名称和描述
	bank := &po.ProblemBank{
		Name:        ctx.PostForm("name"),
		Description: ctx.PostForm("description"),
	}
	pageQuery.Query = bank
	pageInfo, err := p.problemBankService.GetProblemBankList(pageQuery)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(pageInfo)
}

func (p *problemBankManagementController) GetProblemBankByID(ctx *gin.Context) {
	result := r.NewResult(ctx)
	ids := ctx.Param("id")
	id, convertErr := strconv.Atoi(ids)
	if convertErr != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	bank, err2 := p.problemBankService.GetProblemBankByID(uint(id))
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessData(bank)
}

func (p *problemBankManagementController) getBank(ctx *gin.Context) *po.ProblemBank {
	bank := &po.ProblemBank{}
	bank.Name = ctx.PostForm("name")
	bank.Description = ctx.PostForm("description")
	bank.Icon = ctx.PostForm("icon")
	return bank
}
