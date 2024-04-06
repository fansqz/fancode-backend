package user

import (
	"FanCode/controller/user"
	"github.com/gin-gonic/gin"
)

func SetupDebugRoutes(r *gin.Engine, debugController user.DebugController) {
	//用户相关
	judge := r.Group("/debug")
	{
		judge.POST("/start", debugController.Start)
		judge.GET("/sse/:key", debugController.CreateSseConnect)
		judge.POST("/next", debugController.Next)
		judge.POST("/step", debugController.Step)
		judge.POST("/continue", debugController.Continue)
		judge.POST("/sendToConsole", debugController.SendToConsole)
		judge.POST("/addBreakpoints", debugController.AddBreakpoints)
		judge.POST("/removeBreakpoints", debugController.RemoveBreakpoints)
		judge.POST("/terminate", debugController.Terminate)
	}
}
