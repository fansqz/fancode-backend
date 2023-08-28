// Package routers
// @Author: fzw
// @Create: 2023/7/14
// @Description: 路由相关
package routers

import (
	"FanCode/controllers"
	"FanCode/global"
	"FanCode/interceptor"
	"FanCode/routers/admin"
	"FanCode/routers/user"
	"github.com/gin-gonic/gin"
)

// Run
//
//	@Description: 启动路由
func Run() {
	if global.Conf.Release {
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

	admin.SetupAuthRoutes(r)

	admin.SetupProblemRoutes(r)
	admin.SetupSysApiRoutes(r)
	admin.SetupSysMenuRoutes(r)
	admin.SetupSysRoleRoutes(r)
	admin.SetupSysUserRoutes(r)

	user.SetupJudgeRoutes(r)
	user.SetupProblemRoutes(r)

	err := r.Run(":" + global.Conf.Port)
	if err != nil {
		return
	}
}
