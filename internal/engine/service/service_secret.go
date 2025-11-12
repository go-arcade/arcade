package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"slices"

	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/pkg/ctx"
	"github.com/go-arcade/arcade/pkg/id"
	"github.com/go-arcade/arcade/pkg/log"
	"gorm.io/gorm"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/01/15
 * @file: service_secret.go
 * @description: secret service
 */

type SecretService struct {
	ctx        *ctx.Context
	secretRepo *repo.SecretRepo
	encryptKey []byte // 32 bytes for AES-256
}

func NewSecretService(ctx *ctx.Context, secretRepo *repo.SecretRepo) *SecretService {
	// TODO: load encryption key from config or environment variable
	// For now, using a default key (should be replaced in production)
	encryptKey := []byte("arcade-secret-encryption-key-32b") // 32 bytes for AES-256

	return &SecretService{
		ctx:        ctx,
		secretRepo: secretRepo,
		encryptKey: encryptKey,
	}
}

// encrypt encrypts plain text using AES-256-GCM
func (ss *SecretService) encrypt(plainText string) (string, error) {
	block, err := aes.NewCipher(ss.encryptKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherText := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// decrypt decrypts cipher text using AES-256-GCM
func (ss *SecretService) decrypt(cipherText string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(ss.encryptKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("invalid cipher text")
	}

	nonce, cipherTextBytes := data[:nonceSize], data[nonceSize:]
	plainText, err := gcm.Open(nil, nonce, cipherTextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}

// CreateSecret creates a new secret
func (ss *SecretService) CreateSecret(secret *model.Secret, createdBy string) error {
	// validate required fields
	if secret.Name == "" {
		return errors.New("name is required")
	}
	if secret.SecretType == "" {
		return errors.New("secretType is required")
	}
	if secret.SecretValue == "" {
		return errors.New("secretValue is required")
	}
	if secret.Scope == "" {
		return errors.New("scope is required")
	}

	// validate secret type
	validTypes := []string{"password", "token", "ssh_key", "env"}
	isValidType := slices.Contains(validTypes, secret.SecretType)
	if !isValidType {
		return fmt.Errorf("invalid secretType, must be one of: %v", validTypes)
	}

	// validate scope
	validScopes := []string{"global", "pipeline", "user", "project", "team"}
	isValidScope := slices.Contains(validScopes, secret.Scope)
	if !isValidScope {
		return fmt.Errorf("invalid scope, must be one of: %v", validScopes)
	}

	// encrypt secret value
	encryptedValue, err := ss.encrypt(secret.SecretValue)
	if err != nil {
		log.Errorf("failed to encrypt secret value: %v", err)
		return errors.New("failed to encrypt secret value")
	}

	// generate secret ID
	secret.SecretId = id.GetUild()
	secret.SecretValue = encryptedValue
	secret.CreatedBy = createdBy

	if err := ss.secretRepo.CreateSecret(secret); err != nil {
		log.Errorf("failed to create secret: %v", err)
		return errors.New("failed to create secret")
	}

	log.Infof("secret created successfully: %s", secret.SecretId)
	return nil
}

// UpdateSecret updates a secret
func (ss *SecretService) UpdateSecret(secretId string, secret *model.Secret) error {
	// check if secret exists
	_, err := ss.secretRepo.GetSecretByID(secretId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("secret not found")
		}
		log.Errorf("failed to get secret: %v", err)
		return errors.New("failed to get secret")
	}

	// if updating secret value, encrypt it
	if secret.SecretValue != "" {
		encryptedValue, err := ss.encrypt(secret.SecretValue)
		if err != nil {
			log.Errorf("failed to encrypt secret value: %v", err)
			return errors.New("failed to encrypt secret value")
		}
		secret.SecretValue = encryptedValue
	}

	secret.SecretId = secretId

	if err := ss.secretRepo.UpdateSecret(secret); err != nil {
		log.Errorf("failed to update secret: %v", err)
		return errors.New("failed to update secret")
	}

	log.Infof("secret updated successfully: %s", secretId)
	return nil
}

// GetSecretByID gets a secret by ID (without decrypting value)
func (ss *SecretService) GetSecretByID(secretId string) (*model.Secret, error) {
	secret, err := ss.secretRepo.GetSecretByID(secretId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("secret not found")
		}
		log.Errorf("failed to get secret: %v", err)
		return nil, errors.New("failed to get secret")
	}

	// mask secret value for security
	secret.SecretValue = "***MASKED***"
	return secret, nil
}

// GetSecretValue gets the decrypted secret value (use with caution)
func (ss *SecretService) GetSecretValue(secretId string) (string, error) {
	encryptedValue, err := ss.secretRepo.GetSecretValue(secretId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("secret not found")
		}
		log.Errorf("failed to get secret value: %v", err)
		return "", errors.New("failed to get secret value")
	}

	// decrypt secret value
	plainValue, err := ss.decrypt(encryptedValue)
	if err != nil {
		log.Errorf("failed to decrypt secret value: %v", err)
		return "", errors.New("failed to decrypt secret value")
	}

	return plainValue, nil
}

// GetSecretList gets secret list with pagination and filters
func (ss *SecretService) GetSecretList(pageNum, pageSize int, secretType, scope, scopeId, createdBy string) ([]*model.Secret, int64, error) {
	// set default pagination
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	secrets, total, err := ss.secretRepo.GetSecretList(pageNum, pageSize, secretType, scope, scopeId, createdBy)
	if err != nil {
		log.Errorf("failed to get secret list: %v", err)
		return nil, 0, errors.New("failed to get secret list")
	}

	// secret values are already excluded in repo layer
	return secrets, total, nil
}

// DeleteSecret deletes a secret
func (ss *SecretService) DeleteSecret(secretId string) error {
	// check if secret exists
	_, err := ss.secretRepo.GetSecretByID(secretId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("secret not found")
		}
		log.Errorf("failed to get secret: %v", err)
		return errors.New("failed to get secret")
	}

	if err := ss.secretRepo.DeleteSecret(secretId); err != nil {
		log.Errorf("failed to delete secret: %v", err)
		return errors.New("failed to delete secret")
	}

	log.Infof("secret deleted successfully: %s", secretId)
	return nil
}

// GetSecretsByScope gets secrets by scope and scope_id
func (ss *SecretService) GetSecretsByScope(scope, scopeId string) ([]*model.Secret, error) {
	secrets, err := ss.secretRepo.GetSecretsByScope(scope, scopeId)
	if err != nil {
		log.Errorf("failed to get secrets by scope: %v", err)
		return nil, errors.New("failed to get secrets by scope")
	}
	return secrets, nil
}
