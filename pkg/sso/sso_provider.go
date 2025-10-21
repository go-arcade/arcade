package sso

import (
	"context"

	"golang.org/x/oauth2"
)

type ISSOProvider interface {
	AuthCodeURL(state string) string
	ExchangeToken(ctx context.Context, code string) (*oauth2.Token, error)
	GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfoAdapter, error)
}

type UserInfoAdapter struct {
	ID        string `json:"sub"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Username  string `json:"preferred_username"`
	Picture   string `json:"picture"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}
