package model

import (
	"golang.org/x/oauth2"
	"gorm.io/datatypes"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/11/6 19:33
 * @file: model_oauth_provider.go
 * @description: model oauth provider
 */

type OauthProvider struct {
	BaseModel
	Name    string         `gorm:"column:name" json:"name"`
	Content datatypes.JSON `gorm:"column:content" json:"content"`
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
