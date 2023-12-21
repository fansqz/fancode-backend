package admin

import (
	"FanCode/controller/admin"
	"github.com/gin-gonic/gin"
)

func SetupProblemCaseRoutes(r *gin.Engine, problemCaseController admin.ProblemCaseManagementController) {
	//题目相关路由
	bank := r.Group("/manage/problem/case")
	{
		bank.POST("", problemCaseController.InsertProblemCase)
		bank.PUT("", problemCaseController.UpdateProblemCase)
		bank.DELETE("/:id", problemCaseController.DeleteProblemCase)
		bank.GET("/list", problemCaseController.GetProblemCaseList)
		bank.GET("/:id", problemCaseController.GetProblemCaseByID)
	}
}
