package verifycode

import (
	"github.com/superwhys/billiard-helper/internal/errcode"
	"github.com/superwhys/billiard-helper/internal/pkg/account"
)

type SenderFactory interface {
	Pick(account string) (VerifyCodeSender, error)
}

// DefaultSenderFactory 根据账号选择验证码发送策略
type DefaultSenderFactory struct {
	emailSender VerifyCodeSender
	smsSender   VerifyCodeSender
}

func NewSenderFactory(emailSender VerifyCodeSender, smsSender VerifyCodeSender) *DefaultSenderFactory {
	return &DefaultSenderFactory{
		emailSender: emailSender,
		smsSender:   smsSender,
	}
}

func (f *DefaultSenderFactory) Pick(accountValue string) (VerifyCodeSender, error) {
	if account.IsEmailAccount(accountValue) {
		return f.emailSender, nil
	}
	if account.IsPhoneAccount(accountValue) {
		return f.smsSender, nil
	}
	return nil, errcode.ErrBadRequest
}
