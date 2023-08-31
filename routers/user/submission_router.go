package user

import (
	"FanCode/controllers/user"
	"github.com/gin-gonic/gin"
)

func SetupSubmissionRoutes(r *gin.Engine) {
	//题目相关路由
	submission := r.Group("/submission")
	{
		submissionHandler := user.NewSubmissionHandler()
		submission.GET("/active/year", submissionHandler.GetUserActivityYear)
		submission.GET("/active/map/:year", submissionHandler.GetUserActivityMap)
		submission.GET("/list/:page/:pageSize", submissionHandler.GetUserSubmissionList)
	}
}
