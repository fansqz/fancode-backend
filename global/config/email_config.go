package config

import "gopkg.in/ini.v1"

type EmailConfig struct {
	Host     string `ini:"host"`
	Port     int    `int:"port"`
	UserName string `ini:"userName"`
	Password string `ini:"password"`
}

func NewEmailConfig(cfg *ini.File) *EmailConfig {
	emailConfig := &EmailConfig{}
	cfg.Section("email").MapTo(emailConfig)
	return emailConfig
}
