// Package routers
// @Author: fzw
// @Create: 2023/7/14
// @Description: 路由相关
package routers

import (
	"FanCode/controllers"
	"FanCode/setting"
	"github.com/gin-gonic/gin"
)

//
// Run
//  @Description: 启动路由
//
func Run() {
	if setting.Conf.Release {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	//设置静态文件位置
	r.Static("/static", "/")
	//ping
	r.GET("/ping", controllers.Ping)

	err := r.Run(":" + setting.Conf.Port)
	if err != nil {
		return
	}
}
