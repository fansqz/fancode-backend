package service

import (
	"FanCode/dao/mock"
	"FanCode/global"
	"FanCode/models/dto"
	"FanCode/models/po"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
	"time"
)

func TestProblemBankService_InsertProblemBank(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	problemBankDao := mock.NewMockProblemBankDao(mockCtl)
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

func TestProblemBankService_UpdateProblemBank(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	problemBankDao := mock.NewMockProblemBankDao(mockCtl)
	// mock数据
	bank := &po.ProblemBank{
		Name:        "name",
		Icon:        "icon",
		Description: "description",
	}
	problemBankDao.EXPECT().UpdateProblemBank(global.Mysql, gomock.Any()).Return(nil).Do(
		func(db *gorm.DB, bank *po.ProblemBank) {
			assert.NotNil(t, bank.UpdatedAt)
			bank.UpdatedAt = time.Time{}
			bank2 := &po.ProblemBank{
				Name:        "name",
				Icon:        "icon",
				Description: "description",
			}
			assert.Equal(t, bank, bank2)
		})

	bankService := NewProblemBankService(problemBankDao, nil, nil)
	err := bankService.UpdateProblemBank(bank)
	assert.Nil(t, err)
}

func TestProblemBankService_DeleteProblemBank(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	problemDao := mock.NewMockProblemDao(mockCtl)
	problemBankDao := mock.NewMockProblemBankDao(mockCtl)

	problemDao.EXPECT().GetProblemCount(global.Mysql, gomock.Any()).DoAndReturn(
		func(db *gorm.DB, problem *po.Problem) (int64, error) {
			assert.NotEqual(t, problem.BankID, nil)
			if *problem.BankID == 1 {
				return 1, nil
			} else {
				return 0, nil
			}
		}).MaxTimes(4)

	problemBankDao.EXPECT().DeleteProblemBankByID(global.Mysql, gomock.Any()).Return(nil).MaxTimes(4)

	bankService := NewProblemBankService(problemBankDao, problemDao, nil)
	err := bankService.DeleteProblemBank(1, false)
	assert.NotNil(t, err)
	assert.Equal(t, err.Message, "题库不为空，请问是否需要强制删除")
	err = bankService.DeleteProblemBank(1, true)
	assert.Nil(t, err)
	err = bankService.DeleteProblemBank(2, false)
	assert.Nil(t, err)
	err = bankService.DeleteProblemBank(2, true)
	assert.Nil(t, err)
}
