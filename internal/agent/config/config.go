package config

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/go-arcade/arcade/pkg/metrics"
	"github.com/go-arcade/arcade/pkg/pprof"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

// AgentConfig holds all configuration settings
type AgentConfig struct {
	Agent   AgentInfo             `mapstructure:"agent"`
	Log     log.Conf              `mapstructure:"log"`
	Http    http.Http             `mapstructure:"http"`
	Grpc    GrpcClientConfig      `mapstructure:"grpc"`
	Metrics metrics.MetricsConfig `mapstructure:"metrics"`
	Pprof   pprof.PprofConfig     `mapstructure:"pprof"`
}

// AgentInfo agent information
type AgentInfo struct {
	ID       string            `mapstructure:"id"`
	Name     string            `mapstructure:"name"`
	APIkey   string            `mapstructure:"apiKey"`
	Interval int               `mapstructure:"interval"`
	Labels   map[string]string `mapstructure:"labels"`
}

// GrpcClientConfig gRPC client configuration
type GrpcClientConfig struct {
	ServerAddr           string `mapstructure:"serverAddr"`           // server address (host:port format, e.g., "localhost:9090")
	Token                string `mapstructure:"token"`                // Bearer token for authentication (uses APIkey if not set)
	ReadWriteTimeout     int    `mapstructure:"readWriteTimeout"`     // read write timeout (seconds, default: 30)
	MaxMsgSize           int    `mapstructure:"maxMsgSize"`           // max message size (bytes), 0 means use default value
	MaxReconnectAttempts int    `mapstructure:"maxReconnectAttempts"` // max reconnection attempts, 0 means unlimited (default: 0)
}

var (
	cfg  AgentConfig
	once sync.Once
)

func NewConf(confDir string) AgentConfig {
	once.Do(func() {
		var err error
		cfg, err = loadConfigFile(confDir)
		if err != nil {
			panic(fmt.Sprintf("load config file error: %s", err))
		}
	})
	return cfg
}

// LoadConfigFile load config file
func loadConfigFile(confDir string) (AgentConfig, error) {

	config := viper.New()
	config.SetConfigFile(confDir) //文件名
	// 设置配置键名匹配规则，支持大小写不敏感和自动转换
	config.SetConfigType("toml")
	config.AutomaticEnv()

	if err := config.ReadInConfig(); err != nil {
		return cfg, fmt.Errorf("failed to read configuration file: %v", err)
	}

	config.WatchConfig()
	config.OnConfigChange(func(e fsnotify.Event) {
		log.Infow("The configuration changes, re-analyze the configuration file", "file", e.Name)
		if err := config.Unmarshal(&cfg); err != nil {
			_ = fmt.Errorf("failed to unmarshal configuration file: %v", err)
		}
	})
	if err := config.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("failed to unmarshal configuration file: %v", err)
	}

	// parse ServerAddr to gRPC client config
	if err := cfg.parseServerAddr(); err != nil {
		return cfg, fmt.Errorf("failed to parse server address: %v", err)
	}

	// generate token from apikey if token is not set
	if err := cfg.generateTokenFromAPIkey(); err != nil {
		return cfg, fmt.Errorf("failed to generate token from apikey: %v", err)
	}

	log.Infow("config file loaded",
		"path", confDir,
		"grpc.serverAddr", cfg.Grpc.ServerAddr,
		"agent.id", cfg.Agent.ID,
	)

	return cfg, nil
}

// parseServerAddr parses ServerAddr and sets gRPC client config
func (c *AgentConfig) parseServerAddr() error {
	// set default values
	if c.Grpc.ReadWriteTimeout == 0 {
		c.Grpc.ReadWriteTimeout = 30
	}

	// if ServerAddr is already set, validate and normalize it
	if c.Grpc.ServerAddr != "" {
		// try to parse as URL first
		if parsedURL, err := url.Parse(c.Grpc.ServerAddr); err == nil && parsedURL.Scheme != "" {
			// URL format: https://host:port or https://host
			host := parsedURL.Hostname()
			port := parsedURL.Port()
			if port != "" {
				c.Grpc.ServerAddr = fmt.Sprintf("%s:%s", host, port)
			} else {
				// default to port 9090 if no port specified
				c.Grpc.ServerAddr = fmt.Sprintf("%s:9090", host)
			}
			return nil
		}
		// if already in host:port format, keep it as is
		if strings.Contains(c.Grpc.ServerAddr, ":") {
			return nil
		}
		// if only host, add default port
		c.Grpc.ServerAddr = fmt.Sprintf("%s:9090", c.Grpc.ServerAddr)
		return nil
	}

	// ServerAddr is required
	return fmt.Errorf("serverAddr is required in [grpc] section")
}

