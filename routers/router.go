// Package routers
// @Author: fzw
// @Create: 2023/7/14
// @Description: 路由相关
package routers

import (
	"FanCode/controllers"
	"FanCode/interceptor"
	"FanCode/setting"
	"github.com/gin-gonic/gin"
)

// Run
//
//	@Description: 启动路由
func Run() {
	if setting.Conf.Release {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// 允许跨域
	r.Use(interceptor.Cors())

	// 拦截非法用户
	r.Use(interceptor.TokenAuthorize())

	//设置静态文件位置
	r.Static("/static", "/")

	//ping
	r.GET("/ping", controllers.Ping)

	SetupUserRoutes(r)
	SetupQuestionRoutes(r)

	err := r.Run(":" + setting.Conf.Port)
	if err != nil {
		return
	}
}
