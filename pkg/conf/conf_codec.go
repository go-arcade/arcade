package conf

import (
	"github.com/go-kratos/kratos/v2/encoding"
	"github.com/pelletier/go-toml/v2"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/4 19:53
 * @file: config.go
 * @description: config
 */

const Name = "toml"

func init() {
	encoding.RegisterCodec(codec{})
}

type codec struct{}

func (codec) Marshal(v interface{}) ([]byte, error) {
	return toml.Marshal(v)
}

func (codec) Unmarshal(data []byte, v interface{}) error {
	return toml.Unmarshal(data, v)
}

func (codec) Name() string {
	return Name
}
