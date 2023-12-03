package admin

import (
	e "FanCode/error"
	"FanCode/models/dto"
	"FanCode/models/po"
	r "FanCode/models/vo"
	"FanCode/service"
	"github.com/gin-gonic/gin"
	"strconv"
)

// ProblemManagementController
// @Description: 题目管理相关功能
type ProblemManagementController interface {
	CheckProblemNumber(ctx *gin.Context)
	InsertProblem(ctx *gin.Context)
	DeleteProblem(ctx *gin.Context)
	GetProblemList(ctx *gin.Context)
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

func NewProblemManagementController(problemService service.ProblemService) ProblemManagementController {
	return &problemManagementController{
		problemService: problemService,
	}
}

func (q *problemManagementController) CheckProblemNumber(ctx *gin.Context) {
	result := r.NewResult(ctx)
	code := ctx.Param("number")
	b, err := q.problemService.CheckProblemNumber(code)
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
	problem, err := q.getProblemByForm(ctx)
	if err != nil {
		result.Error(err)
	}
	//插入
	pID, err2 := q.problemService.InsertProblem(problem, ctx)
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.Success("题目添加成功", pID)
}

func (q *problemManagementController) UpdateProblem(ctx *gin.Context) {
	result := r.NewResult(ctx)
	problem, err2 := q.getProblemByForm(ctx)
	if err2 != nil {
		result.Error(err2)
		return
	}
	// 读取id
	problemIDString := ctx.PostForm("id")
	problemID, err := strconv.Atoi(problemIDString)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	problem.ID = uint(problemID)
	// 读取文件
	file, _ := ctx.FormFile("file")
	err2 = q.problemService.UpdateProblem(problem, ctx, file)
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
	pageQuery, err := GetPageQueryByQuery(ctx)
	if err != nil {
		result.Error(err)
		return
	}
	problem, err := q.getProblemByQuery(ctx)
	if err != nil {
		result.Error(err)
		return
	}
	pageQuery.Query = problem
	pageInfo, err := q.problemService.GetProblemList(pageQuery)
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
	var enable int
	if enableStr == "1" {
		enable = 1
	} else {
		enable = -1
	}
	err2 := q.problemService.UpdateProblemEnable(uint(problemID), enable)
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessMessage("操作成功")
}

func (q *problemManagementController) getProblemByForm(ctx *gin.Context) (*po.Problem, *e.Error) {
	problem := &po.Problem{}
	problem.Number = ctx.PostForm("number")
	problem.Name = ctx.PostForm("name")
	problem.Description = ctx.PostForm("description")
	problem.Title = ctx.PostForm("title")
	difficultyStr := ctx.PostForm("difficulty")
	bankIDStr := ctx.PostForm("bankID")
	problem.Languages = ctx.PostForm("languages")
	enableStr := ctx.PostForm("enable")
	var err error
	// 难度设置
	if difficultyStr == "" {
		// 题目难度默认为1
		problem.Difficulty = 0
	} else {
		problem.Difficulty, err = strconv.Atoi(difficultyStr)
		if err != nil {
			return nil, e.ErrBadRequest
		}
	}
	if problem.Difficulty > 5 || problem.Difficulty < 0 {
		return nil, e.ErrBadRequest
	}
	// 题库id设置
	var bankID int
	if bankIDStr != "" {
		bankID, err = strconv.Atoi(bankIDStr)
		if err != nil {
			return nil, e.ErrBadRequest
		}
		uintBankID := uint(bankID)
		problem.BankID = &uintBankID
	}
	// 是否启用
	if enableStr == "1" {
		problem.Enable = 1
	} else if enableStr == "-1 " {
		problem.Enable = -1
	}
	return problem, nil
}

func (q *problemManagementController) getProblemByQuery(ctx *gin.Context) (*po.Problem, *e.Error) {
	problem := &po.Problem{}
	problem.Number = ctx.Query("number")
	problem.Name = ctx.Query("name")
	problem.Description = ctx.Query("description")
	problem.Title = ctx.Query("title")
	difficultyStr := ctx.Query("difficulty")
	bankIDStr := ctx.Query("bankID")
	problem.Languages = ctx.Query("languages")
	enableStr := ctx.Query("enable")
	var err error
	// 难度设置
	if difficultyStr == "" {
		// 题目难度默认为1
		problem.Difficulty = 0
	} else {
		problem.Difficulty, err = strconv.Atoi(difficultyStr)
		if err != nil {
			return nil, e.ErrBadRequest
		}
	}
	if problem.Difficulty > 5 || problem.Difficulty < 0 {
		return nil, e.ErrBadRequest
	}
	// 题库id设置
	var bankID int
	if bankIDStr != "" {
		bankID, err = strconv.Atoi(bankIDStr)
		if err != nil {
			return nil, e.ErrBadRequest
		}
		uintBankID := uint(bankID)
		problem.BankID = &uintBankID
	}
	// 是否启用
	if enableStr == "1" {
		problem.Enable = 1
	} else if enableStr == "-1" {
		problem.Enable = -1
	}
	return problem, nil
}

func GetPageQueryByQuery(ctx *gin.Context) (*dto.PageQuery, *e.Error) {
	pageStr := ctx.Query("page")
	pageSizeStr := ctx.Query("pageSize")
	var page int
	var pageSize int
	var convertErr error
	page, convertErr = strconv.Atoi(pageStr)
	if convertErr != nil {
		return nil, e.ErrBadRequest
	}
	pageSize, convertErr = strconv.Atoi(pageSizeStr)
	if convertErr != nil {
		return nil, e.ErrBadRequest
	}
	if pageSize > 50 {
		pageSize = 50
	}
	sortProperty := ctx.Query("sortProperty")
	sortRule := ctx.Query("sortRule")
	answer := &dto.PageQuery{
		Page:         page,
		PageSize:     pageSize,
		SortProperty: sortProperty,
		SortRule:     sortRule,
	}
	return answer, nil
}
