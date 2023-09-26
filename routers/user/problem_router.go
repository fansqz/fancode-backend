package user

import (
	"FanCode/controllers/user"
	"github.com/gin-gonic/gin"
)

func SetupProblemRoutes(r *gin.Engine) {
	//题目相关路由
	problem := r.Group("/problem")
	{
		problemController := user.NewProblemController()
		problem.GET("/list", problemController.GetProblemList)
		problem.GET("/:number", problemController.GetProblem)
		problem.GET("/code/:number/:language/:codeType", problemController.GetProblemTemplateCode)
	}
}
