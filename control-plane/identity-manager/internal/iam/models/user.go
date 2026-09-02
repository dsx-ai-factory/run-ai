/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import "time"

type RequiredAction string

const (
	RequiredActionVerifyProfile RequiredAction = "VERIFY_PROFILE"
)

type User struct {
	Username      *string
	Email         *string
	ID            *string
	Enabled       *bool
	ResetPassword *bool
	CreatedBy     *string
	CreatedAt     *time.Time
	UpdatedAt     *time.Time
	LastLogin     *time.Time
	TempPassword  *string
	IsLocal       *bool
	Groups        *[]string
	FirstName     *string
	LastName      *string
}

type UserFilters struct {
	Username       *string
	UsernameExact  *bool
	Q              *string
	Max            *int
	Offset         *int
	SortBy         *string
	SortOrder      *string
	CreationFilter *string
	LoginFilter    *string
	UpdateFilter   *string
	CreatedBy      *string
	Search         *string
	FullName       *string
	MinimalSearch  *bool
}

type RequiredUserAction struct {
	Alias         *string
	Config        *map[string]string
	DefaultAction *bool
	Enabled       *bool
	Name          *string
	Priority      *int32
	ProviderID    *string
}
