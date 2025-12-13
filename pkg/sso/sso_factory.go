package sso

import (
	"context"
	"fmt"

	"github.com/go-arcade/arcade/pkg/sso/oauth"
	"github.com/go-arcade/arcade/pkg/sso/oidc"
	"golang.org/x/oauth2"
)

type ProviderConfig struct {
	Type         string // "oauth" or "oidc"
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	Endpoint     oauth2.Endpoint
	FieldMap     map[string]string // field mapping

	// OIDC fields
	Issuer     string
	SkipVerify bool

	// OAuth fields
	UserInfoURL string
}

func NewSSOProvider(conf *ProviderConfig) (ISSOProvider, error) {
	switch conf.Type {
	case "oidc":
		provider, err := oidc.NewOIDCProvider(
			conf.Issuer,
			conf.ClientID,
			conf.ClientSecret,
			conf.RedirectURL,
			conf.Scopes,
			conf.SkipVerify,
		)
		if err != nil {
			return nil, err
		}
		return &oidcAdapter{provider}, nil
	case "oauth":
		return &oauthAdapter{
			provider: oauth.NewOAuthProvider(
				conf.ClientID,
				conf.ClientSecret,
				conf.RedirectURL,
				conf.Scopes,
				conf.Endpoint,
				conf.UserInfoURL,
			),
		}, nil
	default:
		return nil, fmt.Errorf("unknown provider type: %s", conf.Type)
	}
}

// oauthAdapter is a adapter for oauth provider
type oauthAdapter struct{ provider *oauth.OAuthProvider }

// AuthCodeURL get auth url
func (a *oauthAdapter) AuthCodeURL(state string) string {
	return a.provider.GetAuthURL(state)
}

// ExchangeToken exchange token
func (a *oauthAdapter) ExchangeToken(ctx context.Context, code string) (*oauth2.Token, error) {
	return a.provider.ExchangeToken(ctx, code)
}

// GetUserInfo get user info
func (a *oauthAdapter) GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfoAdapter, error) {
	u, err := a.provider.GetUserInfo(ctx, token)
	if err != nil {
		return nil, err
	}
	return &UserInfoAdapter{
		Username:  u.Username,
		Email:     u.Email,
		Nickname:  u.Nickname,
		AvatarURL: u.AvatarURL,
	}, nil
}

// oidcAdapter is a adapter for oidc provider
type oidcAdapter struct{ provider *oidc.OIDCProvider }

// AuthCodeURL get auth url
func (a *oidcAdapter) AuthCodeURL(state string) string {
	return a.provider.GetAuthURL(state)
}

// ExchangeToken exchange token
func (a *oidcAdapter) ExchangeToken(ctx context.Context, code string) (*oauth2.Token, error) {
	return a.provider.ExchangeToken(ctx, code)
}

// GetUserInfo get user info
func (a *oidcAdapter) GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfoAdapter, error) {
	u, err := a.provider.GetUserInfo(ctx, token)
	if err != nil {
		return nil, err
	}
	return &UserInfoAdapter{
		Username:  u.Username,
		Email:     u.Email,
		Nickname:  u.Nickname,
		AvatarURL: u.AvatarURL,
	}, nil
}
