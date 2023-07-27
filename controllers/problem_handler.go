package controllers

import (
	e "FanCode/error"
	"FanCode/models/po"
	r "FanCode/models/vo"
	"FanCode/service"
	"github.com/gin-gonic/gin"
	"strconv"
)

// ProblemController
// @Description: 题目管理相关功能
type ProblemController interface {
	CheckProblemCode(ctx *gin.Context)
	InsertProblem(ctx *gin.Context)
	UpdateProblem(ctx *gin.Context)
	DeleteProblem(ctx *gin.Context)
	GetProblemList(ctx *gin.Context)
	// 上传题目文件
	UploadProblemFile(ctx *gin.Context)
	// 这两个方法用于获取题目全部信息，题目元数据，题目文件列表
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

func (q *problemController) CheckProblemCode(ctx *gin.Context) {
	result := r.NewResult(ctx)
	code := ctx.Param("code")
	b, err := q.problemService.CheckProblemCode(code)
	if err != nil {
		result.Error(err)
	}
	if !b {
		result.Success("编号重复，请更换其他编号", b)
	} else {
		result.Success("编号可用", b)
	}
}

func (q *problemController) InsertProblem(ctx *gin.Context) {
	result := r.NewResult(ctx)
	problem := &po.Problem{}
	problem.Code = ctx.PostForm("code")
	problem.Name = ctx.PostForm("name")
	problem.Description = ctx.PostForm("description")
	problem.Title = ctx.PostForm("title")
	//插入
	pID, err := q.problemService.InsertProblem(problem)
	if err != nil {
		result.Error(err)
		return
	}
	result.Success("题库添加成功", pID)
}

func (q *problemController) UpdateProblem(ctx *gin.Context) {
	result := r.NewResult(ctx)
	problemIDString := ctx.PostForm("id")
	problemID, err := strconv.Atoi(problemIDString)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	problem := &po.Problem{}
	problem.ID = uint(problemID)
	problem.Code = ctx.PostForm("code")
	problem.Name = ctx.PostForm("name")
	problem.Description = ctx.PostForm("description")
	problem.Title = ctx.PostForm("title")
	problem.Path = ctx.PostForm("path")
	file, _ := ctx.FormFile("file")
	err2 := q.problemService.UpdateProblem(problem, ctx, file)
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessData("修改成功")
}

func (q *problemController) DeleteProblem(ctx *gin.Context) {
	result := r.NewResult(ctx)
	ids := ctx.Param("id")
	id, convertErr := strconv.Atoi(ids)
	if convertErr != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	err2 := q.problemService.DeleteProblem(uint(id))
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessData("删除成功")
}

// GetProblemList 读取一个列表的题目
func (q *problemController) GetProblemList(ctx *gin.Context) {
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
	pageInfo, err := q.problemService.GetProblemList(page, pageSize)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(pageInfo)
}

func (q *problemController) GetProblemByID(ctx *gin.Context) {
	result := r.NewResult(ctx)
	ids := ctx.Param("id")
	id, convertErr := strconv.Atoi(ids)
	if convertErr != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	problem, err2 := q.problemService.GetProblemByID(uint(id))
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessData(problem)
}

func (q *problemController) UploadProblemFile(ctx *gin.Context) {
	result := r.NewResult(ctx)
	file, err := ctx.FormFile("problemFile")
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	problemCode := ctx.PostForm("problemCode")
	// 保存文件到本地
	uploadErr := q.problemService.UploadProblemFile(ctx, file, problemCode)
	if uploadErr != nil {
		result.Error(uploadErr)
		return
	}
	result.SuccessData("题目文件上传成功")
}
