package routers

import (
	"FanCode/controllers"
	"github.com/gin-gonic/gin"
)

func SetupProblemRoutes(r *gin.Engine) {
	//题目相关路由
	problem := r.Group("/problem")
	{
		problemController := controllers.NewProblemController()
		problem.GET("/code/check/:code", problemController.CheckProblemCode)
		problem.POST("/insert", problemController.InsertProblem)
		problem.PUT("/update", problemController.UpdateProblem)
		problem.DELETE("/delete/:id", problemController.DeleteProblem)
		problem.GET("/list/:page/:pageSize", problemController.GetProblemList)
		problem.GET("/get/:id", problemController.GetProblemByID)
		problem.GET("/file/download/:id", problemController.DownloadProblemFile)
		problem.GET("/file/download/template", problemController.DownloadProblemTemplateFile)
		problem.POST("/enable", problemController.UpdateProblemEnable)
	}
}
