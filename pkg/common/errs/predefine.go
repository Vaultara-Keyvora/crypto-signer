package errs

var (
	ErrArgs                 = NewCodeError(ArgsError, "ArgsError")
	ErrNoPermission         = NewCodeError(NotPermissionError, "NoPermissionError")
	ErrInternalServer       = NewCodeError(ServerInternalError, "ServerInternalError")
	ErrRecordNotFound       = NewCodeError(RecordNotFoundError, "RecordNotFoundError")
	ErrNoAvailableAgentUser = NewCodeError(NoAvailableAgentUserError, "NoAvailableAgentUser")
	ErrIdentityDoesNotExist = NewCodeError(IdentityDoesNotExistError, "IdentityDoesNotExistError")
	ErrSignFormat           = NewCodeError(SignFormatError, "SignFormatError")
	ErrSignExpiredTime      = NewCodeError(SignExpiredTimeError, "SignExpiredTimeError")
	ErrSignInvalid          = NewCodeError(SignInvalidError, "SignInvalidError")

	ErrTokenExpired            = NewCodeError(TokenExpiredError, "TokenExpiredError")
	ErrTokenInvalid            = NewCodeError(TokenInvalidError, "TokenInvalidError")
	ErrTokenMalformed          = NewCodeError(TokenMalformedError, "TokenMalformedError")
	ErrTokenNotValidYet        = NewCodeError(TokenNotValidYetError, "TokenNotValidYetError")
	ErrTokenUnknown            = NewCodeError(TokenUnknownError, "TokenUnknownError")
	ErrTokenNotExist           = NewCodeError(TokenNotExistError, "TokenNotExistError")
	ErrTokenAndJa3HashMismatch = NewCodeError(TokenAndJa3HashMismatchError, "TokenAndJa3HashMismatchError")

	ErrWithdrawRewardNotEnough = NewCodeError(WithdrawRewardNotEnoughError, "WithdrawRewardNotEnoughError")
)
