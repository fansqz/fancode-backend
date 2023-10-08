package user

import (
	e "FanCode/error"
	"FanCode/models/dto"
	r "FanCode/models/vo"
	"FanCode/service"
	"github.com/gin-gonic/gin"
	"strconv"
)

// JudgeController
// @Description: 判题模块
type JudgeController interface {
	// Execute 执行
	Execute(ctx *gin.Context)
	// Submit 提交
	Submit(ctx *gin.Context)
	// 保存代码
	SaveCode(ctx *gin.Context)
}

type judgeController struct {
	judgeService service.JudgeService
}

func NewJudgeController(judgeService service.JudgeService) JudgeController {
	return &judgeController{
		judgeService: judgeService,
	}
}

func (j *judgeController) Execute(ctx *gin.Context) {
	result := r.NewResult(ctx)
	problemIDStr := ctx.PostForm("problemID")
	problemID, err := strconv.Atoi(problemIDStr)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	judgeRequest := &dto.ExecuteRequestDto{
		Code:      ctx.PostForm("code"),
		Input:     ctx.PostForm("input"),
		CodeType:  ctx.PostForm("codeType"),
		Language:  ctx.PostForm("language"),
		ProblemID: uint(problemID),
	}
	// 读取题目id
	response, err2 := j.judgeService.Execute(judgeRequest)
	if err2 != nil {
		result.Error(err2)
	} else {
		result.SuccessData(response)
	}
}

func (j *judgeController) Submit(ctx *gin.Context) {
	result := r.NewResult(ctx)
	problemIDStr := ctx.PostForm("problemID")
	problemID, err := strconv.Atoi(problemIDStr)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	judgeRequest := &dto.SubmitRequestDto{
		Code:      ctx.PostForm("code"),
		CodeType:  ctx.PostForm("codeType"),
		Language:  ctx.PostForm("language"),
		ProblemID: uint(problemID),
	}
	// 读取题目id
	response, err2 := j.judgeService.Submit(ctx, judgeRequest)
	if err2 != nil {
		result.Error(err2)
	} else {
		result.SuccessData(response)
	}
}

func (j *judgeController) SaveCode(ctx *gin.Context) {
	result := r.NewResult(ctx)
	// 题库id
	problemIDStr := ctx.PostForm("problemID")
	code := ctx.PostForm("code")
	problemID, err := strconv.Atoi(problemIDStr)
	if err != nil {
		result.Error(e.ErrBadRequest)
		return
	}
	err2 := j.judgeService.SaveCode(ctx, uint(problemID), code)
	if err2 != nil {
		result.Error(err2)
		return
	}
	result.SuccessMessage("保存成功")
}
