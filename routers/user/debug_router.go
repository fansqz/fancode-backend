package user

import (
	"FanCode/controller/user"
	"github.com/gin-gonic/gin"
)

func SetupDebugRoutes(r *gin.Engine, debugController user.DebugController) {
	//用户相关
	judge := r.Group("/debug")
	{
		judge.POST("/session/create", debugController.CreateDebugSession)
		judge.GET("/sse/:key", debugController.CreateSseConnect)
		judge.POST("/start", debugController.Start)
		judge.POST("/step/in", debugController.StepIn)
		judge.POST("/step/out", debugController.StepOut)
		judge.POST("/step/over", debugController.StepOver)
		judge.POST("/continue", debugController.Continue)
		judge.POST("/sendToConsole", debugController.SendToConsole)
		judge.POST("/addBreakpoints", debugController.AddBreakpoints)
		judge.POST("/removeBreakpoints", debugController.RemoveBreakpoints)
		judge.POST("/session/close", debugController.CloseDebugSession)
	}
}
