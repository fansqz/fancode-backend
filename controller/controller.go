package controller

import (
	"FanCode/controller/admin"
	"FanCode/controller/user"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	NewAccountController,
	NewAuthController,
	NewFileController,
	admin.NewProblemBankManagementController,
	admin.NewProblemManagementController,
	admin.NewSysApiController,
	admin.NewSysMenuController,
	admin.NewSysRoleController,
	admin.NewSysUserController,
	user.NewJudgeController,
	user.NewProblemController,
	user.NewProblemBankController,
	user.NewSubmissionController,
	NewController,
)

type Controller struct {
	ProblemBankManagementController admin.ProblemBankManagementController
	ProblemManagementController     admin.ProblemManagementController
	ApiController                   admin.SysApiController
	MenuController                  admin.SysMenuController
	RoleController                  admin.SysRoleController
	UserController                  admin.SysUserController
	JudgeController                 user.JudgeController
	ProblemController               user.ProblemController
	ProblemBankController           user.ProblemBankController
	SubmissionController            user.SubmissionController
	AccountController               AccountController
	AuthController                  AuthController
}

func NewController(
	problemBankManagementController admin.ProblemBankManagementController,
	problemManagementController admin.ProblemManagementController,
	apiController admin.SysApiController,
	menuController admin.SysMenuController,
	roleController admin.SysRoleController,
	userController admin.SysUserController,
	judgeController user.JudgeController,
	problemController user.ProblemController,
	problemBankController user.ProblemBankController,
	submissionController user.SubmissionController,
	accountController AccountController,
	authController AuthController,
) *Controller {
	return &Controller{
		ProblemBankManagementController: problemBankManagementController,
		ProblemManagementController:     problemManagementController,
		ApiController:                   apiController,
		MenuController:                  menuController,
		RoleController:                  roleController,
		UserController:                  userController,
		JudgeController:                 judgeController,
		ProblemController:               problemController,
		ProblemBankController:           problemBankController,
		SubmissionController:            submissionController,
		AccountController:               accountController,
		AuthController:                  authController,
	}
}
