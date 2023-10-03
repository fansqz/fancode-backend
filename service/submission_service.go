package service

import (
	"FanCode/dao"
	e "FanCode/error"
	"FanCode/global"
	"FanCode/models/dto"
	"FanCode/models/po"
	"github.com/gin-gonic/gin"
	"log"
	"strconv"
	"time"
)

type SubmissionService interface {
	// GetActivityMap 获取活动图
	GetActivityMap(ctx *gin.Context, year int) ([]*dto.ActivityItem, *e.Error)
	// GetActivityYear 获取用户有活动的年份
	GetActivityYear(ctx *gin.Context) ([]string, *e.Error)
	// GetUserSubmissionList 获取用户
	GetUserSubmissionList(ctx *gin.Context, pageQuery *dto.PageQuery) (*dto.PageInfo, *e.Error)
}

func NewSubmissionService(submissionDao dao.SubmissionDao, problemDao dao.ProblemDao) SubmissionService {
	return &submissionService{
		submissionDao: submissionDao,
		problemDao:    problemDao,
	}
}

type submissionService struct {
	submissionDao dao.SubmissionDao
	problemDao    dao.ProblemDao
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
	submissions, err := u.submissionDao.GetUserSimpleSubmissionsByTime(global.Mysql, user.ID, startDate, endDate)
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
		b, err := u.submissionDao.CheckUserIsSubmittedByTime(global.Mysql, user.ID, beginDate, endDate)
		if err != nil {
			return nil, e.ErrMysql
		}
		if b {
			answer = append(answer, strconv.Itoa(i))
		}
	}
	return answer, nil
}

func (u *submissionService) GetUserSubmissionList(ctx *gin.Context, pageQuery *dto.PageQuery) (*dto.PageInfo, *e.Error) {
	user := ctx.Keys["user"].(*dto.UserInfo)
	submission := &po.Submission{
		UserID: user.ID,
	}
	pageQuery.Query = submission
	submissions, err := u.submissionDao.GetSubmissionList(global.Mysql, pageQuery)
	submissions2 := make([]*dto.SubmissionDtoForList, len(submissions))
	for i := 0; i < len(submissions); i++ {
		submissions2[i] = dto.NewSubmissionDtoForList(submissions[i])
		name, err := u.problemDao.GetProblemNameByID(global.Mysql, submissions[i].ProblemID)
		if err != nil {
			return nil, e.ErrMysql
		}
		submissions2[i].ProblemName = name
	}
	if err != nil {
		log.Println(err)
		return nil, e.ErrMysql
	}
	count, err2 := u.submissionDao.GetSubmissionCount(global.Mysql, submission)
	if err2 != nil {
		log.Println(err)
		return nil, e.ErrMysql
	}
	return &dto.PageInfo{
		Total: count,
		Size:  int64(len(submissions2)),
		List:  submissions2,
	}, nil
}

func getYearRange(year int) (time.Time, time.Time) {
	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	endDate := time.Date(year, 12, 31, 23, 59, 59, 999999999, time.Local)
	return startDate, endDate
}
