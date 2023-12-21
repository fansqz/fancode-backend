package service

import (
	conf "FanCode/config"
	"FanCode/dao"
	e "FanCode/error"
	"FanCode/global"
	"FanCode/models/dto"
	"FanCode/models/po"
	"FanCode/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"os"
	"time"
)

type ProblemService interface {
	// CheckProblemNumber 检测题目编码
	CheckProblemNumber(problemCode string) (bool, *e.Error)
	// InsertProblem 添加题目
	InsertProblem(problem *po.Problem, ctx *gin.Context) (uint, *e.Error)
	// UpdateProblem 更新题目
	UpdateProblem(Problem *po.Problem, ctx *gin.Context) *e.Error
	// DeleteProblem 删除题目
	DeleteProblem(id uint) *e.Error
	// GetProblemList 获取题目列表
	GetProblemList(query *dto.PageQuery) (*dto.PageInfo, *e.Error)
	// GetUserProblemList 用户获取题目列表
	GetUserProblemList(ctx *gin.Context, query *dto.PageQuery) (*dto.PageInfo, *e.Error)
	// GetProblemByID 获取题目信息
	GetProblemByID(id uint) (*dto.ProblemDtoForGet, *e.Error)
	// GetProblemByNumber 根据题目编号获取题目信息
	GetProblemByNumber(number string) (*dto.ProblemDtoForGet, *e.Error)
	// GetProblemTemplateCode 获取题目的模板代码
	GetProblemTemplateCode(problemID uint, language string) (string, *e.Error)
	// UpdateProblemEnable 设置题目可用
	UpdateProblemEnable(id uint, enable int) *e.Error
}

type problemService struct {
	config            *conf.AppConfig
	problemDao        dao.ProblemDao
	problemCaseDao    dao.ProblemCaseDao
	problemAttemptDao dao.ProblemAttemptDao
}

func NewProblemService(config *conf.AppConfig, pd dao.ProblemDao, pcd dao.ProblemCaseDao, ad dao.ProblemAttemptDao) ProblemService {
	return &problemService{
		config:            config,
		problemDao:        pd,
		problemCaseDao:    pcd,
		problemAttemptDao: ad,
	}
}

func (q *problemService) CheckProblemNumber(problemCode string) (bool, *e.Error) {
	b, err := q.problemDao.CheckProblemNumberExists(global.Mysql, problemCode)
	if err != nil {
		return !b, e.ErrProblemCodeCheckFailed
	}
	return !b, nil
}

func (q *problemService) InsertProblem(problem *po.Problem, ctx *gin.Context) (uint, *e.Error) {
	problem.CreatorID = ctx.Keys["user"].(*dto.UserInfo).ID
	// 对设置值的数据设置默认值
	if problem.Name == "" {
		problem.Name = "未命名题目"
	}
	if problem.Title == "" {
		problem.Title = "标题信息"
	}
	if problem.Description == "" {
		problemDescription, err := os.ReadFile(q.config.FilePathConfig.ProblemDescriptionTemplate)
		if err != nil {
			return 0, e.ErrProblemInsertFailed
		}
		problem.Description = string(problemDescription)
	}
	if problem.Number == "" {
		problem.Number = "未命名编号" + utils.GetGenerateUniqueCode()
	}
	// 检测编号是否重复
	if problem.Number != "" {
		b, checkError := q.problemDao.CheckProblemNumberExists(global.Mysql, problem.Number)
		if checkError != nil {
			return 0, e.ErrMysql
		}
		if b {
			return 0, e.ErrProblemCodeIsExist
		}
	}
	// 题目难度不在范围，那么都设置为1
	if problem.Difficulty > 5 || problem.Difficulty < 1 {
		problem.Difficulty = 1
	}
	problem.Enable = -1
	// 添加
	err := q.problemDao.InsertProblem(global.Mysql, problem)
	if err != nil {
		return 0, e.ErrMysql
	}
	return problem.ID, nil
}

func (q *problemService) UpdateProblem(problem *po.Problem, ctx *gin.Context) *e.Error {
	problem.UpdatedAt = time.Now()
	// 更新题目
	if err := q.problemDao.UpdateProblem(global.Mysql, problem); err != nil {
		log.Println(err)
		return e.ErrProblemUpdateFailed
	}
	return nil
}

