package repo

import (
	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/pkg/database"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/01/15
 * @file: repo_secret.go
 * @description: secret repository
 */

type ISecretRepository interface {
	CreateSecret(secret *model.Secret) error
	UpdateSecret(secret *model.Secret) error
	GetSecretByID(secretId string) (*model.Secret, error)
	GetSecretList(pageNum, pageSize int, secretType, scope, scopeId, createdBy string) ([]*model.Secret, int64, error)
	DeleteSecret(secretId string) error
	GetSecretsByScope(scope, scopeId string) ([]*model.Secret, error)
	GetSecretValue(secretId string) (string, error)
}

type SecretRepo struct {
	db          database.DB
	secretModel *model.Secret
}

func NewSecretRepo(db database.DB) ISecretRepository {
	return &SecretRepo{
		db:          db,
		secretModel: &model.Secret{},
	}
}

// CreateSecret creates a new secret
func (sr *SecretRepo) CreateSecret(secret *model.Secret) error {
	return sr.db.DB().Table(sr.secretModel.TableName()).Create(secret).Error
}

// UpdateSecret updates a secret by secret_id
func (sr *SecretRepo) UpdateSecret(secret *model.Secret) error {
	return sr.db.DB().Table(sr.secretModel.TableName()).
		Omit("id", "secret_id", "created_at").
		Where("secret_id = ?", secret.SecretId).
		Updates(secret).Error
}

// GetSecretByID gets a secret by secret_id
func (sr *SecretRepo) GetSecretByID(secretId string) (*model.Secret, error) {
	var secret model.Secret
	err := sr.db.DB().Table(sr.secretModel.TableName()).
		Where("secret_id = ?", secretId).
		First(&secret).Error
	return &secret, err
}

// GetSecretList gets secret list with pagination and filters
func (sr *SecretRepo) GetSecretList(pageNum, pageSize int, secretType, scope, scopeId, createdBy string) ([]*model.Secret, int64, error) {
	var secrets []*model.Secret
	var total int64

	query := sr.db.DB().Table(sr.secretModel.TableName())

	// apply filters
	if secretType != "" {
		query = query.Where("secret_type = ?", secretType)
	}
	if scope != "" {
		query = query.Where("scope = ?", scope)
	}
	if scopeId != "" {
		query = query.Where("scope_id = ?", scopeId)
	}
	if createdBy != "" {
		query = query.Where("created_by = ?", createdBy)
	}

	// get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// get paginated list (don't select secret_value, created_at and updated_at for list view)
	offset := (pageNum - 1) * pageSize
	err := query.Select("id", "secret_id", "name", "secret_type", "description", "scope", "scope_id", "created_by").
		Order("id DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&secrets).Error

	return secrets, total, err
}

// DeleteSecret deletes a secret by secret_id
func (sr *SecretRepo) DeleteSecret(secretId string) error {
	return sr.db.DB().Table(sr.secretModel.TableName()).
		Where("secret_id = ?", secretId).
		Delete(&model.Secret{}).Error
}

// GetSecretsByScope gets secrets by scope and scope_id
func (sr *SecretRepo) GetSecretsByScope(scope, scopeId string) ([]*model.Secret, error) {
	var secrets []*model.Secret
	err := sr.db.DB().Table(sr.secretModel.TableName()).
		Select("id", "secret_id", "name", "secret_type", "description", "scope", "scope_id", "created_by").
		Where("scope = ? AND scope_id = ?", scope, scopeId).
		Find(&secrets).Error
	return secrets, err
}

// GetSecretValue gets the secret value (use with caution, only when needed)
func (sr *SecretRepo) GetSecretValue(secretId string) (string, error) {
	var secret model.Secret
	err := sr.db.DB().Table(sr.secretModel.TableName()).
		Select("secret_value").
		Where("secret_id = ?", secretId).
		First(&secret).Error
	return secret.SecretValue, err
}
