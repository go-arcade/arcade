package oidc

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type UserInfo struct {
	Username  string
	Email     string
	Nickname  string
	AvatarURL string
}

type OIDCProvider struct {
	Provider *oidc.Provider
	Verifier *oidc.IDTokenVerifier
	Config   *oauth2.Config
}

func NewOIDCProvider(issuer, clientID, clientSecret, redirectURL string, scopes []string, skipVerify bool) (*OIDCProvider, error) {
	provider, err := oidc.NewProvider(context.Background(), issuer)
	if err != nil {
		return nil, err
	}

	return &OIDCProvider{
		Provider: provider,
		Verifier: provider.Verifier(&oidc.Config{ClientID: clientID, SkipClientIDCheck: skipVerify}),
		Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       append([]string{oidc.ScopeOpenID, "email", "profile"}, scopes...),
			Endpoint:     provider.Endpoint(),
		},
	}, nil
}

func (p *OIDCProvider) GetAuthURL(state string) string {
	return p.Config.AuthCodeURL(state)
}

func (p *OIDCProvider) ExchangeToken(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.Config.Exchange(ctx, code)
}

func (p *OIDCProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("missing id_token")
	}

	idToken, err := p.Verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("invalid id_token: %w", err)
	}

	var claims struct {
		Sub     string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, err
	}

	return &UserInfo{
		Username:  claims.Sub,
		Email:     claims.Email,
		Nickname:  claims.Name,
		AvatarURL: claims.Picture,
	}, nil
}
