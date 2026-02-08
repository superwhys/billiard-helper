package errcode

import "errors"

type Error struct {
	ErrCode int
	Message string
}

func (e Error) Error() string {
	return e.Message
}

func (e Error) String() string {
	return e.Message
}

func (e Error) Code() int {
	return e.ErrCode
}

func (e Error) Is(target error) bool {
	switch t := target.(type) {
	case Error:
		return e.ErrCode != 0 && e.ErrCode == t.ErrCode
	case *Error:
		if t == nil {
			return false
		}
		return e.ErrCode != 0 && e.ErrCode == t.ErrCode
	default:
		return false
	}
}

func (e Error) WithErrCode(code int) Error {
	e.ErrCode = code
	return e
}

func (e Error) WithMessage(message string) Error {
	e.Message = message
	return e
}

const (
	CodeNormal = iota + 10000
	CodeInvalidRequest
	// Auth error codes
	CodeNoToken
	CodeInvalidToken
	CodeTokenExpired
	CodeSendEmailCodeFailed
	CodeUserRegisterFailed
	CodeUserLoginFailed
	CodeInvalidCode
	CodeUserAlreadyExists
	CodeUserNotFound
	CodeInvalidPassword

	// Match error codes
	CodeCreateMatchFailed
	CodeJoinMatchFailed
	CodeStartMatchFailed
	CodeEndMatchFailed
	CodeLeaveMatchFailed
	CodeKickPlayerFailed
	CodeMatchPlayerFull
	CodeMatchNotPending
	CodeMatchPlayerAlreadyJoined
	CodeMatchNotExists
)

var (
	ErrSysInternal     = Error{ErrCode: 100500, Message: "系统内部错误"}
	ErrNotFound        = Error{ErrCode: 100404, Message: "资源不存在"}
	ErrBadRequest      = Error{ErrCode: 100400, Message: "请求参数错误"}
	ErrUnauthorized    = Error{ErrCode: 100401, Message: "未登录"}
	ErrForbidden       = Error{ErrCode: 100403, Message: "禁止访问"}
	ErrTooManyRequests = Error{ErrCode: 100429, Message: "请求过多"}

	ErrNoToken      = Error{ErrCode: 400001, Message: "No Token"}
	ErrTokenExpired = Error{ErrCode: 400002, Message: "Token Expired"}
	ErrInvalidToken = Error{ErrCode: 400003, Message: "Invalid Token"}

	ErrCodeSendEmailCodeFailed      = ErrSysInternal.WithMessage("发送邮箱验证码失败")
	ErrCodeUserGetInfoFailed        = ErrSysInternal.WithMessage("获取用户信息失败")
	ErrCodeUserUpdateSelfInfoFailed = ErrSysInternal.WithMessage("更新用户信息失败")
	ErrCodeUserRegisterFailed       = ErrSysInternal.WithMessage("用户注册失败")
	ErrCodeUserLoginFailed          = ErrSysInternal.WithMessage("用户登录失败")
	ErrCodeInvalidCode              = ErrBadRequest.WithMessage("无效的验证码")
	ErrCodeUserAlreadyExists        = ErrBadRequest.WithMessage("用户已存在")
	ErrCodeUserNotFound             = ErrNotFound.WithMessage("用户不存在")
	ErrCodeInvalidPassword          = ErrBadRequest.WithMessage("无效的密码")
	ErrCodeCreateMatchFailed        = ErrSysInternal.WithMessage("创建比赛失败")
	ErrCodeJoinMatchFailed          = ErrSysInternal.WithMessage("加入比赛失败")
	ErrCodeStartMatchFailed         = ErrSysInternal.WithMessage("开始比赛失败")
	ErrCodeEndMatchFailed           = ErrSysInternal.WithMessage("结束比赛失败")
	ErrCodeLeaveMatchFailed         = ErrSysInternal.WithMessage("离开比赛失败")
	ErrCodeKickPlayerFailed         = ErrSysInternal.WithMessage("踢出比赛失败")
	ErrCodeMatchPlayerFull          = ErrBadRequest.WithMessage("比赛人员已满")
	ErrCodeMatchPlayerOutOfLimit    = ErrBadRequest.WithMessage("比赛人员超出限制")
	ErrCodeMatchNotPending          = ErrBadRequest.WithMessage("比赛已开始")
	ErrCodeMatchNotInProgress       = ErrBadRequest.WithMessage("比赛未开始")
	ErrCodeMatchPlayerAlreadyJoined = ErrBadRequest.WithMessage("玩家已加入")
	ErrCodeMatchNotExists           = ErrNotFound.WithMessage("比赛不存在")
	ErrCodeListMatchesFailed        = ErrSysInternal.WithMessage("获取比赛列表失败")
	ErrCodeMatchDetailFailed        = ErrSysInternal.WithMessage("获取比赛详情失败")
	ErrCodeMatchNotFound            = ErrNotFound.WithMessage("比赛不存在")
)

func AsErrcode(err error) (Error, bool) {
	if err == nil {
		return Error{}, false
	}

	ec := new(Error)
	if errors.As(err, ec) {
		return *ec, true
	}

	return Error{}, false
}
