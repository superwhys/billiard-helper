package user

import (
	"errors"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

type Email struct {
	value string
}

func NewEmail(email string) (Email, error) {
	if email == "" {
		return Email{}, errors.New("email cannot be empty")
	}
	// 简单的正则校验
	pattern := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(pattern, email)
	if !match {
		return Email{}, errors.New("invalid email format")
	}
	return Email{value: email}, nil
}

func (e Email) String() string { return e.value }

type Password struct {
	hash string
}

// NewPasswordFromPlain 从明文创建
func NewPasswordFromPlain(plain string) (Password, error) {
	if len(plain) < 6 {
		return Password{}, errors.New("password must be at least 6 characters")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return Password{}, err
	}
	return Password{hash: string(hash)}, nil
}

// NewPasswordFromHash 从哈希加载
func NewPasswordFromHash(hash string) Password {
	return Password{hash: hash}
}

func (p Password) Hash() string { return p.hash }

func (p Password) Compare(plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(p.hash), []byte(plain))
	return err == nil
}
