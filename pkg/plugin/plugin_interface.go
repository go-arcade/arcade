package plugin

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/12 22:31
 * @file: plugin_interface.go
 * @description: plugin interface
 */

type Plugin interface {
	Name() string
	Version() string
	Init() error
	Run() string
}
