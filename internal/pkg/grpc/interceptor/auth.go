package interceptor

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	grpcauth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TokenInfo struct {
	ID    string
	Roles []string
}

// contextKey custom key type for context.WithValue
type contextKey string

const (
	// tokenInfoKey key for storing TokenInfo in context
	tokenInfoKey contextKey = "tokenInfo"
)

// excluded methods that need to skip authentication
var excludedAuthMethods = map[string]bool{
	"/api.agent.v1.Agent/Heartbeat":  true,
	"/api.job.v1.Job/Ping":           true,
	"/api.stream.v1.Stream/Ping":     true,
	"/api.pipeline.v1.Pipeline/Ping": true,
}

// AuthUnaryInterceptor unary server interceptor (can skip heartbeat interface)
func AuthUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		// check if need to skip authentication
		if excludedAuthMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		// execute authentication
		newCtx, err := AuthInterceptor(ctx)
		if err != nil {
			return nil, err
		}

		return handler(newCtx, req)
	}
}

// AuthStreamInterceptor stream server interceptor (can skip heartbeat interface)
func AuthStreamInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// check if need to skip authentication
		if excludedAuthMethods[info.FullMethod] {
			return handler(srv, ss)
		}

		// execute authentication
		newCtx, err := AuthInterceptor(ss.Context())
		if err != nil {
			return err
		}

		// create new ServerStream wrapper
		wrapped := &wrappedServerStream{
			ServerStream: ss,
			ctx:          newCtx,
		}

		return handler(srv, wrapped)
	}
}

// wrappedServerStream wrap ServerStream to replace context
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
	// use custom type as key, avoid conflict with other package's context key
	newCtx := context.WithValue(ctx, tokenInfoKey, tokenInfo)
	return newCtx, nil
}

// GetTokenInfofromContext get TokenInfo from context
func GetTokenInfofromContext(ctx context.Context) (*TokenInfo, bool) {
	value := ctx.Value(tokenInfoKey)
	if value == nil {
		return nil, false
	}
	tokenInfo, ok := value.(TokenInfo)
	if !ok {
		return nil, false
	}
	return &tokenInfo, true
}

// TokenVerifier interface for token verification
// This allows different implementations (config-based, database-based, etc.)
type TokenVerifier interface {
	// GetAPIkey returns the apikey for the given agent ID
	// Returns empty string if agent not found
	GetAPIkey(agentID string) (string, error)
}

// defaultTokenVerifier default token verifier using hardcoded token
type defaultTokenVerifier struct{}

func (v *defaultTokenVerifier) GetAPIkey(agentID string) (string, error) {
	// Default implementation: return empty to use hardcoded token fallback
	return "", nil
}

var tokenVerifier TokenVerifier = &defaultTokenVerifier{}

// SetTokenVerifier sets the token verifier (can be called during initialization)
func SetTokenVerifier(verifier TokenVerifier) {
	tokenVerifier = verifier
}

// parse token and validate
func parseToken(token string) (TokenInfo, error) {
	var tokenInfo TokenInfo

	// Support legacy hardcoded token
	if token == "grpc.auth.token" {
		tokenInfo.ID = "1"
		tokenInfo.Roles = []string{"admin"}
		return tokenInfo, nil
	}

	// Try to parse as JWT token (generated from apikey)
	agentID, err := verifyJWTToken(token)
	if err == nil && agentID != "" {
		tokenInfo.ID = agentID
		tokenInfo.Roles = []string{"agent"} // Default role for agents
		return tokenInfo, nil
	}

	return tokenInfo, fmt.Errorf("invalid token: %v", err)
}

// verifyJWTToken verifies a JWT token generated from apikey
func verifyJWTToken(tokenString string) (string, error) {
	// First, parse token without verification to extract agent_id
	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	// Extract agent_id from claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	agentID, ok := claims["agent_id"].(string)
	if !ok || agentID == "" {
		return "", fmt.Errorf("agent_id not found in token claims")
	}

	// Get apikey for this agent
	apikey, err := tokenVerifier.GetAPIkey(agentID)
	if err != nil {
		return "", fmt.Errorf("failed to get apikey for agent %s: %w", agentID, err)
	}

	// If apikey is empty, return error
	if apikey == "" {
		return "", fmt.Errorf("apikey not found for agent %s", agentID)
	}

	// Derive signing key from apikey (same as generation)
	h := hmac.New(sha256.New, []byte(apikey))
	h.Write([]byte("arcade-agent-token"))
	signingKey := h.Sum(nil)

	// Now verify the token with the correct signing key
	verifiedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return signingKey, nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to verify token: %w", err)
	}

	if !verifiedToken.Valid {
		return "", fmt.Errorf("invalid token signature")
	}

	return agentID, nil
}
