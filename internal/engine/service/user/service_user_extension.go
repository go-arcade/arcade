package user

import (
	"fmt"
	"slices"
	"time"

	"github.com/go-arcade/arcade/internal/engine/model/user"
	userrepo "github.com/go-arcade/arcade/internal/engine/repo/user"
	"github.com/go-arcade/arcade/pkg/log"
)

type UserExtensionService struct {
	userExtRepo userrepo.IUserExtensionRepository
}

func NewUserExtensionService(userExtRepo userrepo.IUserExtensionRepository) *UserExtensionService {
	return &UserExtensionService{
		userExtRepo: userExtRepo,
	}
}

// GetUserExtension gets user extension information
func (ues *UserExtensionService) GetUserExtension(userId string) (*user.UserExtension, error) {
	extension, err := ues.userExtRepo.GetByUserId(userId)
	if err != nil {
		log.Errorf("failed to get user extension: %v", err)
		return nil, err
	}
	return extension, nil
}

// CreateUserExtension creates user extension record
func (ues *UserExtensionService) CreateUserExtension(extension *user.UserExtension) error {
	// check if already exists
	exists, err := ues.userExtRepo.Exists(extension.UserId)
	if err != nil {
		log.Errorf("failed to check user extension exists: %v", err)
		return err
	}
	if exists {
		return fmt.Errorf("user extension already exists for user: %s", extension.UserId)
	}

	if err := ues.userExtRepo.Create(extension); err != nil {
		log.Errorf("failed to create user extension: %v", err)
		return err
	}

	return nil
}

// UpdateUserExtension updates user extension information
func (ues *UserExtensionService) UpdateUserExtension(userId string, extension *user.UserExtension) error {
	// check if exists
	exists, err := ues.userExtRepo.Exists(userId)
	if err != nil {
		log.Errorf("failed to check user extension exists: %v", err)
		return err
	}
	if !exists {
		return fmt.Errorf("user extension not found for user: %s", userId)
	}

	if err := ues.userExtRepo.Update(userId, extension); err != nil {
		log.Errorf("failed to update user extension: %v", err)
		return err
	}

	return nil
}

// UpdateLastLogin updates user's last login timestamp
func (ues *UserExtensionService) UpdateLastLogin(userId string) error {
	// create extension record if not exists
	exists, err := ues.userExtRepo.Exists(userId)
	if err != nil {
		log.Errorf("failed to check user extension exists: %v", err)
		return err
	}

	if !exists {
		// auto-create extension record with default values
		now := time.Now()
		extension := &user.UserExtension{
			UserId:           userId,
			Timezone:         "UTC",
			LastLoginAt:      &now,
			InvitationStatus: user.UserInvitationStatusAccepted,
		}
		if err := ues.userExtRepo.Create(extension); err != nil {
			log.Errorf("failed to create user extension: %v", err)
			return err
		}
		return nil
	}

	if err := ues.userExtRepo.UpdateLastLogin(userId); err != nil {
		log.Errorf("failed to update last login: %v", err)
		return err
	}

	return nil
}

// UpdateTimezone updates user timezone
func (ues *UserExtensionService) UpdateTimezone(userId, timezone string) error {
	if err := ues.userExtRepo.UpdateTimezone(userId, timezone); err != nil {
		log.Errorf("failed to update timezone: %v", err)
		return err
	}
	return nil
}

// UpdateInvitationStatus updates invitation status
func (ues *UserExtensionService) UpdateInvitationStatus(userId, status string) error {
	// validate status
	validStatuses := []string{
		user.UserInvitationStatusPending,
		user.UserInvitationStatusAccepted,
		user.UserInvitationStatusExpired,
		user.UserInvitationStatusRevoked,
	}

	isValid := slices.Contains(validStatuses, status)
	if !isValid {
		return fmt.Errorf("invalid invitation status: %s", status)
	}

	if err := ues.userExtRepo.UpdateInvitationStatus(userId, status); err != nil {
		log.Errorf("failed to update invitation status: %v", err)
		return err
	}

	return nil
}
