package repo

import (
	"time"

	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/pkg/database"
)

type IUserExtensionRepository interface {
	GetByUserId(userId string) (*model.UserExtension, error)
	Create(extension *model.UserExtension) error
	Update(userId string, extension *model.UserExtension) error
	UpdateLastLogin(userId string) error
	UpdateTimezone(userId, timezone string) error
	UpdateInvitationStatus(userId, status string) error
	Delete(userId string) error
	Exists(userId string) (bool, error)
}

type UserExtensionRepo struct {
	database.IDatabase
}

func NewUserExtensionRepo(db database.IDatabase) IUserExtensionRepository {
	return &UserExtensionRepo{
		IDatabase: db,
	}
}

// GetByUserId gets user extension by user ID
func (uer *UserExtensionRepo) GetByUserId(userId string) (*model.UserExtension, error) {
	var extension model.UserExtension
	err := uer.Database().Table(extension.TableName()).
		Where("user_id = ?", userId).
		First(&extension).Error
	return &extension, err
}

// Create creates a user extension record
func (uer *UserExtensionRepo) Create(extension *model.UserExtension) error {
	return uer.Database().Table(extension.TableName()).Create(extension).Error
}

// Update updates user extension information
func (uer *UserExtensionRepo) Update(userId string, extension *model.UserExtension) error {
	return uer.Database().Table(extension.TableName()).
		Where("user_id = ?", userId).
		Updates(extension).Error
}

// UpdateLastLogin updates the last login timestamp
func (uer *UserExtensionRepo) UpdateLastLogin(userId string) error {
	now := time.Now()
	var extension model.UserExtension
	return uer.Database().Table(extension.TableName()).
		Where("user_id = ?", userId).
		Update("last_login_at", now).Error
}

// UpdateTimezone updates user timezone
func (uer *UserExtensionRepo) UpdateTimezone(userId, timezone string) error {
	var extension model.UserExtension
	return uer.Database().Table(extension.TableName()).
		Where("user_id = ?", userId).
		Update("timezone", timezone).Error
}

// UpdateInvitationStatus updates invitation status
func (uer *UserExtensionRepo) UpdateInvitationStatus(userId, status string) error {
	updates := map[string]interface{}{
		"invitation_status": status,
	}

	// if status is accepted, set accepted_at timestamp
	if status == model.UserInvitationStatusAccepted {
		updates["accepted_at"] = time.Now()
	}

	var extension model.UserExtension
	return uer.Database().Table(extension.TableName()).
		Where("user_id = ?", userId).
		Updates(updates).Error
}

// Delete deletes user extension record
func (uer *UserExtensionRepo) Delete(userId string) error {
	var extension model.UserExtension
	return uer.Database().Table(extension.TableName()).
		Where("user_id = ?", userId).
		Delete(&model.UserExtension{}).Error
}

// Exists checks if user extension exists
func (uer *UserExtensionRepo) Exists(userId string) (bool, error) {
	var count int64
	var extension model.UserExtension
	err := uer.Database().Table(extension.TableName()).
		Where("user_id = ?", userId).
		Count(&count).Error
	return count > 0, err
}
