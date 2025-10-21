package agent

import (
	"context"
	"fmt"
	"os"
	"time"

	agentv1 "github.com/observabil/arcade/api/agent/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/10/19
 * @file: agent_id_manager_example.go
 * @description: Agent ID管理器使用示例
 */

// ExampleAgentRegistration 示例：Agent注册流程（确保ID一致性）
func ExampleAgentRegistration() {
	// 1. 创建ID管理器
	idManager := NewAgentIDManager("./data/agent.id")

	// 2. 获取主机名
	hostname, _ := os.Hostname()
	serverAddr := "localhost:9090"

	// 3. 加载或生成Agent ID
	agentID, isNew, err := idManager.LoadOrGenerate(hostname, serverAddr)
	if err != nil {
		fmt.Printf("加载Agent ID失败: %v\n", err)
		return
	}

	if isNew {
		fmt.Printf("生成新的Agent ID: %s\n", agentID)
	} else {
		fmt.Printf("使用已存在的Agent ID: %s\n", agentID)
	}

	// 4. 连接到gRPC Server
	conn, err := grpc.Dial(serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		fmt.Printf("连接Server失败: %v\n", err)
		return
	}
	defer conn.Close()

	client := agentv1.NewAgentServiceClient(conn)

	// 5. 构建注册请求
	req := &agentv1.RegisterRequest{
		AgentId:           agentID, // 使用本地保存的或新生成的ID
		Hostname:          hostname,
		Ip:                getLocalIP(),
		Os:                getOS(),
		Arch:              getArch(),
		Version:           "1.0.0",
		MaxConcurrentJobs: 5,
		Labels: map[string]string{
			"env":    "production",
			"region": "us-west",
		},
		InstalledPlugins: []string{"bash", "docker"},
	}

	// 6. 发送注册请求
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.Register(ctx, req)
	if err != nil {
		fmt.Printf("注册失败: %v\n", err)
		return
	}

	if !resp.Success {
		fmt.Printf("注册被拒绝: %s\n", resp.Message)
		return
	}

	// 7. 保存Server确认的Agent ID
	// 重要：Server可能返回不同的ID（例如检测到重复注册）
	if err := idManager.UpdateAfterRegister(resp.AgentId, serverAddr); err != nil {
		fmt.Printf("保存Agent ID失败: %v\n", err)
		return
	}

	if resp.AgentId != agentID {
		fmt.Printf("Server分配了不同的Agent ID: %s -> %s\n", agentID, resp.AgentId)
		agentID = resp.AgentId
	}

	fmt.Printf("注册成功！Agent ID: %s, 心跳间隔: %d秒\n",
		resp.AgentId, resp.HeartbeatInterval)

	// 8. 启动心跳（使用确认的Agent ID）
	startHeartbeat(client, agentID, resp.HeartbeatInterval)
}

// startHeartbeat 启动心跳协程
func startHeartbeat(client agentv1.AgentServiceClient, agentID string, interval int64) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		resp, err := client.Heartbeat(ctx, &agentv1.HeartbeatRequest{
			AgentId:           agentID,
			Status:            agentv1.AgentStatus_AGENT_STATUS_IDLE,
			RunningJobsCount:  0,
			MaxConcurrentJobs: 5,
			Metrics: map[string]string{
				"cpu":    "20%",
				"memory": "512MB",
			},
		})

		cancel()

		if err != nil {
			fmt.Printf("心跳失败: %v\n", err)
		} else if resp.Success {
			fmt.Printf("心跳成功，服务器时间: %d\n", resp.Timestamp)
		}
	}
}

// 辅助函数
func getLocalIP() string {
	// 实现获取本机IP的逻辑
	return "192.168.1.100"
}

func getOS() string {
	return "linux"
}

func getArch() string {
	return "amd64"
}

// ExampleAgentRestart 示例：Agent重启场景
func ExampleAgentRestart() {
	fmt.Println("\n=== Agent重启场景 ===")

	idManager := NewAgentIDManager("./data/agent.id")
	hostname, _ := os.Hostname()

	// Agent重启后，会自动加载之前保存的ID
	agentID, isNew, err := idManager.LoadOrGenerate(hostname, "localhost:9090")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	if isNew {
		fmt.Println("这是首次启动，生成了新ID")
	} else {
		fmt.Printf("Agent重启，使用已保存的ID: %s\n", agentID)

		// 显示ID信息
		info := idManager.GetInfo()
		fmt.Printf("注册时间: %s\n", info.RegisteredAt.Format(time.RFC3339))
		fmt.Printf("上次更新: %s\n", info.LastUpdateAt.Format(time.RFC3339))
	}
}

// ExampleIDMigration 示例：Agent迁移场景
func ExampleIDMigration() {
	fmt.Println("\n=== Agent迁移场景 ===")

	// 场景：需要将Agent从一台机器迁移到另一台机器，保持相同的Agent ID

	// 步骤1：在旧机器上导出Agent ID
	oldManager := NewAgentIDManager("/old/path/agent.id")
	oldInfo := oldManager.GetInfo()
	if oldInfo != nil {
		fmt.Printf("旧机器的Agent ID: %s\n", oldInfo.AgentID)

		// 步骤2：在新机器上创建ID管理器，手动设置ID
		newManager := NewAgentIDManager("/new/path/agent.id")
		newManager.info = &AgentIDInfo{
			AgentID:       oldInfo.AgentID, // 使用旧的ID
			RegisteredAt:  time.Now(),
			ServerAddress: oldInfo.ServerAddress,
			Hostname:      "new-hostname",
			LastUpdateAt:  time.Now(),
		}

		// 步骤3：保存到新位置
		if err := newManager.saveToFile(); err != nil {
			fmt.Printf("保存失败: %v\n", err)
		} else {
			fmt.Println("Agent ID迁移成功")
		}
	}
}
