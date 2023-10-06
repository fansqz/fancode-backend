package user

import (
	r "FanCode/models/vo"
	"FanCode/service"
	"github.com/gin-gonic/gin"
)

type ProblemBankController interface {
	// GetProblemBankList 读取题库列表
	GetProblemBankList(ctx *gin.Context)
}
type problemBankController struct {
	problemBankService service.ProblemBankService
}

func NewProblemBankController(bankService service.ProblemBankService) ProblemBankController {
	return &problemBankController{
		problemBankService: bankService,
	}
}

func (p *problemBankController) GetProblemBankList(ctx *gin.Context) {
	result := r.NewResult(ctx)
	pageQuery, err := GetPageQueryByQuery(ctx)
	if err != nil {
		result.Error(err)
		return
	}
	pageInfo, err := p.problemBankService.GetProblemBankList(pageQuery)
	if err != nil {
		result.Error(err)
		return
	}
	result.SuccessData(pageInfo)
}
