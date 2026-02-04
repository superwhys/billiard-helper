package account

import (
	"net/mail"
	"regexp"
)

var phoneRegex = regexp.MustCompile(`^\+?\d{7,15}$`)

// IsEmailAccount 判断账号是否为邮箱
func IsEmailAccount(account string) bool {
	addr, err := mail.ParseAddress(account)
	if err != nil {
		return false
	}
	return addr.Address == account
}

// IsPhoneAccount 判断账号是否为手机号
func IsPhoneAccount(account string) bool {
	return phoneRegex.MatchString(account)
}
