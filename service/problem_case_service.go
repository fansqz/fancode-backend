package service

import (
	conf "FanCode/config"
	"FanCode/dao"
	e "FanCode/error"
	"FanCode/global"
	"FanCode/models/dto"
	"FanCode/models/po"
	"errors"
	"gorm.io/gorm"
	"log"
	"strconv"
	"unicode"
)

// ProblemCaseService
// 题目用例管理
type ProblemCaseService interface {
	// GetProblemCaseList 获取用例列表
	GetProblemCaseList(query *dto.PageQuery) (*dto.PageInfo, *e.Error)
	// GetProblemCaseByID 通过id获取题目用例
	GetProblemCaseByID(id uint) (*dto.ProblemCaseDtoForGet, *e.Error)
	// DeleteProblemCaseByID 通过id删除题目用例
	DeleteProblemCaseByID(id uint) *e.Error
	// InsertProblemCase 添加题目用例
	InsertProblemCase(problemCase *po.ProblemCase) (uint, *e.Error)
	// UpdateProblemCase 更新题目用例
	UpdateProblemCase(problemCase *po.ProblemCase) *e.Error
	// CheckProblemCaseName 检测用例名称是否重复
	CheckProblemCaseName(name string, problemID uint) (bool, *e.Error)
	// GenerateNewProblemCaseName 生成一个题目唯一用例名称，递增
	GenerateNewProblemCaseName(problemID uint) (string, *e.Error)
}

type problemCaseService struct {
	config         *conf.AppConfig
	problemCaseDao dao.ProblemCaseDao
	problemDao     dao.ProblemDao
}

func NewProblemCaseService(config *conf.AppConfig, pcd dao.ProblemCaseDao, pd dao.ProblemDao) ProblemCaseService {
	return &problemCaseService{
		config:         config,
		problemCaseDao: pcd,
		problemDao:     pd,
	}
}

func (p *problemCaseService) GetProblemCaseList(query *dto.PageQuery) (*dto.PageInfo, *e.Error) {
	var problemCase *po.ProblemCase
	if query.Query != nil {
		problemCase = query.Query.(*po.ProblemCase)
	}
	// 获取用例列表
	cases, err := p.problemCaseDao.GetProblemCaseList(global.Mysql, query)
	if err != nil {
		log.Println("Error while getting problem case list:", err)
		return nil, e.ErrMysql
	}
	newCases := make([]*dto.ProblemCaseDtoForList, len(cases))
	for i := 0; i < len(cases); i++ {
		newCases[i] = dto.NewProblemCaseDtoForList(cases[i])
	}
	// 获取所有用例总数目
	var count int64
	count, err = p.problemCaseDao.GetProblemCaseCount(global.Mysql, problemCase)
	if err != nil {
		log.Println("Error while getting problem case list:", err)
		return nil, e.ErrMysql
	}
	pageInfo := &dto.PageInfo{
		Total: count,
		Size:  int64(len(newCases)),
		List:  newCases,
	}
	return pageInfo, nil
}

func (p *problemCaseService) GetProblemCaseByID(id uint) (*dto.ProblemCaseDtoForGet, *e.Error) {
	problemCase, err := p.problemCaseDao.GetProblemCaseByID(global.Mysql, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, e.ErrProblemNotExist
	}
	if err != nil {
		log.Println("Error while getting problem case name:", err)
		return nil, e.ErrMysql
	}
	return dto.NewProblemCaseDtoForGet(problemCase), nil
}

func (p *problemCaseService) DeleteProblemCaseByID(id uint) *e.Error {
	err := p.problemCaseDao.DeleteProblemCaseByID(global.Mysql, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return e.ErrProblemNotExist
	}
	if err != nil {
		log.Println("Error while deleting problem case:", err)
		return e.ErrMysql
	}
	return nil
}

func (p *problemCaseService) InsertProblemCase(problemCase *po.ProblemCase) (uint, *e.Error) {
	err := p.problemCaseDao.InsertProblemCase(global.Mysql, problemCase)
	if err != nil {
		log.Println("Error while inserting problem case:", err)
		return 0, e.ErrMysql
	}
	return problemCase.ID, nil
}

func (p *problemCaseService) UpdateProblemCase(problemCase *po.ProblemCase) *e.Error {
	err := p.problemCaseDao.UpdateProblemCase(global.Mysql, problemCase)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return e.ErrProblemNotExist
	}
	if err != nil {
		log.Println("Error while updating problem case:", err)
		return e.ErrMysql
	}
	return nil
}

func (p *problemCaseService) CheckProblemCaseName(name string, problemID uint) (bool, *e.Error) {
	l, err := p.problemCaseDao.GetProblemCaseList(global.Mysql, &dto.PageQuery{
		Page:     1,
		PageSize: 1,
		Query: &po.ProblemCase{
			Name:      name,
			ProblemID: problemID,
		},
	})
	if err != nil {
		log.Println("Error while checking problem case name:", err)
		return false, e.ErrMysql
	}
	return len(l) == 0, nil
}

func (p *problemCaseService) GenerateNewProblemCaseName(problemID uint) (string, *e.Error) {
	// 获取与给定问题ID相关的问题用例列表
	problemCases, err := p.problemCaseDao.GetProblemCaseList(global.Mysql, &dto.PageQuery{
		Page:         1,
		PageSize:     1,
		SortProperty: "name",
		SortRule:     "desc",
		Query: &po.ProblemCase{
			ProblemID: problemID,
		},
	})
	if err != nil {
		// 记录错误并返回
		log.Println("Error while getting problem case list:", err)
		return "", e.ErrMysql
	}

	// 如果没有找到问题用例，直接返回 "1"
	if len(problemCases) == 0 {
		return "1", nil
	}

	// 获取最新的问题用例
	latestCase := problemCases[0]

	// 寻找最后一个非数字字符的索引
	i := len(latestCase.Name) - 1
	for i >= 0 && unicode.IsDigit(rune(latestCase.Name[i])) {
		i--
	}

	// 截取数字部分并转换为整数
	numericPart := latestCase.Name[i+1:]
	num, err := strconv.Atoi(numericPart)
	if err != nil {
		// 记录错误并返回
		log.Println("Error converting numeric part to integer:", err)
		return "", e.ErrUnknown
	}

	// 递增数字部分并生成新名称
	num++
	newName := latestCase.Name[:i+1] + strconv.Itoa(num)
	return newName, nil
}
