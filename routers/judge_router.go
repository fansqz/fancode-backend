package routers

import (
	"FanCode/controllers/user"
	"github.com/gin-gonic/gin"
)

func SetupJudgeRoutes(r *gin.Engine) {
	//用户相关
	judge := r.Group("/judge")
	{
		judgeController := user.NewJudgeController()
		judge.POST("/submit", judgeController.Submit)
		judge.POST("/execute", judgeController.Execute)
	}
}
