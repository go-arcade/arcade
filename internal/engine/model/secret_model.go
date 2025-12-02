package model

// Secret secret model
type Secret struct {
	BaseModel
	SecretId    string `gorm:"column:secret_id" json:"secretId"`
	Name        string `gorm:"column:name" json:"name"`
	SecretType  string `gorm:"column:secret_type" json:"secretType"`
	SecretValue string `gorm:"column:secret_value" json:"secretValue"`
	Description string `gorm:"column:description" json:"description"`
	Scope       string `gorm:"column:scope" json:"scope"`
	ScopeId     string `gorm:"column:scope_id" json:"scopeId"`
	CreatedBy   string `gorm:"column:created_by" json:"createdBy"`
}

func (Secret) TableName() string {
	return "t_secret"
}
