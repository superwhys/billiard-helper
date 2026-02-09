package sms

import (
	"context"

	"github.com/miebyte/goutils/logging"
	"github.com/superwhys/billiard-helper/internal/errcode"
	"github.com/superwhys/billiard-helper/internal/infra/verifycode"
)

// SMSSender 提供短信验证码发送
type SMSSender struct{}

func NewSMSSender() *SMSSender {
	return &SMSSender{}
}

func (s *SMSSender) SendVerifyCode(ctx context.Context, phone string, code string) error {
	logging.Infoc(ctx, "send sms verify code, phone: %s", phone)
	return errcode.ErrCodeSendSMSCodeFailed.AddMessage("暂不支持短信验证码")
}

func (s *SMSSender) Channel() string {
	return "sms"
}

var _ verifycode.VerifyCodeSender = (*SMSSender)(nil)
