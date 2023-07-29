package routers

import (
	"FanCode/controllers"
	"github.com/gin-gonic/gin"
)

func SetupSysRoleRoutes(r *gin.Engine) {
	//题目相关路由
	role := r.Group("/manage/role")
	{
		roleController := controllers.NewSysRoleController()
		role.POST("", roleController.InsertSysRole)
		role.PUT("", roleController.UpdateSysRole)
		role.DELETE("/:id", roleController.DeleteSysRole)
		role.GET("/list/:page/:pageSize", roleController.GetSysRoleList)
	}
}
