package conf

import (
	"github.com/google/wire"
)

// ProviderSet 提供配置相关的依赖
var ProviderSet = wire.NewSet(ProvideConf)

// ProvideConf 提供完整配置实例
func ProvideConf(configFile string) AppConfig {
	return NewConf(configFile)
}
