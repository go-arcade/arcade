package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"

	agentv1 "github.com/go-arcade/arcade/api/agent/v1"
	"github.com/go-arcade/arcade/internal/agent/config"
	grpcclient "github.com/go-arcade/arcade/internal/pkg/grpc"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/go-arcade/arcade/pkg/version"
	"github.com/spf13/cobra"
)

// RegisterCmd creates and returns the register command
func RegisterCmd() *cobra.Command {
	registerCmd := &cobra.Command{
		Use:   "register",
		Short: "Register Agent to Server",
		Long:  "Register Agent to Server using the provided token and url, the configuration will be written back to the configuration file after successful registration",
		Run: func(cmd *cobra.Command, args []string) {
			token, _ := cmd.Flags().GetString("token")
			url, _ := cmd.Flags().GetString("url")

			var err error

			if token == "" {
				_, err = fmt.Fprintf(os.Stderr, "error: token parameter is required\n")
				if err != nil {
					return
				}
				err = cmd.Usage()
				if err != nil {
					return
				}
				os.Exit(1)
			}

			if url == "" {
				_, err = fmt.Fprintf(os.Stderr, "error: url parameter is required\n")
				if err != nil {
					return
				}
				err = cmd.Usage()
				if err != nil {
					return
				}
				os.Exit(1)
			}

			if err = registerAgent(configFile, token, url); err != nil {
				_, err = fmt.Fprintf(os.Stderr, "register failed: %v\n", err)
				if err != nil {
					return
				}
				os.Exit(1)
			}

			fmt.Println("register success! configuration has been saved to the configuration file")
		},
	}

	registerCmd.Flags().StringP("token", "t", "", "Server generated token (required)")
	registerCmd.Flags().StringP("url", "u", "", "Server address, e.g. localhost:9090 (required)")
	err := registerCmd.MarkFlagRequired("token")
	if err != nil {
		return nil
	}
	err = registerCmd.MarkFlagRequired("url")
	if err != nil {
		return nil
	}

	return registerCmd
}

// registerAgent register Agent to Server
func registerAgent(configFile, token, serverURL string) error {
	serverAddr := normalizeServerAddr(serverURL)
	if serverAddr == "" {
		return fmt.Errorf("invalid server address: %s", serverURL)
	}

	// get system information
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	localIP := getLocalIP()
	osName := runtime.GOOS
	arch := runtime.GOARCH
	agentVersion := version.Version

	// connect to gRPC Server using internal/pkg/grpc client
	grpcClient, err := grpcclient.NewGrpcClient(grpcclient.ClientConf{
		ServerAddr:       serverAddr,
		Token:            token,
		ReadWriteTimeout: 10, // 10 seconds timeout
		MaxMsgSize:       0,  // use default
	})
	if err != nil {
		return fmt.Errorf("connect to Server failed: %w", err)
	}
	defer func(grpcClient *grpcclient.ClientWrapper) {
		err = grpcClient.Close()
		if err != nil {
			log.Errorw("failed to close gRPC client", "error", err)
		}
	}(grpcClient)

	// create authenticated context with timeout
	ctx, cancel := grpcClient.WithTimeoutAndAuth(context.Background())
	defer cancel()

	// call Register API (do not pass agent_id, let the server generate)
	client := grpcClient.AgentClient
	req := &agentv1.RegisterRequest{
		Hostname:          hostname,
		Ip:                localIP,
		Os:                osName,
		Arch:              arch,
		Version:           agentVersion,
		MaxConcurrentJobs: 10,                      // default value, can be read from configuration
		Labels:            make(map[string]string), // do not pass labels, let the server allocate
		InstalledPlugins:  []string{},              // can be read from configuration or runtime
	}

	resp, err := client.Register(ctx, req)
	if err != nil {
		return fmt.Errorf("register request failed: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("register rejected: %s", resp.Message)
	}

	// write back configuration file (including Agent ID, heartbeat interval and labels)
	if err = config.UpdateConfigFile(configFile, serverAddr, token, resp.AgentId, int(resp.HeartbeatInterval), resp.Labels); err != nil {
		return fmt.Errorf("update configuration file failed: %w", err)
	}
	fmt.Printf("register agent success: agent_id: %s, server_addr: %s, heartbeat_interval: %d, labels: %v\n", resp.AgentId, serverAddr, resp.HeartbeatInterval, resp.Labels)
	return nil
}

// normalizeServerAddr normalize server address
func normalizeServerAddr(url string) string {
	url = strings.TrimSpace(url)
	if url == "" {
		return ""
	}

	// if contains ://, try to parse as URL
	if strings.Contains(url, "://") {
		if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
			url = strings.TrimPrefix(url, "http://")
			url = strings.TrimPrefix(url, "https://")
		}
	}

	if strings.Contains(url, ":") {
		return url
	}

	return fmt.Sprintf("%s:9090", url)
}

// getLocalIP get local IP address
func getLocalIP() string {
	adders, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}

	for _, addr := range adders {
		if inet, ok := addr.(*net.IPNet); ok && !inet.IP.IsLoopback() {
			if inet.IP.To4() != nil {
				return inet.IP.String()
			}
		}
	}

	return "127.0.0.1"
}
