// Package routers
// @Author: fzw
// @Create: 2023/7/14
// @Description: 路由相关
package routers

import (
	c "FanCode/controller"
	"FanCode/global"
	"FanCode/interceptor"
	"FanCode/routers/admin"
	"FanCode/routers/user"
	"github.com/gin-gonic/gin"
)

// SetupRouter
//
//	@Description: 启动路由
func SetupRouter(
	controller *c.Controller,
	corsInterceptor *interceptor.CorsInterceptor,
	requestInterceptor *interceptor.RequestInterceptor,
) *gin.Engine {
	if global.Conf.Release {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// 允许跨域
	r.Use(corsInterceptor.Cors())

	// 拦截非法用户
	r.Use(requestInterceptor.TokenAuthorize())

	//设置静态文件位置
	r.Static("/static", "/")

	//ping
	r.GET("/ping", c.Ping)

	SetupAuthRoutes(r, controller.AuthController)
	SetupAccountRoutes(r, controller.AccountController)
	admin.SetupSysApiRoutes(r, controller.ApiController)
	admin.SetupSysMenuRoutes(r, controller.MenuController)
	admin.SetupSysRoleRoutes(r, controller.RoleController)
	admin.SetupSysUserRoutes(r, controller.UserController)
	admin.SetupProblemBankRoutes(r, controller.ProblemBankManagementController)
	admin.SetupProblemRoutes(r, controller.ProblemManagementController)
	user.SetupJudgeRoutes(r, controller.JudgeController)
	user.SetupProblemRoutes(r, controller.ProblemController)
	user.SetupSubmissionRoutes(r, controller.SubmissionController)

	return r
}
