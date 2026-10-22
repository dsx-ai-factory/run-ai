/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

type Idp struct {
	Alias                     *string
	ProviderID                *string
	DisplayName               *string
	Enabled                   *bool
	Config                    *map[string]string
	AddReadTokenRoleOnCreate  *bool
	FirstBrokerLoginFlowAlias *string
	InternalID                *string
	LinkOnly                  *bool
	PostBrokerLoginFlowAlias  *string
	SupportTokenExchange      *bool
	ValidateSignature         *bool
}

type IdpMapper struct {
	ID                     *string
	Name                   *string
	IdentityProviderMapper *string
	IdentityProviderAlias  *string
	Config                 *map[string]string
}
