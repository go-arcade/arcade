package model

import (
	"github.com/observabil/arcade/pkg/datatype"
	"golang.org/x/oauth2"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/11/6 19:33
 * @file: model_oauth_provider.go
 * @description: model oauth provider
 */

type OauthProvider struct {
	BaseModel
	Name    string        `gorm:"column:name" json:"name"`
	Content datatype.JSON `gorm:"column:content" json:"content"`
}

func (s *OauthProvider) TableName() string {
	return "t_oauth_provider"
}

type OauthProviderContent struct {
	ClientID     string          `json:"clientId"`
	ClientSecret string          `json:"clientSecret"`
	RedirectURL  string          `json:"redirectURL"`
	Scopes       []string        `json:"scopes"`
	Endpoint     oauth2.Endpoint `json:"endpoint"`
}
