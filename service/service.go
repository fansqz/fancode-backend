package service

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewAccountService,
	NewAuthService,
	NewJudgeService,
	NewDebugService,
	NewProblemBankService,
	NewProblemService,
	NewProblemCaseService,
	NewSubmissionService,
	NewSysApiService,
	NewSysMenuService,
	NewSysRoleService,
	NewSysUserService,
	NewWsService,
)
