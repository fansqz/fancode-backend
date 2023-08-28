package config

import "gopkg.in/ini.v1"

type EmailConfig struct {
	Host     string `ini:"host"`
	Port     int    `int:"port"`
	User     string `ini:"user"`
	Password string `ini:"password"`
}

func NewEmailConfig(cfg *ini.File) *EmailConfig {
	emailConfig := &EmailConfig{}
	cfg.Section("email").MapTo(emailConfig)
	return emailConfig
}
