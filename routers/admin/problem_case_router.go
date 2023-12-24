package admin

import (
	"FanCode/controller/admin"
	"github.com/gin-gonic/gin"
)

func SetupProblemCaseRoutes(r *gin.Engine, problemCaseController admin.ProblemCaseManagementController) {
	//题目相关路由
	pcase := r.Group("/manage/problem/case")
	{
		pcase.POST("", problemCaseController.InsertProblemCase)
		pcase.PUT("", problemCaseController.UpdateProblemCase)
		pcase.DELETE("/:id", problemCaseController.DeleteProblemCase)
		pcase.GET("/list", problemCaseController.GetProblemCaseList)
		pcase.GET("/:id", problemCaseController.GetProblemCaseByID)
		pcase.GET("/name/new", problemCaseController.GenerateNewProblemCaseName)
		pcase.GET("/name/check", problemCaseController.CheckProblemCaseName)
	}
}
