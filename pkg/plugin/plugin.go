package plugin

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/12 22:38
 * @file: plugin.go
 * @description:
 */

func NewPlugin() interface{} {

	manager := NewManager()

	if err := manager.Register("plugin.so"); err != nil {
		panic(err)
	}

	manager.ListPlugins()

	if _, err := manager.Run("plugin"); err != nil {
		panic(err)
	}

	return nil
}
