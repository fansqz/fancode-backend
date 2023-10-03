package service

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewAccountService,
	NewAuthService,
	NewJudgeService,
	NewProblemBankService,
	NewProblemService,
	NewSubmissionService,
	NewSysApiService,
	NewSysMenuService,
	NewSysRoleService,
	NewSysUserService,
)
