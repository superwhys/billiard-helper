package verifycode

import "context"

// VerifyCodeSender 发送验证码通用接口
type VerifyCodeSender interface {
	SendVerifyCode(ctx context.Context, account string, code string) error
	Channel() string
}
