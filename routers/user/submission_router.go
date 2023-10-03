package user

import (
	"FanCode/controller/user"
	"github.com/gin-gonic/gin"
)

func SetupSubmissionRoutes(r *gin.Engine, submissionController user.SubmissionController) {
	//题目相关路由
	submission := r.Group("/submission")
	{
		submission.GET("/active/year", submissionController.GetUserActivityYear)
		submission.GET("/active/map/:year", submissionController.GetUserActivityMap)
		submission.GET("/list", submissionController.GetUserSubmissionList)
	}
}
