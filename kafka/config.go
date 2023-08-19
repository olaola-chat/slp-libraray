package kafka

import (
	"fmt"

	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/frame/g"
)

// Config config配置中的kafka定义
type Config struct {
	Host    []string
	Version string
}

// GetConfig 根据名字解析配置
func GetConfig(name string) (*Config, error) {
	config := &Config{}
	err := g.Cfg().GetStruct(fmt.Sprintf("go-kafka.%s", name), config)
	if err != nil {
		return config, gerror.Wrap(err, "redis config error")
	}
	return config, nil
}