// generateTokenFromAPIkey generates a permanent token from apikey if token is not set
func (c *AgentConfig) generateTokenFromAPIkey() error {
	// if token is already set, use it
	if c.Grpc.Token != "" {
		return nil
	}

	// if apikey is not set, return error
	if c.Agent.APIkey == "" {
		return fmt.Errorf("either token or apiKey must be set in configuration")
	}

	// generate permanent token from apikey
	token, err := generatePermanentToken(c.Agent.APIkey, c.Agent.ID)
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	c.Grpc.Token = token
	log.Infow("generated token from apikey",
		"agent.id", c.Agent.ID,
	)
	return nil
}

// generatePermanentToken generates a permanent JWT token from apikey
// The token has no expiration time (expires in 100 years)
func generatePermanentToken(apikey, agentID string) (string, error) {
	// Use HMAC-SHA256 to create a signature key from apikey
	h := hmac.New(sha256.New, []byte(apikey))
	h.Write([]byte("arcade-agent-token"))
	signingKey := h.Sum(nil)

	// Create claims with agent ID and very long expiration (100 years)
	now := time.Now()
	claims := jwt.MapClaims{
		"agent_id": agentID,
		"iss":      "arcade-agent",
		"iat":      now.Unix(),
		"nbf":      now.Unix(),
		// Set expiration to 100 years from now (effectively permanent)
		"exp": now.Add(100 * 365 * 24 * time.Hour).Unix(),
	}

	// Create token with HMAC-SHA256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with the derived key
	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// GenerateTokenFromAPIkey is a public function to generate token from apikey
// This can be used for testing or external token generation
func GenerateTokenFromAPIkey(apikey, agentID string) (string, error) {
	return generatePermanentToken(apikey, agentID)
}

// VerifyToken verifies a token using apikey
func VerifyToken(tokenString, apikey string) (string, error) {
	// Derive signing key from apikey (same as generation)
	h := hmac.New(sha256.New, []byte(apikey))
	h.Write([]byte("arcade-agent-token"))
	signingKey := h.Sum(nil)

	// Parse and verify token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return signingKey, nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	// Extract agent ID from claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	agentID, ok := claims["agent_id"].(string)
	if !ok {
		return "", fmt.Errorf("agent_id not found in token claims")
	}

	return agentID, nil
}

// UpdateConfigFile update configuration file, write serverAddr, token, agentId, heartbeatInterval and labels
func UpdateConfigFile(configFile, serverAddr, token, agentID string, heartbeatInterval int, labels map[string]string) error {
	// use viper to read existing configuration
	v := viper.New()
	v.SetConfigFile(configFile)
	v.SetConfigType("toml")

	// read existing configuration (if file exists)
	if err := v.ReadInConfig(); err != nil {
		// if file does not exist, create a new one
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// create default configuration
			v.Set("grpc.serverAddr", serverAddr)
			v.Set("grpc.token", token)
			v.Set("agent.id", agentID)
			if heartbeatInterval > 0 {
				v.Set("agent.interval", heartbeatInterval)
			}
			if len(labels) > 0 {
				v.Set("agent.labels", labels)
			}
		} else {
			return fmt.Errorf("read configuration file failed: %w", err)
		}
	} else {
		// update configuration values
		v.Set("grpc.serverAddr", serverAddr)
		v.Set("grpc.token", token)
		v.Set("agent.id", agentID)
		if heartbeatInterval > 0 {
			v.Set("agent.interval", heartbeatInterval)
		}
		if len(labels) > 0 {
			v.Set("agent.labels", labels)
		}
	}

	// ensure configuration file directory exists
	configDir := configFile
	if idx := strings.LastIndex(configFile, "/"); idx != -1 {
		configDir = configFile[:idx]
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("create configuration directory failed: %w", err)
		}
	}

	// write back configuration file
	if err := v.WriteConfig(); err != nil {
		// if WriteConfig fails, try SafeWriteConfig
		if err := v.SafeWriteConfigAs(configFile); err != nil {
			return fmt.Errorf("write configuration file failed: %w", err)
		}
	}

	return nil
}
