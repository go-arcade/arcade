package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

type UserInfo struct {
	Username  string
	Email     string
	Nickname  string
	AvatarURL string
}

type OAuthProvider struct {
	Config      *oauth2.Config
	UserInfoURL string
}

func NewOAuthProvider(clientID, clientSecret, redirectURL string, scopes []string, endpoint oauth2.Endpoint, userInfoURL string) *OAuthProvider {
	return &OAuthProvider{
		Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       scopes,
			Endpoint:     endpoint,
		},
		UserInfoURL: userInfoURL,
	}
}

func (p *OAuthProvider) GetAuthURL(state string) string {
	return p.Config.AuthCodeURL(state)
}

func (p *OAuthProvider) ExchangeToken(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.Config.Exchange(ctx, code)
}

func (p *OAuthProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}
	token.SetAuthHeader(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user info request failed: %s", resp.Status)
	}

	var data struct {
		Login     string `json:"login"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &UserInfo{
		Username:  data.Login,
		Email:     data.Email,
		Nickname:  data.Name,
		AvatarURL: data.AvatarURL,
	}, nil
}
