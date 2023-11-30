package user

import (
	e "FanCode/error"
	"FanCode/models/dto"
	r "FanCode/models/vo"
	"FanCode/service"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

type SubmissionController interface {
	// GetUserActivityMap 获取用户活动图
	GetUserActivityMap(ctx *gin.Context)
	// GetUserActivityYear 获取用户有活动的年份
	GetUserActivityYear(ctx *gin.Context)
	// GetUserSubmissionList 获取用户提交列表
	GetUserSubmissionList(ctx *gin.Context)
}

func NewSubmissionController(submissionService service.SubmissionService) SubmissionController {
	return &submissionController{
		submissionService: submissionService,
	}
}

type submissionController struct {
	submissionService service.SubmissionService
}

func (a *submissionController) GetUserActivityMap(ctx *gin.Context) {
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

func (a *submissionController) GetUserActivityYear(ctx *gin.Context) {
	result := r.NewResult(ctx)
	years, err := a.submissionService.GetActivityYear(ctx)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(years)
}

func (a *submissionController) GetUserSubmissionList(ctx *gin.Context) {
	result := r.NewResult(ctx)
	pageQuery, err := GetPageQueryByQuery(ctx)
	if err != nil {
		result.Error(err)
		return
	}
	pageInfo, err := a.submissionService.GetUserSubmissionList(ctx, pageQuery)
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
