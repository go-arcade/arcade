// Copyright 2025 Arcade Team
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"fmt"
	"slices"
	"time"

	"github.com/go-arcade/arcade/internal/engine/model"
	userrepo "github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/pkg/log"
)

type UserExtensionService struct {
	userExtRepo userrepo.IUserExtRepository
}

func NewUserExtensionService(userExtRepo userrepo.IUserExtRepository) *UserExtensionService {
	return &UserExtensionService{
		userExtRepo: userExtRepo,
	}
}

// GetUserExtension gets user extension information
func (ues *UserExtensionService) GetUserExtension(userId string) (*model.UserExt, error) {
	extension, err := ues.userExtRepo.GetByUserId(userId)
	if err != nil {
		log.Errorw("failed to get user extension", "userId", userId, "error", err)
		return nil, err
	}
	return extension, nil
}

// CreateUserExtension creates user extension record
func (ues *UserExtensionService) CreateUserExtension(extension *model.UserExt) error {
	// check if already exists
	exists, err := ues.userExtRepo.Exists(extension.UserId)
	if err != nil {
		log.Errorw("failed to check user extension exists", "userId", extension.UserId, "error", err)
		return err
	}
	if exists {
		return fmt.Errorf("user extension already exists for user: %s", extension.UserId)
	}

	if err := ues.userExtRepo.Create(extension); err != nil {
		log.Errorw("failed to create user extension", "userId", extension.UserId, "error", err)
		return err
	}

	return nil
}

// UpdateUserExtension updates user extension information
func (ues *UserExtensionService) UpdateUserExtension(userId string, extension *model.UserExt) error {
	// check if exists
	exists, err := ues.userExtRepo.Exists(userId)
	if err != nil {
		log.Errorw("failed to check user extension exists", "userId", userId, "error", err)
		return err
	}
	if !exists {
		return fmt.Errorf("user extension not found for user: %s", userId)
	}

	if err := ues.userExtRepo.Update(userId, extension); err != nil {
		log.Errorw("failed to update user extension", "userId", userId, "error", err)
		return err
	}

	return nil
}

// UpdateLastLogin updates user's last login timestamp
func (ues *UserExtensionService) UpdateLastLogin(userId string) error {
	// create extension record if not exists
	exists, err := ues.userExtRepo.Exists(userId)
	if err != nil {
		log.Errorw("failed to check user extension exists", "userId", userId, "error", err)
		return err
	}

	if !exists {
		// auto-create extension record with default values
		now := time.Now()
		extension := &model.UserExt{
			UserId:           userId,
			Timezone:         "UTC",
			LastLoginAt:      &now,
			InvitationStatus: model.UserInvitationStatusAccepted,
		}
		if err := ues.userExtRepo.Create(extension); err != nil {
			log.Errorw("failed to create user extension", "userId", userId, "error", err)
			return err
		}
		return nil
	}

	if err := ues.userExtRepo.UpdateLastLogin(userId); err != nil {
		log.Errorw("failed to update last login", "userId", userId, "error", err)
		return err
	}

	return nil
}

// UpdateTimezone updates user timezone
func (ues *UserExtensionService) UpdateTimezone(userId, timezone string) error {
	if err := ues.userExtRepo.UpdateTimezone(userId, timezone); err != nil {
		log.Errorw("failed to update timezone", "userId", userId, "timezone", timezone, "error", err)
		return err
	}
	return nil
}

// UpdateInvitationStatus updates invitation status
func (ues *UserExtensionService) UpdateInvitationStatus(userId, status string) error {
	// validate status
	validStatuses := []string{
		model.UserInvitationStatusPending,
		model.UserInvitationStatusAccepted,
		model.UserInvitationStatusExpired,
		model.UserInvitationStatusRevoked,
	}

	isValid := slices.Contains(validStatuses, status)
	if !isValid {
		return fmt.Errorf("invalid invitation status: %s", status)
	}

	if err := ues.userExtRepo.UpdateInvitationStatus(userId, status); err != nil {
		log.Errorw("failed to update invitation status", "userId", userId, "status", status, "error", err)
		return err
	}

	return nil
}
