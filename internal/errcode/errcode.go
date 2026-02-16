package errcode

import (
	"errors"
	"fmt"
)

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

func (e Error) AddMessage(message string) Error {
	e.Message = fmt.Sprintf("%s, %s", e.Message, message)
	return e
}

var (
	ErrSysInternal     = Error{ErrCode: 100500, Message: "系统内部错误"}
	ErrNotFound        = Error{ErrCode: 100404, Message: "资源不存在"}
	ErrBadRequest      = Error{ErrCode: 100400, Message: "请求参数错误"}
	ErrUnauthorized    = Error{ErrCode: 100401, Message: "未登录"}
	ErrForbidden       = Error{ErrCode: 100403, Message: "禁止访问"}
	ErrTooManyRequests = Error{ErrCode: 100429, Message: "请求过于频繁，请稍后再试"}

	ErrNoToken      = Error{ErrCode: 400001, Message: "No Token"}
	ErrTokenExpired = Error{ErrCode: 400002, Message: "Token Expired"}
	ErrInvalidToken = Error{ErrCode: 400003, Message: "Invalid Token"}

	ErrCodeSendEmailCodeFailed      = ErrSysInternal.WithMessage("发送邮箱验证码失败")
	ErrCodeSendSMSCodeFailed        = ErrSysInternal.WithMessage("发送短信验证码失败")
	ErrCodeUserGetInfoFailed        = ErrSysInternal.WithMessage("获取用户信息失败")
	ErrCodeUserUpdateSelfInfoFailed = ErrSysInternal.WithMessage("更新用户信息失败")
	ErrCodeUserRegisterFailed       = ErrSysInternal.WithMessage("用户注册失败")
	ErrCodeUserLoginFailed          = ErrSysInternal.WithMessage("用户登录失败")
	ErrCodeUserLogoutFailed         = ErrSysInternal.WithMessage("用户登出失败")
	ErrCodeUserWechatLoginFailed    = ErrSysInternal.WithMessage("微信登录失败")
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
	ErrCodeUpdateMatchFailed        = ErrSysInternal.WithMessage("更新比赛失败")
	ErrCodeDeleteMatchFailed        = ErrSysInternal.WithMessage("删除比赛失败")
	ErrCodeMatchPlayerFull          = ErrBadRequest.WithMessage("比赛人员已满")
	ErrCodeMatchPlayerOutOfLimit    = ErrBadRequest.WithMessage("比赛人员超出限制")
	ErrCodeMatchNotPending          = ErrBadRequest.WithMessage("比赛已开始")
	ErrCodeMatchNotInProgress       = ErrBadRequest.WithMessage("比赛未开始")
	ErrCodeMatchPlayerAlreadyJoined = ErrBadRequest.WithMessage("玩家已加入")
	ErrCodeMatchNotExists           = ErrNotFound.WithMessage("比赛不存在")
	ErrCodeListMatchesFailed        = ErrSysInternal.WithMessage("获取比赛列表失败")
	ErrCodeMatchDetailFailed        = ErrSysInternal.WithMessage("获取比赛详情失败")
	ErrCodeMatchNotFound            = ErrNotFound.WithMessage("比赛不存在")
	ErrCodeMatchPlayerNotEnough     = ErrBadRequest.WithMessage("比赛人数不足")
	ErrCodeMatchRoundNotMatch       = ErrBadRequest.WithMessage("比赛轮数不匹配")
	ErrCodeNextRoundFailed          = ErrSysInternal.WithMessage("下一轮失败")
	ErrCodeMatchMaxRoundReached     = ErrBadRequest.WithMessage("比赛轮数已达最大值").WithErrCode(200001)

	ErrCodeSyncScoreFailed      = ErrSysInternal.WithMessage("同步分数失败")
	ErrCodeUndoScoreFailed      = ErrSysInternal.WithMessage("撤回分数失败")
	ErrCodeListScoresFailed     = ErrSysInternal.WithMessage("获取分数历史记录失败")
	ErrCodeReportFeedbackFailed = ErrSysInternal.WithMessage("反馈失败")
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
