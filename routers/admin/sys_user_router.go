package admin

import (
	"FanCode/controllers/admin"
	"github.com/gin-gonic/gin"
)

func SetupSysUserRoutes(r *gin.Engine) {
	//题目相关路由
	user := r.Group("/manage/user")
	{
		userController := admin.NewSysUserController()
		user.POST("", userController.InsertSysUser)
		user.PUT("", userController.UpdateSysUser)
		user.DELETE("/:id", userController.DeleteSysUser)
		user.GET("/list", userController.GetSysUserList)
		user.GET("/role/:id", userController.GetRoleIDsByUserID)
		user.PUT("/role", userController.UpdateUserRoles)
		user.GET("/simpleRole/list", userController.GetAllSimpleRole)
	}
}
