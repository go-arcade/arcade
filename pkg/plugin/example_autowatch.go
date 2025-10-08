package plugin

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/observabil/arcade/pkg/log"
)

// ExampleAutoWatch 演示如何使用插件自动监控功能
func ExampleAutoWatch() {
	// 1. 创建插件管理器
	manager := NewManager()
	manager.SetContext(context.Background())

	// 2. 从配置文件加载初始插件
	configPath := "./conf.d/plugins.yaml"
	if err := manager.LoadPluginsFromConfig(configPath); err != nil {
		log.Errorf("加载插件配置失败: %v", err)
		return
	}

	// 3. 初始化所有插件
	if err := manager.Init(context.Background()); err != nil {
		log.Errorf("初始化插件失败: %v", err)
		return
	}

	// 4. 启动自动监控
	// 监控 ./plugins 目录和配置文件
	watchDirs := []string{
		"./plugins",        // 插件目录
		"./plugins/notify", // 子目录也可以监控
		"./plugins/ci",
		"./plugins/cd",
	}

	if err := manager.StartAutoWatch(watchDirs, configPath); err != nil {
		log.Errorf("启动自动监控失败: %v", err)
		return
	}
	defer manager.StopAutoWatch()

	log.Info("=" + string([]byte{61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61}))
	log.Info("插件自动加载系统已启动")
	log.Info("=" + string([]byte{61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61}))
	log.Info("")
	log.Info("功能说明:")
	log.Info("  1. 将 .so 插件文件放入监控目录,系统会自动加载")
	log.Info("  2. 删除 .so 插件文件,系统会自动卸载")
	log.Info("  3. 修改 plugins.yaml 配置文件,系统会自动重载配置")
	log.Info("  4. 支持多级目录监控")
	log.Info("")
	log.Info("当前已加载的插件:")
	listLoadedPlugins(manager)
	log.Info("")

	// 5. 等待信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 定期显示插件列表
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Info("--- 当前插件状态 ---")
			listLoadedPlugins(manager)
			log.Info("")

		case sig := <-sigChan:
			log.Infof("收到信号 %v, 正在关闭...", sig)
			manager.StopAutoWatch()
			if err := manager.Cleanup(); err != nil {
				log.Errorf("清理插件失败: %v", err)
			}
			return
		}
	}
}

// ExampleManualReload 演示手动重载插件
func ExampleManualReload() {
	manager := NewManager()
	manager.SetContext(context.Background())

	// 加载初始插件
	if err := manager.Register("./plugins/stdout.so"); err != nil {
		log.Errorf("加载插件失败: %v", err)
		return
	}

	if err := manager.Init(context.Background()); err != nil {
		log.Errorf("初始化插件失败: %v", err)
		return
	}

	log.Info("初始插件已加载:")
	listLoadedPlugins(manager)

	// 等待一段时间后手动重载
	log.Info("5秒后将手动重载插件...")
	time.Sleep(5 * time.Second)

	// 手动重载插件
	if err := manager.ReloadPlugin("stdout"); err != nil {
		log.Errorf("重载插件失败: %v", err)
		return
	}

	log.Info("插件重载完成:")
	listLoadedPlugins(manager)
}

// listLoadedPlugins 列出已加载的插件
func listLoadedPlugins(manager *Manager) {
	plugins := manager.ListPlugins()
	if len(plugins) == 0 {
		log.Info("  (无插件)")
		return
	}

	for i, p := range plugins {
		log.Info(fmt.Sprintf("  %d. [%s] %s v%s", i+1, p.Type, p.Name, p.Version))
		log.Info(fmt.Sprintf("     路径: %s", p.Path))
		if p.Description != "" {
			log.Info(fmt.Sprintf("     描述: %s", p.Description))
		}
	}
}
