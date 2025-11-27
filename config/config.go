package config

import (
	"errors"
	"time"

	"github.com/miebyte/goutils/emailutils"
)

type JwtConfig struct {
	JwtSecret string
	// trial token timeout, if not trial, will not be expired
	JwtTimeout time.Duration
}

func (c *JwtConfig) Validate() error {
	if c.JwtSecret == "" {
		return errors.New("jwt secret is required")
	}
	if c.JwtTimeout <= 0 {
		return errors.New("jwt timeout is required")
	}
	return nil
}

type Config struct {
	JwtConfig   *JwtConfig
	EmailConfig *emailutils.EmailConfig
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
	return nil
}
