package admin

import (
	"FanCode/controllers/admin"
	"github.com/gin-gonic/gin"
)

func SetupProblemRoutes(r *gin.Engine) {
	//题目相关路由
	problem := r.Group("/manage/problem")
	{
		problemController := admin.NewProblemManagementController()
		problem.GET("/code/check/:number", problemController.CheckProblemNumber)
		problem.POST("", problemController.InsertProblem)
		problem.PUT("", problemController.UpdateProblem)
		problem.DELETE("/:id", problemController.DeleteProblem)
		problem.GET("/list", problemController.GetProblemList)
		problem.GET("/:id", problemController.GetProblemByID)
		problem.GET("/file/download/:id", problemController.DownloadProblemFile)
		problem.GET("/file/download/template", problemController.DownloadProblemTemplateFile)
		problem.POST("/enable", problemController.UpdateProblemEnable)
	}
}
