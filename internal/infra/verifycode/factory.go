package verifycode

import (
	"net/mail"
	"regexp"

	"github.com/superwhys/billiard-helper/internal/errcode"
)

var phoneRegex = regexp.MustCompile(`^\+?\d{7,15}$`)

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

func (f *DefaultSenderFactory) Pick(account string) (VerifyCodeSender, error) {
	if isEmailAccount(account) {
		return f.emailSender, nil
	}
	if isPhoneAccount(account) {
		return f.smsSender, nil
	}
	return nil, errcode.ErrCodeInvalidRequest
}

func isEmailAccount(account string) bool {
	addr, err := mail.ParseAddress(account)
	if err != nil {
		return false
	}
	return addr.Address == account
}

func isPhoneAccount(account string) bool {
	return phoneRegex.MatchString(account)
}
