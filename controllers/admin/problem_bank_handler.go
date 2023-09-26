package admin

import (
	"FanCode/service"
	"github.com/gin-gonic/gin"
)

// ProblemBankManagementController
// @Description: 题库管理相关功能
type ProblemBankManagementController interface {
	// InsertProblemBank 添加题库
	InsertProblemBank(ctx *gin.Context)
	// UpdateProblemBank 更新题库
	UpdateProblemBank(ctx *gin.Context)
	// DeleteProblemBank 删除题库
	DeleteProblemBank(ctx *gin.Context)
	// GetProblemBankList 读取题库列表
	GetProblemBankList(ctx *gin.Context)
	// GetProblemBankByID 读取题库信息
	GetProblemBankByID(ctx *gin.Context)
}

type problemBankManagementController struct {
	service.SysApiService
}
