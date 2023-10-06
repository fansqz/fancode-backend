package user

import (
	"FanCode/controller/user"
	"github.com/gin-gonic/gin"
)

func SetupProblemBankRoutes(r *gin.Engine, problemBankController user.ProblemBankController) {
	//题目相关路由
	problem := r.Group("/problemBank")
	{
		problem.GET("/list", problemBankController.GetProblemBankList)
	}
}
