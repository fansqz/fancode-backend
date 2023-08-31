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
	}
	avatar := r.Group("/avatar")
	{
		avatar.GET("/user/:avatarName", accountController.ReadAvatar)
		avatar.POST("", accountController.UploadAvatar)
	}
}
