/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package protocol_mapper

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/utils"
	"github.com/stretchr/testify/assert"
)

func TestNewRunaiGroupsProtocolMapper(t *testing.T) {
	t.Parallel()
	mapper := NewRunaiGroupsProtocolMapper(true, "tenantId")
	assert.Equal(t, "Runai Groups Mapper", *mapper.Name)
	assert.Equal(t, "runai-oidc-custom-groups-mapper", *mapper.ProtocolMapper)
	assert.Equal(t, utils.AuthProtocolOidc, *mapper.Protocol)
	assert.NotNil(t, mapper.Config)
	assert.Equal(t, "groups", (*mapper.Config)["claim.name"])
	assert.Equal(t, "GROUPS", (*mapper.Config)["groups.user.attribute"])
	assert.Equal(t, "tenantId", (*mapper.Config)["tenant.id"])
	assert.Equal(t, "true", (*mapper.Config)["id.token.claim"])
	assert.Equal(t, "true", (*mapper.Config)["access.token.claim"])
	assert.Equal(t, "true", (*mapper.Config)["userinfo.token.claim"])
	assert.Equal(t, "true", (*mapper.Config)["lightweight.claim"])
	assert.Equal(t, "true", (*mapper.Config)["groups.intersection.enabled"])
	assert.Equal(t, "false", (*mapper.Config)["introspection.token.claim"])
}

func TestNewRunaiGroupsProtocolMapper_WithFalseIdTokenGroupsClaim(t *testing.T) {
	t.Parallel()
	mapper := NewRunaiGroupsProtocolMapper(false, "tenantId")
	assert.Equal(t, "Runai Groups Mapper", *mapper.Name)
	assert.Equal(t, "runai-oidc-custom-groups-mapper", *mapper.ProtocolMapper)
	assert.Equal(t, utils.AuthProtocolOidc, *mapper.Protocol)
	assert.NotNil(t, mapper.Config)
	assert.Equal(t, "groups", (*mapper.Config)["claim.name"])
	assert.Equal(t, "GROUPS", (*mapper.Config)["groups.user.attribute"])
	assert.Equal(t, "tenantId", (*mapper.Config)["tenant.id"])
	assert.Equal(t, "false", (*mapper.Config)["id.token.claim"])
	assert.Equal(t, "true", (*mapper.Config)["access.token.claim"])
	assert.Equal(t, "true", (*mapper.Config)["userinfo.token.claim"])
	assert.Equal(t, "true", (*mapper.Config)["lightweight.claim"])
	assert.Equal(t, "true", (*mapper.Config)["groups.intersection.enabled"])
	assert.Equal(t, "false", (*mapper.Config)["introspection.token.claim"])
}
