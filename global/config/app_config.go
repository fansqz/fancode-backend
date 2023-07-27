package config

// AppConfig
// @Description:应用配置
type AppConfig struct {
	Release          bool   `ini:"release"` //是否是上线模式
	Port             string `ini:"port"`    //端口
	ReleaseStartPath string `ini:"releaseStartPath"`
	ProUrl           string `ini:"proUrl"`
	*MySqlConfig
	*ReleasePathConfig
	*COSConfig
	*FilePathConfig
	*CasbinConfig
}

type ReleasePathConfig struct {
	StartWith []string
}
