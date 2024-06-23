package user

import (
	"FanCode/constants"
	"FanCode/controller/utils"
	"FanCode/models/dto"
	r "FanCode/models/vo"
	"FanCode/service"
	"github.com/gin-gonic/gin"
)

// JudgeController
// @Description: 判题模块
type JudgeController interface {
	// Execute 执行
	Execute(ctx *gin.Context)
	// Submit 提交
	Submit(ctx *gin.Context)
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
	judgeRequest := &dto.ExecuteRequestDto{
		Code:      ctx.PostForm("code"),
		Input:     ctx.PostForm("input"),
		Language:  constants.LanguageType(ctx.PostForm("language")),
		ProblemID: uint(utils.AtoiOrDefault(ctx.PostForm("problemID"), 0)),
	}
	// 读取题目id
	response, err := j.judgeService.Execute(judgeRequest)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(response)
}

func (j *judgeController) Submit(ctx *gin.Context) {
	result := r.NewResult(ctx)
	problemID := utils.AtoiOrDefault(ctx.PostForm("problemID"), 0)
	judgeRequest := &dto.SubmitRequestDto{
		Code:      ctx.PostForm("code"),
		Language:  constants.LanguageType(ctx.PostForm("language")),
		ProblemID: uint(problemID),
	}
	// 读取题目id
	response, err := j.judgeService.Submit(ctx, judgeRequest)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(response)
}
