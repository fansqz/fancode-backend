package user

import (
	"FanCode/controller/user"
	"github.com/gin-gonic/gin"
)

func SetupJudgeRoutes(r *gin.Engine, judgeController user.JudgeController) {
	//用户相关
	judge := r.Group("/judge")
	{
		judge.POST("/submit", judgeController.Submit)
		judge.POST("/execute", judgeController.Execute)
		judge.POST("/code", judgeController.SaveCode)
		judge.GET("/code/:problemID/:language/:codeType", judgeController.GetCode)
	}
}
