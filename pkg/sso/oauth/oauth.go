package oauth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-resty/resty/v2"
	httpx "github.com/observabil/arcade/pkg/http"
	"golang.org/x/oauth2"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/11/6 19:58
 * @file: oauth.go
 * @description: oauth2.0
 */

type Provider string

const (
	ProviderGitHub Provider = "github"
	ProviderGoogle Provider = "google"
	ProviderSlack  Provider = "slack"
)

type UserInfo struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	Picture   string `json:"picture"`
	Nickname  string `json:"nickname"`
}

type GitHubUserInfo struct {
	Id        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

// GetUserInfoFunc 获取用户信息函数
// 所有oauth全部在这里定义
var GetUserInfoFunc = map[Provider]func(token *oauth2.Token) (*UserInfo, error){
	ProviderGitHub: getGitHubUserInfo,
	ProviderGoogle: getGoogleUserInfo,
	ProviderSlack:  getSlackUserInfo,
}

// StateStore 用于存储状态参数
var StateStore = &sync.Map{}

var (
	githubUserInfoURL = "https://api.github.com/user"
	googleUserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	slackUserInfoURL  = "https://slack.com/api/users.identity"
)

// GenState 生成随机状态字符串
func GenState() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

// LoadAndDeleteState 从 stateStore 中加载并删除 state
func LoadAndDeleteState(state string) (string, bool) {
	value, ok := StateStore.Load(state)
	if ok {
		StateStore.Delete(state)
		return value.(string), true
	}
	return "", false
}

// getGitHubUserInfo 获取GitHub用户信息
func getGitHubUserInfo(token *oauth2.Token) (*UserInfo, error) {
	client := resty.New()
	client.SetAuthToken(token.AccessToken)

	var ghUserInfo GitHubUserInfo
	resp, err := client.R().
		SetHeader("Accept", "application/vnd.github.v3+json").
		SetResult(&ghUserInfo).
		Get(githubUserInfoURL)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("%s, code: %d, response: %s", httpx.Failed.Msg, resp.StatusCode(), resp.String())
	}

	userInfo := &UserInfo{
		ID:        fmt.Sprintf("%d", ghUserInfo.Id),
		Nickname:  ghUserInfo.Name,
		Username:  ghUserInfo.Login,
		AvatarURL: ghUserInfo.AvatarURL,
	}

	return userInfo, nil
}

// getGoogleUserInfo 获取Google用户信息
func getGoogleUserInfo(token *oauth2.Token) (*UserInfo, error) {
	client := resty.New()
	client.SetAuthToken(token.AccessToken)

	var userInfo UserInfo
	resp, err := client.R().
		SetResult(&userInfo).
		Get(googleUserInfoURL)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("request faild, code: %d, response: %s", resp.StatusCode(), resp.String())
	}

	return &userInfo, nil
}

// getSlackUserInfo 获取Slack用户信息
func getSlackUserInfo(token *oauth2.Token) (*UserInfo, error) {
	client := resty.New()
	client.SetAuthToken(token.AccessToken)

	var userInfo UserInfo
	resp, err := client.R().
		SetResult(&userInfo).
		Get(slackUserInfoURL)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("request faild, code: %d, response: %s", resp.StatusCode(), resp.String())
	}

	return &userInfo, nil
}
