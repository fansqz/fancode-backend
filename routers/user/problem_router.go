package user

import (
	"FanCode/controllers/user"
	"github.com/gin-gonic/gin"
)

func SetupProblemRoutes(r *gin.Engine) {
	//题目相关路由
	problem := r.Group("/user/problem")
	{
		problemController := user.NewProblemController()
		problem.GET("/list/:page/:pageSize", problemController.GetProblemList)
		problem.GET("/:number", problemController.GetProblemByNumber)
		problem.POST("/code/:number", problemController.GetProblemCodeByNumber)
	}
}
