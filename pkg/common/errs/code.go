package errs

const (
	ServerInternalError       = 500  // Server internal error
	ArgsError                 = 1001 // Input parameter error
	RecordNotFoundError       = 1002 // Record does not exist
	RecordAlreadyExistError   = 1003 // Record already exist
	AccessDeniedError         = 1004 // Access denied
	SignFormatError           = 1007 // 签名格式错误
	NoAvailableAgentUserError = 1005 // 没有空闲客服
	IdentityDoesNotExistError = 1006 // 商户不存在
	SignExpiredTimeError      = 1008 // 签名过期
	SignInvalidError          = 1009 // 签名无效

	// Token error codes.
	TokenUnknownError            = 1100 // Token unknown
	TokenNotExistError           = 1101 // Token not exist
	TokenExpiredError            = 1102 // Token expired
	TokenInvalidError            = 1103 // Token invalid
	TokenMalformedError          = 1104 // Token malformed
	TokenNotValidYetError        = 1105 // Token not valid
	TokenAndJa3HashMismatchError = 1106 // token and ja3hash not match

	// User error codes.
	AlreadyRegisteredError = 1200 // 账号已注册
	AccountNotFoundError   = 1201 // 账号没找到
	PasswordError          = 1202 // 密码错误
	AccountIsBlockedError  = 1203 // 账号被禁用

	PermissionIDNotFoundError = 1300 // permission id 错误
	NotPermissionError        = 1301 // 没有权限
	RoleIDNotFoundError       = 1302 // role id 错误

	WithdrawRewardNotEnoughError = 1400 // 可提现收益不足
)
