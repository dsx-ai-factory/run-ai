/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/generated/pkg/api"
	iamModel "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
)

func BuildUserResponse(user iamModel.User) api.User {
	u := api.User{
		Id:        *user.ID,
		Username:  *user.Username,
		CreatedBy: *user.CreatedBy,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		LastLogin: user.LastLogin,
		IsLocal:   user.IsLocal,
		Groups:    user.Groups,
	}
	if user.FirstName != nil {
		u.FirstName = *user.FirstName
	}
	if user.LastName != nil {
		u.LastName = *user.LastName
	}
	return u
}

func BuildUserMinimalResponse(user iamModel.User) api.UserMinimal {
	u := api.UserMinimal{
		Id:       *user.ID,
		Username: *user.Username,
	}
	if user.FirstName != nil {
		u.FirstName = user.FirstName
	}
	if user.LastName != nil {
		u.LastName = user.LastName
	}
	return u
}

func BuildUserCreationResponse(userId string, email string, tempPassword string) api.UserPostResponse {
	return api.UserPostResponse{
		Id:           userId,
		Username:     email,
		TempPassword: tempPassword,
	}
}
