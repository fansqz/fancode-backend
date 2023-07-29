package authset_error

// 服务器 10000 以上业务错误 服务器错误, 不同模块业务错误码间隔 500
// 前10000 留给公用的错误

/************User错误**************/
const (
	CodeUserNameOrPasswordWrong       = 11000 + iota // 用户名或密码错误
	CodeUserEmailIsNotValid                          // 电子邮件无效
	CodeUserEmailIsExist                             // 邮箱已存在
	CodeUserRegisterFail                             // 注册失败
	CodeUserNameIsExist                              // 用户名已存在
	CodeUserTypeNotSupport                           // 登陆类型错误
	CodeUserInvalidToken                             // 非法Token
	CodeUserPasswordNotEnoughAccuracy                // 用户密码精度不够
	CodeUserCreationFailed                           // 用户插入数据库失败
	CodePasswordEncodeFailed                         // 密码加密失败
	CodeUserNotExist                                 // 用户不存在
	CodeUserUnknownError                             // 用户服务未知错误
)

var (
	ErrUserNameIsExist               = NewError(CodeUserNameIsExist, "username already exist", ErrTypeBus)
	ErrUserPasswordNotEnoughAccuracy = NewError(CodeUserPasswordNotEnoughAccuracy, "The password is not accurate enough", ErrTypeBus)
	ErrUserCreationFailed            = NewError(CodeUserCreationFailed, "Failed to create user", ErrTypeServer)
	ErrPasswordEncodeFailed          = NewError(CodePasswordEncodeFailed, "Failed to encode password", ErrTypeServer)
	ErrUserNotExist                  = NewError(CodeUserNotExist, "The user does not exist", ErrTypeBus)
	ErrUserUnknownError              = NewError(CodeUserUnknownError, "Unknown error", ErrTypeServer)
	ErrUserNameOrPasswordWrong       = NewError(CodeUserNameOrPasswordWrong, "username or password wrong", ErrTypeBus)
	ErrUserEmailIsNotValid           = NewError(CodeUserEmailIsNotValid, "email is not valid", ErrTypeBadReq)
	ErrUserEmailIsExist              = NewError(CodeUserEmailIsExist, "email already exist", ErrTypeBus)

	ErrUserRegisterFail   = NewError(CodeUserRegisterFail, "register fail", ErrTypeBus)
	ErrUserTypeNotSupport = NewError(CodeUserTypeNotSupport, "type not support", ErrTypeBus)
	ErrUserInvalidToken   = NewError(CodeUserInvalidToken, "invalid token", ErrTypeBus)
)

/************Question错误**************/
const (
	CodeProblemCodeIsExist           = 11500 + iota //题目编号已存在
	CodeProblemCodeCheckFailed                      // 题目编号检测失败
	CodeProblemGetFailed                            // 获取题目失败
	CodeProblemInsertFailed                         // 添加题目失败
	CodeProblemUpdateFailed                         // 题目更新失败
	CodeProblemDeleteFailed                         // 题目删除失败
	CodeProblemListFailed                           // 获取题目列表失败
	CodeProblemNotExist                             // 题目不存在
	CodeProblemFileUploadFailed                     // 题目文件更新失败
	CodeProblemFileNotExist                         // 题目文件不存在
	CodeProblemZipFileDownloadFailed                // 题目压缩包文件下载失败
	CodeProblemFilePathNotExist                     // 题目文件路径不存在
)

var (
	ErrProblemCodeIsExist           = NewError(CodeProblemCodeIsExist, "problem code is exist", ErrTypeBus)
	ErrProblemCodeCheckFailed       = NewError(CodeProblemCodeCheckFailed, "The problem code check failed", ErrTypeServer)
	ErrProblemGetFailed             = NewError(CodeProblemGetFailed, "The problem get failed", ErrTypeServer)
	ErrProblemInsertFailed          = NewError(CodeProblemInsertFailed, "The problem insert failed", ErrTypeServer)
	ErrProblemUpdateFailed          = NewError(CodeProblemUpdateFailed, "The problem update failed", ErrTypeServer)
	ErrProblemDeleteFailed          = NewError(CodeProblemDeleteFailed, "The problem delete failed", ErrTypeServer)
	ErrProblemListFailed            = NewError(CodeProblemListFailed, "Failed to get the problem list", ErrTypeServer)
	ErrProblemFileUploadFailed      = NewError(CodeProblemFileUploadFailed, "The problem file storage failed", ErrTypeServer)
	ErrProblemNotExist              = NewError(CodeProblemNotExist, "The problem does not exist", ErrTypeBus)
	ErrProblemFileNotExist          = NewError(CodeProblemFileNotExist, "The problem file is not exist", ErrTypeBus)
	ErrProblemZipFileDownloadFailed = NewError(CodeProblemZipFileDownloadFailed, "The problem zipfile download failed", ErrTypeServer)
	ErrProblemFilePathNotExist      = NewError(CodeProblemFilePathNotExist, "题目编程文件不存在，需要上传编程文件", ErrTypeBus)
)

/************judge错误**************/
const (
	CodeSubmitFailed = 12000 + iota //题目编号已存在
	CodeExecuteFailed
	CodeCompileFailed
)

var (
	ErrSubmitFailed  = NewError(CodeSubmitFailed, "Submit error", ErrTypeBus)
	ErrExecuteFailed = NewError(CodeExecuteFailed, "Execute error", ErrTypeServer)
	ErrCompileFailed = NewError(CodeCompileFailed, "Compilation error", ErrTypeBus)
)

/************api相关错误**************/
const (
	CodeApiUnknownError = 12500 + iota //题目编号已存在
	CodeApiNotExist
)

var (
	ErrApiUnknownError = NewError(CodeApiUnknownError, "Unknown error", ErrTypeServer)
	ErrApiNotExist     = NewError(CodeApiNotExist, "The api is not exist", ErrTypeBus)
)

/************菜单管理**************/
const (
	CodeMenuUnknownError = 12500 + iota //题目编号已存在
	CodeMenuNotExist
)

var (
	ErrMenuUnknownError = NewError(CodeMenuUnknownError, "Unknown error", ErrTypeServer)
	ErrMenuNotExist     = NewError(CodeMenuNotExist, "The Menu is not exist", ErrTypeBus)
)
