package errcode

import "errors"

type ErrCode uint

const (
	ErrCodeNormal ErrCode = iota + 10000
	ErrCodeInvalidRequest
	// Auth error codes
	ErrCodeNoToken
	ErrCodeSendEmailCodeFailed
	ErrCodeUserRegisterFailed
	ErrCodeUserLoginFailed

	// Room error codes
	ErrCodeCreateRoomFailed
	ErrCodeJoinRoomFailed
	ErrCodeStartRoomFailed
	ErrCodeEndRoomFailed
	ErrCodeLeaveRoomFailed
	ErrCodeKickPlayerFailed
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
	case ErrCodeSendEmailCodeFailed:
		return "发送邮箱验证码失败"
	case ErrCodeUserRegisterFailed:
		return "用户注册失败"
	case ErrCodeUserLoginFailed:
		return "用户登录失败"
	case ErrCodeCreateRoomFailed:
		return "创建房间失败"
	case ErrCodeJoinRoomFailed:
		return "加入房间失败"
	case ErrCodeStartRoomFailed:
		return "开始比赛失败"
	case ErrCodeEndRoomFailed:
		return "结束比赛失败"
	case ErrCodeLeaveRoomFailed:
		return "离开房间失败"
	case ErrCodeKickPlayerFailed:
		return "踢出玩家失败"

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
