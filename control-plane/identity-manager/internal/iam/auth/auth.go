/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	"context"
)

type LoginType string

const (
	LoginTypeUser           LoginType = "user"
	LoginTypeServiceAccount LoginType = "app"
)

type Configuration struct {
	Host     string
	Username string
	Password string
}

type Auth interface {
	GetAdminToken(context.Context) (*string, error)
	GetHost() string
	SetLoginType(context.Context, LoginType) error
}
