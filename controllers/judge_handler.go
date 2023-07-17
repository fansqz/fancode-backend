package controllers

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
	Execute(ctx *gin.Context)
	Submit(ctx *gin.Context)
}

type judgeController struct {
	judgeService service.JudgeService
}

func NewJudgeController() JudgeController {
	return &judgeController{
		judgeService: service.NewJudgeService(),
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
	judgeRequest := &dto.JudgingRequestDTO{
		Code:      ctx.PostForm("code"),
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

}
