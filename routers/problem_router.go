package routers

import (
	"FanCode/controllers"
	"github.com/gin-gonic/gin"
)

func SetupProblemRoutes(r *gin.Engine) {
	//用户相关
	problem := r.Group("/problem")
	{
		problemController := controllers.NewProblemController()
		problem.POST("/insert", problemController.InsertProblem)
		problem.PUT("/update", problemController.UpdateProblem)
		problem.DELETE("/delete/:id", problemController.DeleteProblem)
		problem.GET("/list/:page/:pageSize", problemController.GetProblemList)
		problem.POST("/upload/file", problemController.UploadProblemFile)
	}
}
