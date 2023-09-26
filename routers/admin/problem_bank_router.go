package admin

import (
	"FanCode/controllers/admin"
	"github.com/gin-gonic/gin"
)

func SetupProblemBankRoutes(r *gin.Engine) {
	//题目相关路由
	bank := r.Group("/manage/problemBank")
	{
		problemBankController := admin.NewProblemBankManagementController()
		bank.POST("", problemBankController.InsertProblemBank)
		bank.PUT("", problemBankController.UpdateProblemBank)
		bank.DELETE("/:id/:forceDelete", problemBankController.DeleteProblemBank)
		bank.GET("/list", problemBankController.GetProblemBankList)
		bank.GET("/:id", problemBankController.GetProblemBankByID)
	}
}
