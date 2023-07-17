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
	InsertProblem(ctx *gin.Context)
	UpdateProblem(ctx *gin.Context)
	DeleteProblem(ctx *gin.Context)
	GetProblemList(ctx *gin.Context)
	UploadProblemFile(ctx *gin.Context)
}

type problemController struct {
	ProblemService service.ProblemService
}

func NewProblemController() ProblemController {
	return &problemController{
		ProblemService: service.NewProblemService(),
	}
}

func (q *problemController) InsertProblem(ctx *gin.Context) {
	result := r.NewResult(ctx)
	Problem := &po.Problem{}
	Problem.Number = ctx.PostForm("number")
	Problem.Name = ctx.PostForm("name")
	Problem.Description = ctx.PostForm("description")
	Problem.Title = ctx.PostForm("title")
	Problem.Path = ctx.PostForm("path")
	//插入
	err := q.ProblemService.InsertProblem(Problem)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessMessage("题库添加成功")
}

func (q *problemController) UpdateProblem(ctx *gin.Context) {
	result := r.NewResult(ctx)
	ProblemIDString := ctx.PostForm("id")
	quesetionID, err := strconv.Atoi(ProblemIDString)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	Problem := &po.Problem{}
	Problem.ID = uint(quesetionID)
	Problem.Number = ctx.PostForm("number")
	Problem.Name = ctx.PostForm("name")
	Problem.Description = ctx.PostForm("description")
	Problem.Title = ctx.PostForm("title")
	Problem.Path = ctx.PostForm("path")

	err2 := q.ProblemService.InsertProblem(Problem)
	if err != nil {
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
	err2 := q.ProblemService.DeleteProblem(uint(id))
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessData("删除成功")
}

// 读取一个列表的题目
func (q *problemController) GetProblemList(ctx *gin.Context) {
	result := r.NewResult(ctx)
	pageStr := ctx.Query("page")
	pageSizeStr := ctx.Query("pageSize")
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
	Problems, err := q.ProblemService.GetProblemList(page, pageSize)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(Problems)
}

func (q *problemController) UploadProblemFile(ctx *gin.Context) {
	result := r.NewResult(ctx)
	file, err := ctx.FormFile("ProblemFile")
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	ProblemNumber := ctx.PostForm("ProblemNumber")
	// 保存文件到本地
	uploadErr := q.ProblemService.UploadProblemFile(ctx, file, ProblemNumber)
	if uploadErr != nil {
		result.Error(uploadErr)
		return
	}
	result.SuccessData("题目文件上传成功")
}
