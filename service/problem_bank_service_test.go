package service

import (
	"FanCode/dao"
	"FanCode/global"
	"FanCode/models/dto"
	"FanCode/models/po"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
)

func TestProblemBankService_InsertProblemBank(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	problemBankDao := dao.NewMockProblemBankDao(mockCtl)
	// mock数据
	bank := &po.ProblemBank{
		Name:        "",
		Icon:        "icon",
		Description: "",
	}
	resultID := uint(10)
	problemBankDao.EXPECT().InsertProblemBank(global.Mysql, gomock.Any()).Return(nil).Do(
		func(db *gorm.DB, bank *po.ProblemBank) {
			bank2 := &po.ProblemBank{
				Name:        "未命名题库",
				Icon:        "icon",
				Description: "无描述信息",
				CreatorID:   uint(1),
			}
			assert.Equal(t, bank, bank2)
			bank.ID = resultID
		})

	ctx := &gin.Context{}
	ctx.Keys = make(map[string]interface{})
	ctx.Keys["user"] = &dto.UserInfo{
		ID: 1,
	}
	bankService := NewProblemBankService(problemBankDao, nil, nil)
	id, err := bankService.InsertProblemBank(bank, ctx)
	assert.Equal(t, id, resultID)
	assert.Nil(t, err)
}
