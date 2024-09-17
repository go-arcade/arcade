package plugin

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/12 22:31
 * @file: plugin_interface.go
 * @description: plugin interface
 */

type Plugin interface {
	// Name plugin name
	Name() string
	// Description plugin description
	Description() string
	// Version plugin version
	Version() string
	// Register plugin init
	Register() error
	// AntiRegister plugin remove
	AntiRegister() error
	// Run plugin run
	Run() string
}
