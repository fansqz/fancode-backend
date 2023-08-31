package user

import (
	e "FanCode/error"
	r "FanCode/models/vo"
	"FanCode/service"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

type SubmissionHandler interface {
	// GetUserActivityMap 获取用户活动图
	GetUserActivityMap(ctx *gin.Context)
	// GetUserActivityYear 获取用户有活动的年份
	GetUserActivityYear(ctx *gin.Context)
	// GetUserSubmissionList 获取用户提交列表
	GetUserSubmissionList(ctx *gin.Context)
}

func NewSubmissionHandler() SubmissionHandler {
	return &submissionHandler{
		submissionService: service.NewSubmissionService(),
	}
}

type submissionHandler struct {
	submissionService service.SubmissionService
}

func (a *submissionHandler) GetUserActivityMap(ctx *gin.Context) {
	result := r.NewResult(ctx)
	yearStr := ctx.Param("year")
	// 检测年份是否合理
	var year int
	if yearStr == "0" {
		year = 0
	} else {
		var b bool
		year, b = checkYear(yearStr)
		if !b {
			result.Error(e.ErrBadRequest)
			return
		}
	}
	activityMap, err := a.submissionService.GetActivityMap(ctx, year)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(activityMap)
}

func (a *submissionHandler) GetUserActivityYear(ctx *gin.Context) {
	result := r.NewResult(ctx)
	years, err := a.submissionService.GetActivityYear(ctx)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(years)
}

func (a *submissionHandler) GetUserSubmissionList(ctx *gin.Context) {
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
	pageInfo, err := a.submissionService.GetUserSubmissionList(ctx, page, pageSize)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(pageInfo)
}

func checkYear(str string) (int, bool) {
	year, err := strconv.Atoi(str)
	if err != nil {
		return 0, false
	}

	currentYear := time.Now().Year()
	b := year > 2022 && year <= currentYear
	return year, b
}
