package routers

import (
	"FanCode/controllers"
	"github.com/gin-gonic/gin"
)

func SetupAccountRoutes(r *gin.Engine) {
	accountController := controllers.NewAccountController()
	account := r.Group("/account")
	{
		account.GET("/info", accountController.GetAccountInfo)
		account.PUT("", accountController.UpdateAccountInfo)
		account.POST("/password/reset", accountController.ResetPassword)
		account.POST("/password", accountController.ChangePassword)
		account.GET("/active/year", accountController.GetUserActivityYear)
		account.GET("/active/map", accountController.GetUserActivityMap)
	}
	avatar := r.Group("/avatar")
	{
		avatar.GET("/user/:avatarName", accountController.ReadAvatar)
		avatar.POST("", accountController.UploadAvatar)
	}
}
