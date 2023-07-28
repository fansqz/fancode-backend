package routers

import (
	"FanCode/controllers"
	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(r *gin.Engine) {
	//用户相关
	auth := r.Group("/auth")
	{
		authController := controllers.NewAuthController()
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
		auth.POST("/update/password", authController.ChangePassword)
		auth.GET("/get/info", authController.GetUserInfo)
	}
}
