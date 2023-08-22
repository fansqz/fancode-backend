package routers

import (
	"FanCode/controllers/user"
	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(r *gin.Engine) {
	//用户相关
	auth := r.Group("/auth")
	{
		authController := user.NewAuthController()
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
		auth.POST("/update/password", authController.ChangePassword)
		auth.GET("/get/info", authController.GetUserInfo)
	}
}
