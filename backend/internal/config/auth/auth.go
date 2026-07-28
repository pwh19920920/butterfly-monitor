package auth

import (
	sysConf "github.com/pwh19920920/butterfly/pkg/config"
	"github.com/spf13/viper"
)

const defaultExpireTime = 30 * 60 * 24
const defaultHeaderName = "Authorization"
const defaultHeaderType = "Bearer"

type Config struct {
	ExpireTime       int      `yaml:"expireTime"`       // 过期时间
	HeaderName       string   `yaml:"headerName"`       // 请求头名称
	HeaderType       string   `yaml:"headerType"`       // 请求头类型
	IgnorePath       []string `yaml:"ignorePath"`       // 忽略地址
	IgnorePrefixPath []string `yaml:"ignorePrefixPath"` // 忽略地址
	CommonPath       []string `yaml:"commonPath"`       // 普通登陆地址
}

type auth struct {
	Auth Config `yaml:"auth"`
}

var conf *auth

func GetAuthConf() *Config {
	if conf == nil {
		// 默认配置
		viper.SetDefault("auth.expireTime", defaultExpireTime)
		viper.SetDefault("auth.headerName", defaultHeaderName)
		viper.SetDefault("auth.headerType", defaultHeaderType)

		// 加载配置
		sysConf.LoadConf(&conf)
	}
	return &conf.Auth
}
