package admin

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
type ProblemManagementController interface {
	CheckProblemCode(ctx *gin.Context)
	InsertProblem(ctx *gin.Context)
	DeleteProblem(ctx *gin.Context)
	GetProblemList(ctx *gin.Context)
	// 文件修改需要访问的接口
	GetProblemByID(ctx *gin.Context)
	// UpdateProblem 更新题目
	UpdateProblem(ctx *gin.Context)
	// DownloadProblemFile 下载题目的编程文件
	DownloadProblemFile(ctx *gin.Context)
	// DownloadProblemTemplateFile 下载题目的编程文件的模板文件
	DownloadProblemTemplateFile(ctx *gin.Context)
	// UpdateProblemEnable 设置题目可用
	UpdateProblemEnable(ctx *gin.Context)
}

type problemManagementController struct {
	problemService service.ProblemService
}

func NewProblemManagementController() ProblemManagementController {
	return &problemManagementController{
		problemService: service.NewProblemService(),
	}
}

func (q *problemManagementController) CheckProblemCode(ctx *gin.Context) {
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

func (q *problemManagementController) InsertProblem(ctx *gin.Context) {
	result := r.NewResult(ctx)
	problem := &po.Problem{}
	problem.Number = ctx.PostForm("number")
	problem.Name = ctx.PostForm("name")
	problem.Description = ctx.PostForm("description")
	problem.Title = ctx.PostForm("title")
	difficultyStr := ctx.PostForm("difficulty")
	var err error
	if difficultyStr == "" {
		// 题目难度默认为1
		*problem.Difficulty = 1
	} else {
		*problem.Difficulty, err = strconv.Atoi(difficultyStr)
	}
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	if *problem.Difficulty > 5 || *problem.Difficulty < 1 {
		result.SimpleErrorMessage("题目难度必须在1-5之间")
		return
	}
	//插入
	pID, err2 := q.problemService.InsertProblem(problem)
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.Success("题目添加成功", pID)
}

func (q *problemManagementController) UpdateProblem(ctx *gin.Context) {
	result := r.NewResult(ctx)
	problemIDString := ctx.PostForm("id")
	problemID, err := strconv.Atoi(problemIDString)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	problem := &po.Problem{}
	problem.ID = uint(problemID)
	problem.Number = ctx.PostForm("number")
	problem.Name = ctx.PostForm("name")
	problem.Description = ctx.PostForm("description")
	problem.Title = ctx.PostForm("title")
	problem.Path = ctx.PostForm("path")
	difficultyStr := ctx.PostForm("difficulty")
	var difficulty int
	difficulty, err = strconv.Atoi(difficultyStr)
	problem.Difficulty = &difficulty
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	if *problem.Difficulty > 5 || *problem.Difficulty < 1 {
		result.SimpleErrorMessage("题目难度必须在1-5之间")
		return
	}
	enableStr := ctx.PostForm("enable")
	var enable bool
	enable = enableStr == "true"
	problem.Enable = &(enable)
	file, _ := ctx.FormFile("file")
	err2 := q.problemService.UpdateProblem(problem, ctx, file)
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessData("修改成功")
}

func (q *problemManagementController) DeleteProblem(ctx *gin.Context) {
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
func (q *problemManagementController) GetProblemList(ctx *gin.Context) {
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
	if pageSize > 50 {
		pageSize = 50
	}
	pageInfo, err := q.problemService.GetProblemList(page, pageSize)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(pageInfo)
}

func (q *problemManagementController) GetProblemByID(ctx *gin.Context) {
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

func (q *problemManagementController) DownloadProblemFile(ctx *gin.Context) {
	result := r.NewResult(ctx)
	pidstr := ctx.Param("id")
	pid, err := strconv.Atoi(pidstr)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	q.problemService.DownloadProblemZipFile(ctx, uint(pid))
}

func (q *problemManagementController) DownloadProblemTemplateFile(ctx *gin.Context) {
	q.problemService.DownloadProblemTemplateFile(ctx)
}

func (q *problemManagementController) UpdateProblemEnable(ctx *gin.Context) {
	result := r.NewResult(ctx)
	problemIDStr := ctx.PostForm("problemID")
	problemID, err := strconv.Atoi(problemIDStr)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	enableStr := ctx.PostForm("enable")
	enable := enableStr == "true"
	err2 := q.problemService.UpdateProblemEnable(uint(problemID), enable)
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessMessage("操作成功")
}
