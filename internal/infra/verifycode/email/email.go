package email

import (
	"context"
	"fmt"

	"github.com/miebyte/goutils/emailutils"
	"github.com/superwhys/billiard-helper/internal/infra/verifycode"
)

const (
	emailSubject = "Billiard Helper - 邮箱验证码"
	emailBody    = "你的验证码是: %s"
)

type EmailSender struct {
	emailClient *emailutils.EmailClient
}

func NewEmailSender(emailConfig *emailutils.EmailConfig) *EmailSender {
	return &EmailSender{emailClient: emailutils.NewEmailClient(emailConfig)}
}

func (s *EmailSender) SendVerifyCode(ctx context.Context, email string, code string) error {
	return s.emailClient.Send(email, emailSubject, fmt.Sprintf(emailBody, code))
}

func (s *EmailSender) Channel() string {
	return "email"
}

var _ verifycode.VerifyCodeSender = (*EmailSender)(nil)
