package errcode

import "errors"

type ErrCode uint

const (
	ErrCodeNormal ErrCode = iota + 10000
	ErrCodeInvalidRequest
	// Auth error codes
	ErrCodeNoToken
	ErrCodeSendEmailCodeFailed
	ErrCodeEmailCodeInvalid
	ErrCodeEmailCodeCooldown
	ErrCodeUserAlreadyExists
	ErrCodeUserRegisterFailed
	ErrCodeUserLoginFailed
	ErrCodeUserNotFound
	ErrCodeUserPasswordInvalid
	// Room error codes
	ErrCodeRoomNotFound
	ErrCodeCreateRoomFailed
	ErrCodeGetRoomFailed
	ErrCodeGetUserRoomsFailed
	ErrCodeJoinRoomFailed
	ErrCodeLeaveRoomFailed
	ErrCodeDeleteRoomFailed
	// Player error codes
	ErrCodePlayerAlreadyJoined
	ErrCodePlayerNotFound
	// Scores error codes
	ErrCodeAddScoreFailed
	ErrCodeMinusScoreFailed
	ErrCodeResetScoreFailed
	ErrCodeGetRoomScoresFailed
	// Session error codes
	ErrCodeSessionNotFound
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
	case ErrCodeRoomNotFound:
		return "房间不存在"
	case ErrCodeCreateRoomFailed:
		return "创建房间失败"
	case ErrCodeGetRoomFailed:
		return "获取房间失败"
	case ErrCodeGetUserRoomsFailed:
		return "获取用户房间列表失败"
	case ErrCodeJoinRoomFailed:
		return "加入房间失败"
	case ErrCodeLeaveRoomFailed:
		return "离开房间失败"
	case ErrCodePlayerAlreadyJoined:
		return "玩家已加入房间"
	case ErrCodePlayerNotFound:
		return "玩家不存在"
	case ErrCodeAddScoreFailed:
		return "添加分数失败"
	case ErrCodeMinusScoreFailed:
		return "减少分数失败"
	case ErrCodeResetScoreFailed:
		return "重置分数失败"
	case ErrCodeGetRoomScoresFailed:
		return "获取房间分数失败"
	case ErrCodeSendEmailCodeFailed:
		return "发送邮箱验证码失败"
	case ErrCodeEmailCodeInvalid:
		return "邮箱验证码无效"
	case ErrCodeEmailCodeCooldown:
		return "邮箱验证码冷却中"
	case ErrCodeUserAlreadyExists:
		return "用户已存在"
	case ErrCodeUserRegisterFailed:
		return "用户注册失败"
	case ErrCodeUserNotFound:
		return "用户不存在"
	case ErrCodeUserPasswordInvalid:
		return "用户密码不正确"
	case ErrCodeUserLoginFailed:
		return "用户登录失败"
	case ErrCodeNoToken:
		return "未提供 token"
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
