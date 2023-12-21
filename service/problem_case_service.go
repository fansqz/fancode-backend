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
		return e.ErrMysql
	}
	return nil
}

func (p *problemCaseService) InsertProblemCase(problemCase *po.ProblemCase) (uint, *e.Error) {
	err := p.problemCaseDao.InsertProblemCase(global.Mysql, problemCase)
	if err != nil {
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
		return e.ErrMysql
	}
	return nil
}
