package service

import (
	e "FanCode/error"
	"FanCode/models/dto"
	"FanCode/models/po"
	"github.com/gin-gonic/gin"
)

// ProblemBankService 题库管理的service
type ProblemBankService interface {
	// InsertProblemBank 添加题库
	InsertProblemBank(problemBank *po.ProblemBank) (uint, *e.Error)
	// UpdateProblemBank 更新题库
	UpdateProblemBank(problemBank *po.ProblemBank, ctx *gin.Context) *e.Error
	// DeleteProblemBank 删除题库
	DeleteProblemBank(id uint) *e.Error
	// GetProblemBankList 获取题目列表
	GetProblemBankList(query *dto.PageQuery) (*dto.PageInfo, *e.Error)
	// GetProblemBankByID 获取题目信息
	GetProblemBankByID(id uint) (*dto.ProblemDtoForGet, *e.Error)
}
