package http

import (
	"time"

	"github.com/go-arcade/arcade/pkg/ctx"
)

type Http struct {
	Host            string
	Port            int
	Heartbeat       int64
	AccessLog       bool
	UseFileAssets   bool
	ReadTimeout     int
	WriteTimeout    int
	IdleTimeout     int
	ShutdownTimeout int
	BodyLimit       int // 请求体大小限制（字节），默认 100MB
	TLS             TLS
	Auth            Auth
	Ctx             ctx.Context
}

type TLS struct {
	CertFile string
	KeyFile  string
}

type Auth struct {
	SecretKey     string
	AccessExpire  time.Duration
	RefreshExpire time.Duration
}

func (h *Http) SetDefaults() {
	if h.Host == "" {
		h.Host = "127.0.0.1"
	}
	if h.Port == 0 {
		h.Port = 8080
	}
	if h.Heartbeat == 0 {
		h.Heartbeat = 60
	}
	if h.ReadTimeout == 0 {
		h.ReadTimeout = 60
	}
	if h.WriteTimeout == 0 {
		h.WriteTimeout = 60
	}
	if h.IdleTimeout == 0 {
		h.IdleTimeout = 60
	}
	if h.ShutdownTimeout == 0 {
		h.ShutdownTimeout = 10
	}
	if h.BodyLimit == 0 {
		h.BodyLimit = 100 * 1024 * 1024 // 100MB
	}
	if h.Auth.AccessExpire == 0 {
		h.Auth.AccessExpire = 3600 * time.Minute
	}
	if h.Auth.RefreshExpire == 0 {
		h.Auth.RefreshExpire = 7200 * time.Minute
	}
}
