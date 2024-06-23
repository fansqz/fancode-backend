package user

import (
	"FanCode/controller/user"
	"github.com/gin-gonic/gin"
)

func SetupProblemRoutes(r *gin.Engine, problemController user.ProblemController) {
	//题目相关路由
	problem := r.Group("/problem")
	{
		problem.GET("/list", problemController.GetProblemList)
		problem.GET("/:number", problemController.GetProblem)
		problem.GET("/code/template/:problemID/:language", problemController.GetProblemTemplateCode)
		problem.GET("/code/:problemID/:language", problemController.GetUserCode)
		problem.GET("/code/:problemID", problemController.GetUserCodeByProblemID)
		problem.POST("/code/save", problemController.SaveUserCode)
	}
}
