package service

import (
	"FanCode/dao"
	e "FanCode/error"
	"FanCode/global"
	"FanCode/models/dto"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

type SubmissionService interface {
	// GetActivityMap 获取活动图
	GetActivityMap(ctx *gin.Context, year int) ([]*dto.ActivityItem, *e.Error)
	// GetActivityYear 获取用户有活动的年份
	GetActivityYear(ctx *gin.Context) ([]string, *e.Error)
}

func NewSubmissionService() SubmissionService {
	return &submissionService{}
}

type submissionService struct {
}

func (u *submissionService) GetActivityMap(ctx *gin.Context, year int) ([]*dto.ActivityItem, *e.Error) {
	user := ctx.Keys["user"].(*dto.UserInfo)
	var startDate time.Time
	var endDate time.Time
	// 如果year == 0，获取以今天截至的一年的数据
	if year == 0 {
		endDate = time.Now()
		startDate = time.Date(endDate.Year()-1, endDate.Month()+1, endDate.Day(),
			0, 0, 0, 0, time.Local)
	} else {
		startDate, endDate = getYearRange(year)
	}
	submissions, err := dao.GetUserSimpleSubmissionsByTime(global.Mysql, user.ID, startDate, endDate)
	if err != nil {
		return nil, e.ErrMysql
	}
	// 构建活动数据
	m := make(map[string]int, 366)
	for i := 0; i < len(submissions); i++ {
		date := submissions[i].CreatedAt.Format("2006-01-02")
		m[date]++
	}
	answer := make([]*dto.ActivityItem, len(m))
	i := 0
	for k, v := range m {
		answer[i] = &dto.ActivityItem{
			Date:  k,
			Count: v,
		}
		i++
	}
	return answer, nil
}

func (u *submissionService) GetActivityYear(ctx *gin.Context) ([]string, *e.Error) {
	answer := []string{}
	user := ctx.Keys["user"].(*dto.UserInfo)
	beginYear := 2022
	currentYear := time.Now().Year()
	for i := beginYear; i <= currentYear; i++ {
		beginDate, endDate := getYearRange(i)
		b, err := dao.CheckUserIsSubmittedByTime(global.Mysql, user.ID, beginDate, endDate)
		if err != nil {
			return nil, e.ErrMysql
		}
		if b {
			answer = append(answer, strconv.Itoa(i))
		}
	}
	return answer, nil
}

func getYearRange(year int) (time.Time, time.Time) {
	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	endDate := time.Date(year, 12, 31, 23, 59, 59, 999999999, time.Local)
	return startDate, endDate
}
