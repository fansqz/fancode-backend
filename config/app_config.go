package config

// AppConfig
// @Description:应用配置
type AppConfig struct {
	Release         bool   `ini:"release"` //是否是上线模式
	Port            string `ini:"port"`    //端口
	ProUrl          string `ini:"proUrl"`
	DefaultPassword string `ini:"defaultPassword"`
	*MySqlConfig
	*RedisConfig
	*EmailConfig
	*ReleasePathConfig
	*COSConfig
	*FilePathConfig
}

type ReleasePathConfig struct {
	StartWith []string
}
