package service

import (
	"FanCode/dao/mock"
	"FanCode/global"
	"FanCode/models/dto"
	"FanCode/models/po"
	"FanCode/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"strconv"
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

func TestProblemBankService_GetProblemBankList(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	bankDao := mock.NewMockProblemBankDao(mockCtl)
	problemDao := mock.NewMockProblemDao(mockCtl)
	userDao := mock.NewMockSysUserDao(mockCtl)
	bankService := NewProblemBankService(bankDao, problemDao, userDao)

	// 测试1
	testBankList := []*po.ProblemBank{
		{
			Name: "bank1", Icon: "icon1", Description: "description1", CreatorID: 1,
			Model: gorm.Model{ID: 1, UpdatedAt: time.Now(), CreatedAt: time.Now()},
		},
		{
			Name: "bank2", Icon: "icon2", Description: "description2", CreatorID: 2,
			Model: gorm.Model{ID: 2, UpdatedAt: time.Now(), CreatedAt: time.Now()},
		},
		{
			Name: "bank3", Icon: "icon3", Description: "description3", CreatorID: 3,
			Model: gorm.Model{ID: 3, UpdatedAt: time.Now(), CreatedAt: time.Now()},
		},
	}
	bankDao.EXPECT().GetProblemBankList(gomock.Any(), &dto.PageQuery{
		Page: 1, PageSize: 3, SortProperty: "name", SortRule: "desc", Query: &po.ProblemBank{Name: "bank"},
	}).Return(testBankList, nil)
	bankDao.EXPECT().GetProblemBankCount(gomock.Any(), &po.ProblemBank{Name: "bank"}).Return(int64(10), nil)
	problemDao.EXPECT().GetProblemCount(gomock.Any(), gomock.Any()).
		DoAndReturn(func(db *gorm.DB, problem *po.Problem) (int64, error) {
			return int64(*problem.BankID), nil
		}).Times(3)
	userDao.EXPECT().GetUserNameByID(gomock.Any(), gomock.Any()).
		DoAndReturn(func(db *gorm.DB, id uint) (string, error) {
			return "user" + strconv.Itoa(int(id)), nil
		}).Times(3)
	pageInfo, err2 := bankService.GetProblemBankList(&dto.PageQuery{
		Page: 1, PageSize: 3, SortProperty: "name", SortRule: "desc", Query: &po.ProblemBank{Name: "bank"},
	})
	assert.Nil(t, err2)
	assert.Equal(t, &dto.PageInfo{
		Size:  3,
		Total: 10,
		List: []*dto.ProblemBankDtoForList{
			{
				ID: 1, Name: "bank1", Icon: "icon1", Description: "description1",
				CreatorName: "user1", ProblemCount: 1,
				CreatedAt: utils.Time(testBankList[0].CreatedAt), UpdatedAt: utils.Time(testBankList[0].UpdatedAt),
			},
			{
				ID: 2, Name: "bank2", Icon: "icon2", Description: "description2",
				CreatorName: "user2", ProblemCount: 2,
				CreatedAt: utils.Time(testBankList[1].CreatedAt), UpdatedAt: utils.Time(testBankList[1].UpdatedAt),
			},
			{
				ID: 3, Name: "bank3", Icon: "icon3", Description: "description3",
				CreatorName: "user3", ProblemCount: 3,
				CreatedAt: utils.Time(testBankList[2].CreatedAt), UpdatedAt: utils.Time(testBankList[2].UpdatedAt),
			},
		},
	}, pageInfo)
}
