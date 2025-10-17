package model

import (
	"golang.org/x/oauth2"
	"gorm.io/datatypes"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/11/6 19:33
 * @file: model_oauth_provider.go
 * @description: model sso provider (支持 OAuth, LDAP, OIDC, SAML)
 */

// SSOProvider SSO认证提供者表
type SSOProvider struct {
	BaseModel
	ProviderId   string         `gorm:"column:provider_id" json:"providerId"`
	Name         string         `gorm:"column:name" json:"name"`
	ProviderType string         `gorm:"column:provider_type" json:"providerType"` // oauth/ldap/oidc/saml
	Config       datatypes.JSON `gorm:"column:config" json:"config"`
	Description  string         `gorm:"column:description" json:"description"`
	Priority     int            `gorm:"column:priority" json:"priority"`
	IsEnabled    int            `gorm:"column:is_enabled" json:"isEnabled"` // 0:禁用 1:启用
}

func (s *SSOProvider) TableName() string {
	return "t_sso_provider"
}

// OAuthConfig OAuth 配置
type OAuthConfig struct {
	ClientID     string          `json:"clientId"`
	ClientSecret string          `json:"clientSecret"`
	AuthURL      string          `json:"authURL"`
	TokenURL     string          `json:"tokenURL"`
	UserInfoURL  string          `json:"userInfoURL"`
	RedirectURL  string          `json:"redirectURL"`
	Scopes       []string        `json:"scopes"`
	Endpoint     oauth2.Endpoint `json:"endpoint,omitempty"`
}

// LDAPConfig LDAP 配置
type LDAPConfig struct {
	Host         string            `json:"host"`
	Port         int               `json:"port"`
	UseTLS       bool              `json:"useTLS"`
	SkipVerify   bool              `json:"skipVerify"`
	BaseDN       string            `json:"baseDN"`
	BindDN       string            `json:"bindDN"`
	BindPassword string            `json:"bindPassword"`
	UserFilter   string            `json:"userFilter"`  // (uid=%s)
	UserDN       string            `json:"userDN"`      // ou=users,dc=example,dc=com
	GroupFilter  string            `json:"groupFilter"` // (memberUid=%s)
	GroupDN      string            `json:"groupDN"`     // ou=groups,dc=example,dc=com
	Attributes   map[string]string `json:"attributes"`  // username, email, displayName, groups
}

// OIDCConfig OIDC (OpenID Connect) 配置
type OIDCConfig struct {
	Issuer       string   `json:"issuer"`
	ClientID     string   `json:"clientId"`
	ClientSecret string   `json:"clientSecret"`
	RedirectURL  string   `json:"redirectURL"`
	Scopes       []string `json:"scopes"`
	UserInfoURL  string   `json:"userInfoURL,omitempty"`
	SkipVerify   bool     `json:"skipVerify"`
	HostedDomain string   `json:"hostedDomain,omitempty"` // Google Workspace domain
}

// SAMLConfig SAML 配置
type SAMLConfig struct {
	EntityID             string `json:"entityId"`
	SSOURL               string `json:"ssoURL"`
	Certificate          string `json:"certificate"`
	PrivateKey           string `json:"privateKey"`
	MetadataURL          string `json:"metadataURL,omitempty"`
	NameIDFormat         string `json:"nameIDFormat"`
	AssertionConsumerURL string `json:"assertionConsumerURL"`
	SignatureMethod      string `json:"signatureMethod,omitempty"`
}
