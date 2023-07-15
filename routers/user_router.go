package routers

import (
	"FanCode/controllers"
	"github.com/gin-gonic/gin"
)

func SetupUserRoutes(r *gin.Engine) {
	//用户相关
	user := r.Group("/user")
	{
		userController := controllers.NewUserController()
		user.POST("/register", userController.Register)
		user.POST("/login", userController.Login)
		user.POST("/update/password", userController.ChangePassword)
		user.GET("/get/info", userController.GetUserInfo)
	}
}
