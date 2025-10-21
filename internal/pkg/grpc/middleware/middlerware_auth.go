package middleware

import (
	"context"
	"errors"

	grpcauth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TokenInfo struct {
	ID    string
	Roles []string
}

// 需要跳过认证的 gRPC 方法
var excludedAuthMethods = map[string]bool{
	"/api.agent.v1.Agent/Heartbeat":  true,
	"/api.job.v1.Job/Ping":           true,
	"/api.stream.v1.Stream/Ping":     true,
	"/api.pipeline.v1.Pipeline/Ping": true,
}

// AuthUnaryInterceptor 一元调用认证拦截器（可跳过心跳接口）
func AuthUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		// 检查是否需要跳过认证
		if excludedAuthMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		// 执行认证
		newCtx, err := AuthInterceptor(ctx)
		if err != nil {
			return nil, err
		}

		return handler(newCtx, req)
	}
}

// AuthStreamInterceptor 流式调用认证拦截器（可跳过心跳接口）
func AuthStreamInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// 检查是否需要跳过认证
		if excludedAuthMethods[info.FullMethod] {
			return handler(srv, ss)
		}

		// 执行认证
		newCtx, err := AuthInterceptor(ss.Context())
		if err != nil {
			return err
		}

		// 创建新的 ServerStream 包装器
		wrapped := &wrappedServerStream{
			ServerStream: ss,
			ctx:          newCtx,
		}

		return handler(srv, wrapped)
	}
}

// wrappedServerStream 包装 ServerStream 以替换 context
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

func AuthInterceptor(ctx context.Context) (context.Context, error) {
	token, err := grpcauth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, err
	}
	tokenInfo, err := parseToken(token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, " %v", err)
	}
	//使用context.WithValue添加了值后，可以用Value(key)方法获取值
	newCtx := context.WithValue(ctx, tokenInfo.ID, tokenInfo)
	//log.Println(newCtx.Value(tokenInfo.ID))
	return newCtx, nil
}

// 解析token，并进行验证
func parseToken(token string) (TokenInfo, error) {
	var tokenInfo TokenInfo
	if token == "grpc.auth.token" {
		tokenInfo.ID = "1"
		tokenInfo.Roles = []string{"admin"}
		return tokenInfo, nil
	}
	return tokenInfo, errors.New("Token无效: bearer " + token)
}

// 从token中获取用户唯一标识
func userClaimFromToken(tokenInfo TokenInfo) string {
	return tokenInfo.ID
}
