/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/repositories/users"
)

type UserService interface {
	GetUsers(ctx context.Context, realmName string, filter models.UserFilters) ([]*models.User, error)
	CreateUser(ctx context.Context, realmName string, user *models.User) (string, error)
	GetUserByID(ctx context.Context, realmName string, userID string) (*models.User, error)
	GetUserByEmail(ctx context.Context, realmName string, email string) (*models.User, error)
	DeleteUser(ctx context.Context, realmName string, userID string) error
	LogoutUser(ctx context.Context, realm string, userID string) error
	ChangeUserPassword(ctx context.Context, realm string, user *models.User, currentPassword string, newPassword string) error
	ResetUserPassword(ctx context.Context, realm string, userID string, password string) error
	SendPasswordResetEmail(ctx context.Context, realm string, userID string, expirationSeconds int, redirectURI string, clientId string) error
	ToggleRequiredAction(ctx context.Context, realm string, actionAlias models.RequiredAction, enabled bool) error
	GetRequiredAction(ctx context.Context, realm string, actionAlias string) (*models.RequiredUserAction, error)
	UpdateUser(ctx context.Context, realmName string, userID string, user *models.User) error
	ToggleUnmanagedUsersAttributes(ctx context.Context, realm string, enabled bool) error
	LogoutUserBySessionID(ctx context.Context, realm string, sessionID string) error
}

type userService struct {
	userRepository users.UserRepository
}

func NewUserService(userRepository users.UserRepository) UserService {
	return &userService{
		userRepository: userRepository,
	}
}

func (u *userService) GetUsers(ctx context.Context, realmName string, filter models.UserFilters) ([]*models.User, error) {
	return u.userRepository.GetUsers(ctx, realmName, filter)
}

func (u *userService) CreateUser(ctx context.Context, realmName string, user *models.User) (string, error) {
	return u.userRepository.CreateUser(ctx, realmName, user)
}

func (u *userService) GetUserByID(ctx context.Context, realmName string, userID string) (*models.User, error) {
	return u.userRepository.GetUserByID(ctx, realmName, userID)
}

func (u *userService) GetUserByEmail(ctx context.Context, realmName string, email string) (*models.User, error) {
	return u.userRepository.GetUserByEmail(ctx, realmName, email)
}

func (u *userService) DeleteUser(ctx context.Context, realmName string, userID string) error {
	return u.userRepository.DeleteUser(ctx, realmName, userID)
}

func (u *userService) LogoutUser(ctx context.Context, realm string, userID string) error {
	return u.userRepository.LogoutUser(ctx, realm, userID)
}

func (u *userService) ChangeUserPassword(ctx context.Context, realm string, user *models.User, currentPassword string, newPassword string) error {
	return u.userRepository.ChangeUserPassword(ctx, realm, user, currentPassword, newPassword)
}

func (u *userService) ResetUserPassword(ctx context.Context, realm string, userID string, password string) error {
	return u.userRepository.ResetUserPassword(ctx, realm, userID, password)
}

func (u *userService) SendPasswordResetEmail(ctx context.Context, realm string, userID string, expirationSeconds int, redirectURI string, clientId string) error {
	return u.userRepository.SendPasswordResetEmail(ctx, realm, userID, expirationSeconds, redirectURI, clientId)
}

func (u *userService) ToggleRequiredAction(ctx context.Context, realm string, actionAlias models.RequiredAction, enabled bool) error {
	return u.userRepository.ToggleRequiredAction(ctx, realm, actionAlias, enabled)
}

func (u *userService) UpdateUser(ctx context.Context, realmName string, userID string, user *models.User) error {
	return u.userRepository.UpdateUser(ctx, realmName, userID, user)
}

func (u *userService) ToggleUnmanagedUsersAttributes(ctx context.Context, realm string, enabled bool) error {
	return u.userRepository.ToggleUnmanagedUsersAttributes(ctx, realm, enabled)
}

func (u *userService) LogoutUserBySessionID(ctx context.Context, realm string, sessionID string) error {
	return u.userRepository.LogoutUserBySessionID(ctx, realm, sessionID)
}

func (u *userService) GetRequiredAction(ctx context.Context, realm string, actionAlias string) (*models.RequiredUserAction, error) {
	return u.userRepository.GetRequiredAction(ctx, realm, actionAlias)
}
