package errcode

import "errors"

type ErrCode uint

const (
	ErrCodeNormal ErrCode = iota + 10000
	ErrCodeInvalidRequest
	// Auth error codes
	ErrCodeNoToken
	ErrCodeInvalidToken
	ErrCodeTokenExpired
	ErrCodeSendEmailCodeFailed
	ErrCodeUserRegisterFailed
	ErrCodeUserLoginFailed
	ErrCodeInvalidCode
	ErrCodeUserAlreadyExists
	ErrCodeUserNotFound
	ErrCodeInvalidPassword

	// Match error codes
	ErrCodeCreateMatchFailed
	ErrCodeJoinMatchFailed
	ErrCodeStartMatchFailed
	ErrCodeEndMatchFailed
	ErrCodeLeaveMatchFailed
	ErrCodeKickPlayerFailed
	ErrCodeMatchPlayerFull
	ErrCodeMatchNotPending
	ErrCodeMatchPlayerAlreadyJoined
	ErrCodeMatchNotExists
)

func (c ErrCode) Error() string {
	return c.String()
}

func (c ErrCode) Code() int {
	return int(c)
}

func (c ErrCode) String() string {
	switch c {
	case ErrCodeNormal:
		return "服务异常"
	case ErrCodeInvalidRequest:
		return "无效的请求"
	case ErrCodeNoToken:
		return "未提供令牌"
	case ErrCodeInvalidToken:
		return "无效的令牌"
	case ErrCodeTokenExpired:
		return "令牌已过期"
	case ErrCodeSendEmailCodeFailed:
		return "发送邮箱验证码失败"
	case ErrCodeUserRegisterFailed:
		return "用户注册失败"
	case ErrCodeUserLoginFailed:
		return "用户登录失败"
	case ErrCodeInvalidCode:
		return "无效的验证码"
	case ErrCodeUserAlreadyExists:
		return "用户已存在"
	case ErrCodeUserNotFound:
		return "用户不存在"
	case ErrCodeCreateMatchFailed:
		return "创建比赛失败"
	case ErrCodeJoinMatchFailed:
		return "加入比赛失败"
	case ErrCodeStartMatchFailed:
		return "开始比赛失败"
	case ErrCodeEndMatchFailed:
		return "结束比赛失败"
	case ErrCodeLeaveMatchFailed:
		return "离开比赛失败"
	case ErrCodeKickPlayerFailed:
		return "踢出比赛失败"
	case ErrCodeMatchPlayerFull:
		return "比赛人员已满"
	case ErrCodeMatchNotPending:
		return "比赛已开始"
	case ErrCodeMatchPlayerAlreadyJoined:
		return "玩家已加入"
	case ErrCodeMatchNotExists:
		return "比赛不存在"
	case ErrCodeInvalidPassword:
		return "无效的密码"
	default:
		return "未知错误"
	}
}

func AsErrcode(err error) (ErrCode, bool) {
	ec := new(ErrCode)
	if errors.As(err, ec) {
		return *ec, true
	}

	return 0, false
}
