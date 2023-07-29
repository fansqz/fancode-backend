package routers

import (
	"FanCode/controllers"
	"github.com/gin-gonic/gin"
)

func SetupSysMenuRoutes(r *gin.Engine) {
	//题目相关路由
	menu := r.Group("/manage/menu")
	{
		menuController := controllers.NewSysMenuController()
		menu.GET("/:id", menuController.GetMenuByID)
		menu.POST("", menuController.InsertMenu)
		menu.PUT("", menuController.UpdateMenu)
		menu.DELETE("/:id", menuController.DeleteMenuByID)
		menu.GET("/tree", menuController.GetMenuTree)
	}
}
