package admin

import (
	"FanCode/controllers/admin"
	"github.com/gin-gonic/gin"
)

func SetupSysRoleRoutes(r *gin.Engine) {
	//题目相关路由
	role := r.Group("/manage/role")
	{
		roleController := admin.NewSysRoleController()
		role.POST("", roleController.InsertSysRole)
		role.PUT("", roleController.UpdateSysRole)
		role.DELETE("/:id", roleController.DeleteSysRole)
		role.GET("/list", roleController.GetSysRoleList)
		role.GET("/api/:id", roleController.GetApiIDsByRoleID)
		role.GET("/menu/:id", roleController.GetMenuIDsByRoleID)
		role.PUT("/api", roleController.UpdateRoleApis)
		role.PUT("/menu", roleController.UpdateRoleMenus)
	}
}
