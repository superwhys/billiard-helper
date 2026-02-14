package config

import (
	"errors"
	"time"

	"github.com/miebyte/goutils/emailutils"
)

type JwtConfig struct {
	JwtSecret         string
	JwtTimeout        time.Duration
	JwtRefreshTimeout time.Duration
}

func (c *JwtConfig) Validate() error {
	if c.JwtSecret == "" {
		return errors.New("jwt secret is required")
	}
	if c.JwtTimeout <= 0 {
		return errors.New("jwt timeout is required")
	}
	if c.JwtRefreshTimeout <= 0 {
		c.JwtRefreshTimeout = c.JwtTimeout * 24
	}
	return nil
}

type WechatConfig struct {
	AppID             string
	SecretID          string
	Jscode2SessionApi string
}

func (c *WechatConfig) Validate() error {
	if c.AppID == "" {
		return errors.New("appid is required")
	}
	if c.SecretID == "" {
		return errors.New("secretid is required")
	}
	if c.Jscode2SessionApi == "" {
		return errors.New("jscode2sessionapi is required")
	}
	return nil
}

type Config struct {
	JwtConfig    *JwtConfig
	EmailConfig  *emailutils.EmailConfig
	WechatConfig *WechatConfig
}

func (c *Config) Validate() error {
	if c.JwtConfig == nil {
		return errors.New("jwt config is required")
	}

	if err := c.JwtConfig.Validate(); err != nil {
		return err
	}

	if c.EmailConfig == nil {
		return errors.New("email config is required")
	}

	if c.WechatConfig == nil {
		return errors.New("wechat config is required")
	}

	if err := c.WechatConfig.Validate(); err != nil {
		return err
	}
	return nil
}
