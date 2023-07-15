package routers

import (
	"FanCode/controllers"
	"github.com/gin-gonic/gin"
)

func SetupQuestionRoutes(r *gin.Engine) {
	//用户相关
	question := r.Group("/question")
	{
		questionController := controllers.NewQuestionController()
		question.POST("/insert", questionController.InsertQuestion)
		question.PUT("/update", questionController.UpdateQuestion)
		question.DELETE("/delete/:id", questionController.DeleteQuestion)
		question.GET("/list", questionController.GetQuestionList)
		question.POST("/upload/file", questionController.UploadQuestionFile)
	}
}
