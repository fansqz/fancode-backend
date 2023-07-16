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
	CodeQuestionNumberIsExist = 11500 + iota //题目编号已存在
	CodeQuestionUpdateFailed
	CodeQuestionDeleteFailed
	CodeQuestionListFailed
	CodeQuestionNotExist
	CodeQuestionFileUploadFailed
)

var (
	ErrQuestionNumberIsExist    = NewError(CodeQuestionNumberIsExist, "question number is exist", ErrTypeBus)
	ErrQuestionUpdateFailed     = NewError(CodeQuestionUpdateFailed, "The question update failed", ErrTypeServer)
	ErrQuestionDeleteFailed     = NewError(CodeQuestionDeleteFailed, "The question delete failed", ErrTypeServer)
	ErrQuestionListFailed       = NewError(CodeQuestionListFailed, "Failed to get the question list", ErrTypeServer)
	ErrQuestionFileUploadFailed = NewError(CodeQuestionFileUploadFailed, "The question file storage failed", ErrTypeServer)
	ErrQuestionNotExist         = NewError(CodeQuestionNotExist, "The question does not exist", ErrTypeBus)
)
