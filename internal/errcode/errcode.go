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
	ErrCodeNormal                   = Error{ErrCode: CodeNormal, Message: "服务异常"}
	ErrCodeInvalidRequest           = Error{ErrCode: CodeInvalidRequest, Message: "无效的请求"}
	ErrCodeNoToken                  = Error{ErrCode: CodeNoToken, Message: "未提供令牌"}
	ErrCodeInvalidToken             = Error{ErrCode: CodeInvalidToken, Message: "无效的令牌"}
	ErrCodeTokenExpired             = Error{ErrCode: CodeTokenExpired, Message: "令牌已过期"}
	ErrCodeSendEmailCodeFailed      = Error{ErrCode: CodeSendEmailCodeFailed, Message: "发送邮箱验证码失败"}
	ErrCodeUserRegisterFailed       = Error{ErrCode: CodeUserRegisterFailed, Message: "用户注册失败"}
	ErrCodeUserLoginFailed          = Error{ErrCode: CodeUserLoginFailed, Message: "用户登录失败"}
	ErrCodeInvalidCode              = Error{ErrCode: CodeInvalidCode, Message: "无效的验证码"}
	ErrCodeUserAlreadyExists        = Error{ErrCode: CodeUserAlreadyExists, Message: "用户已存在"}
	ErrCodeUserNotFound             = Error{ErrCode: CodeUserNotFound, Message: "用户不存在"}
	ErrCodeInvalidPassword          = Error{ErrCode: CodeInvalidPassword, Message: "无效的密码"}
	ErrCodeCreateMatchFailed        = Error{ErrCode: CodeCreateMatchFailed, Message: "创建比赛失败"}
	ErrCodeJoinMatchFailed          = Error{ErrCode: CodeJoinMatchFailed, Message: "加入比赛失败"}
	ErrCodeStartMatchFailed         = Error{ErrCode: CodeStartMatchFailed, Message: "开始比赛失败"}
	ErrCodeEndMatchFailed           = Error{ErrCode: CodeEndMatchFailed, Message: "结束比赛失败"}
	ErrCodeLeaveMatchFailed         = Error{ErrCode: CodeLeaveMatchFailed, Message: "离开比赛失败"}
	ErrCodeKickPlayerFailed         = Error{ErrCode: CodeKickPlayerFailed, Message: "踢出比赛失败"}
	ErrCodeMatchPlayerFull          = Error{ErrCode: CodeMatchPlayerFull, Message: "比赛人员已满"}
	ErrCodeMatchNotPending          = Error{ErrCode: CodeMatchNotPending, Message: "比赛已开始"}
	ErrCodeMatchPlayerAlreadyJoined = Error{
		ErrCode: CodeMatchPlayerAlreadyJoined,
		Message: "玩家已加入",
	}
	ErrCodeMatchNotExists = Error{ErrCode: CodeMatchNotExists, Message: "比赛不存在"}
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
