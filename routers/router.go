// Package routers
// @Author: fzw
// @Create: 2023/7/14
// @Description: 路由相关
package routers

import (
	conf "FanCode/config"
	c "FanCode/controller"
	"FanCode/interceptor"
	"FanCode/routers/admin"
	"FanCode/routers/user"
	"github.com/gin-gonic/gin"
)

// SetupRouter
//
//	@Description: 启动路由
func SetupRouter(
	config *conf.AppConfig,
	controller *c.Controller,
	panicInterceptor *interceptor.RecoverPanicInterceptor,
	corsInterceptor *interceptor.CorsInterceptor,
	requestInterceptor *interceptor.RequestInterceptor,
) *gin.Engine {
	if config.Release {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// 拦截panic
	r.Use(panicInterceptor.RecoverPanic())
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
	admin.SetupProblemCaseRoutes(r, controller.ProblemCaseManagementController)
	user.SetupJudgeRoutes(r, controller.JudgeController)
	user.SetupDebugRoutes(r, controller.DebugController)
	user.SetupProblemRoutes(r, controller.ProblemController)
	user.SetupProblemBankRoutes(r, controller.ProblemBankController)
	user.SetupSubmissionRoutes(r, controller.SubmissionController)

	return r
}
