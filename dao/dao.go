package dao

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewProblemAttemptDao,
	NewProblemBankDao,
	NewProblemDao,
	NewProblemCaseDao,
	NewSubmissionDao,
	NewSysApiDao,
	NewSysMenuDao,
	NewSysRoleDao,
	NewSysUserDao,
)
