// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"FanCode/config"
	"FanCode/controller"
	"FanCode/controller/admin"
	"FanCode/controller/user"
	"FanCode/dao"
	"FanCode/interceptor"
	"FanCode/routers"
	"FanCode/service"
	"net/http"
)

// Injectors from wire.go:

func initApp(appConfig *config.AppConfig) (*http.Server, error) {
	problemBankDao := dao.NewProblemBankDao()
	problemDao := dao.NewProblemDao()
	sysUserDao := dao.NewSysUserDao()
	problemBankService := service.NewProblemBankService(appConfig, problemBankDao, problemDao, sysUserDao)
	problemBankManagementController := admin.NewProblemBankManagementController(problemBankService)
	problemCaseDao := dao.NewProblemCaseDao()
	problemAttemptDao := dao.NewProblemAttemptDao()
	problemService := service.NewProblemService(appConfig, problemDao, problemCaseDao, problemAttemptDao)
	problemManagementController := admin.NewProblemManagementController(problemService)
	problemCaseService := service.NewProblemCaseService(appConfig, problemCaseDao, problemDao)
	problemCaseManagementController := admin.NewProblemCaseManagementController(problemCaseService)
	sysApiDao := dao.NewSysApiDao()
	sysApiService := service.NewSysApiService(sysApiDao)
	sysApiController := admin.NewSysApiController(sysApiService)
	sysMenuDao := dao.NewSysMenuDao()
	sysMenuService := service.NewSysMenuService(sysMenuDao)
	sysMenuController := admin.NewSysMenuController(sysMenuService)
	sysRoleDao := dao.NewSysRoleDao()
	sysRoleService := service.NewSysRoleService(sysRoleDao)
	sysRoleController := admin.NewSysRoleController(sysRoleService)
	sysUserService := service.NewSysUserService(appConfig, sysUserDao, sysRoleDao)
	sysUserController := admin.NewSysUserController(sysUserService)
	submissionDao := dao.NewSubmissionDao()
	judgeService := service.NewJudgeService(appConfig, problemService, submissionDao, problemAttemptDao, problemDao, problemCaseDao)
	judgeController := user.NewJudgeController(judgeService)
	debugService := service.NewDebugService(appConfig, judgeService)
	debugController := user.NewDebugController(debugService)
	problemController := user.NewProblemController(problemService)
	problemBankController := user.NewProblemBankController(problemBankService)
	submissionService := service.NewSubmissionService(submissionDao, problemDao)
	submissionController := user.NewSubmissionController(submissionService)
	accountService := service.NewAccountService(appConfig, sysUserDao)
	accountController := controller.NewAccountController(accountService)
	authService := service.NewAuthService(appConfig, sysUserDao, sysMenuDao, sysRoleDao)
	authController := controller.NewAuthController(authService)
	controllerController := controller.NewController(problemBankManagementController, problemManagementController, problemCaseManagementController, sysApiController, sysMenuController, sysRoleController, sysUserController, judgeController, debugController, problemController, problemBankController, submissionController, accountController, authController)
	recoverPanicInterceptor := interceptor.NewRecoverPanicInterceptor()
	corsInterceptor := interceptor.NewCorsInterceptor()
	requestInterceptor := interceptor.NewRequestInterceptor(sysRoleService, sysUserService)
	engine := routers.SetupRouter(appConfig, controllerController, recoverPanicInterceptor, corsInterceptor, requestInterceptor)
	server := newApp(engine, appConfig)
	return server, nil
}
