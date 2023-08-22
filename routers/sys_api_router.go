package routers

import (
	"FanCode/controllers/admin"
	"github.com/gin-gonic/gin"
)

func SetupSysApiRoutes(r *gin.Engine) {
	//题目相关路由
	api := r.Group("/manage/api")
	{
		apiController := admin.NewSysApiController()
		api.GET("/:id", apiController.GetApiByID)
		api.POST("", apiController.InsertApi)
		api.PUT("", apiController.UpdateApi)
		api.DELETE("/:id", apiController.DeleteApiByID)
		api.GET("/tree", apiController.GetApiTree)
	}
}
