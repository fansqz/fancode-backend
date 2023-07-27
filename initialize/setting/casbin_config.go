package setting

import "gopkg.in/ini.v1"

type CasbinConfig struct {
	Mysql       string `ini:"mysql"`
	ModelConfig string `ini:"model_config"`
}

func NewCasbinConfig(cfg *ini.File) *CasbinConfig {
	casbinConfig := &CasbinConfig{}
	cfg.Section("casbin").MapTo(casbinConfig)
	return casbinConfig
}
