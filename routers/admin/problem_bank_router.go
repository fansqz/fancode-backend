package admin

import (
	"FanCode/controller/admin"
	"github.com/gin-gonic/gin"
)

func SetupProblemBankRoutes(r *gin.Engine, problemBankController admin.ProblemBankManagementController) {
	//题目相关路由
	bank := r.Group("/manage/problemBank")
	{
		bank.POST("", problemBankController.InsertProblemBank)
		bank.PUT("", problemBankController.UpdateProblemBank)
		bank.DELETE("/:id/:forceDelete", problemBankController.DeleteProblemBank)
		bank.GET("/list", problemBankController.GetProblemBankList)
		bank.GET("/simple/list", problemBankController.GetSimpleProblemBankList)
		bank.GET("/:id", problemBankController.GetProblemBankByID)
		bank.POST("/icon", problemBankController.UploadProblemBankIcon)
		bank.GET("/icon/:iconName", problemBankController.ReadProblemBankIcon)
	}
}