// todo: 这里有事务相关的问题
func (q *problemService) DeleteProblem(id uint) *e.Error {
	// 读取Problem
	problem, err := q.problemDao.GetProblemByID(global.Mysql, id)
	if err != nil {
		return e.ErrMysql
	}
	if problem == nil || problem.Number == "" {
		return e.ErrProblemNotExist
	}
	// 删除用例
	if err = q.problemCaseDao.DeleteProblemCaseByProblemID(global.Mysql, id); err != nil {
		return e.ErrMysql
	}
	// 删除题目
	if err = q.problemDao.DeleteProblemByID(global.Mysql, id); err != nil {
		return e.ErrMysql
	}
	return nil
}

func (q *problemService) GetProblemList(query *dto.PageQuery) (*dto.PageInfo, *e.Error) {
	var bankQuery *po.Problem
	if query.Query != nil {
		bankQuery = query.Query.(*po.Problem)
	}
	// 获取题目列表
	problems, err := q.problemDao.GetProblemList(global.Mysql, query)
	if err != nil {
		return nil, e.ErrMysql
	}
	newProblems := make([]*dto.ProblemDtoForList, len(problems))
	for i := 0; i < len(problems); i++ {
		newProblems[i] = dto.NewProblemDtoForList(problems[i])
	}
	// 获取所有题目总数目
	var count int64
	count, err = q.problemDao.GetProblemCount(global.Mysql, bankQuery)
	if err != nil {
		return nil, e.ErrMysql
	}
	pageInfo := &dto.PageInfo{
		Total: count,
		Size:  int64(len(newProblems)),
		List:  newProblems,
	}
	return pageInfo, nil
}

func (q *problemService) GetUserProblemList(ctx *gin.Context, query *dto.PageQuery) (*dto.PageInfo, *e.Error) {
	userId := ctx.Keys["user"].(*dto.UserInfo).ID
	if query.Query != nil {
		query.Query.(*po.Problem).Enable = 1
	} else {
		query.Query = &po.Problem{
			Enable: 1,
		}
	}
	// 获取题目列表
	problems, err := q.problemDao.GetProblemList(global.Mysql, query)
	if err != nil {
		return nil, e.ErrMysql
	}
	newProblems := make([]*dto.ProblemDtoForUserList, len(problems))
	for i := 0; i < len(problems); i++ {
		newProblems[i] = dto.NewProblemDtoForUserList(problems[i])
		// 读取题目完成情况
		var status int
		status, err = q.problemAttemptDao.GetProblemAttemptStatus(global.Mysql, userId, problems[i].ID)
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, e.ErrProblemListFailed
		}
		newProblems[i].Status = status
	}
	// 获取所有题目总数目
	var count int64
	count, err = q.problemDao.GetProblemCount(global.Mysql, query.Query.(*po.Problem))
	if err != nil {
		return nil, e.ErrMysql
	}
	pageInfo := &dto.PageInfo{
		Total: count,
		Size:  int64(len(newProblems)),
		List:  newProblems,
	}
	return pageInfo, nil
}

func (q *problemService) GetProblemByID(id uint) (*dto.ProblemDtoForGet, *e.Error) {
	problem, err := q.problemDao.GetProblemByID(global.Mysql, id)
	if err == gorm.ErrRecordNotFound {
		return nil, e.ErrProblemNotExist
	}
	if err != nil {
		return nil, e.ErrMysql
	}
	return dto.NewProblemDtoForGet(problem), nil
}

func (q *problemService) GetProblemByNumber(number string) (*dto.ProblemDtoForGet, *e.Error) {
	problem, err := q.problemDao.GetProblemByNumber(global.Mysql, number)
	if err != nil {
		return nil, e.ErrMysql
	}
	return dto.NewProblemDtoForGet(problem), nil
}

func (q *problemService) GetProblemTemplateCode(problemID uint, language string) (string, *e.Error) {
	// 读取acm模板
	code, err := getAcmCodeTemplate(language)
	if err != nil {
		return "", e.ErrProblemGetFailed
	}
	return code, nil
}

// todo: 是否要加事务
func (q *problemService) UpdateProblemEnable(id uint, enable int) *e.Error {
	if err := q.problemDao.SetProblemEnable(global.Mysql, id, enable); err != nil {
		return e.ErrMysql
	}
	return nil
}
