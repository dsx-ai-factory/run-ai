/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package protocol_mapper

import (
	"testing"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/utils"
	"github.com/stretchr/testify/assert"
)

func TestNewAccessKeyGroupsProtocolMapper(t *testing.T) {
	t.Parallel()
	tenantId := "test-tenant-123"
	mapper := NewAccessKeyGroupsProtocolMapper(tenantId)

	assert.Equal(t, "Access Key Groups Mapper", *mapper.Name)
	assert.Equal(t, "runai-access-key-attribute-mapper", *mapper.ProtocolMapper)
	assert.Equal(t, utils.AuthProtocolOidc, *mapper.Protocol)
	assert.NotNil(t, mapper.Config)
	assert.Equal(t, "groups", (*mapper.Config)["claim.name"])
	assert.Equal(t, "GROUPS", (*mapper.Config)["groups.user.attribute"])
	assert.Equal(t, tenantId, (*mapper.Config)["tenant.id"])
	assert.Equal(t, "true", (*mapper.Config)["id.token.claim"])
	assert.Equal(t, "true", (*mapper.Config)["access.token.claim"])
	assert.Equal(t, "true", (*mapper.Config)["userinfo.token.claim"])
	assert.Equal(t, "true", (*mapper.Config)["lightweight.claim"])
	assert.Equal(t, "false", (*mapper.Config)["introspection.token.claim"])
	assert.Equal(t, "true", (*mapper.Config)["multivalued"])
	assert.Equal(t, "true", (*mapper.Config)["groups.intersection.enabled"])
	assert.False(t, *mapper.ConsentRequired)
}

func TestNewAccessKeyUIDProtocolMapper(t *testing.T) {
	t.Parallel()
	tenantId := "test-tenant-123"
	mapper := NewAccessKeyUIDProtocolMapper(tenantId)

	assert.Equal(t, "Access Key UID Mapper", *mapper.Name)
	assert.Equal(t, "runai-access-key-attribute-mapper", *mapper.ProtocolMapper)
	assert.Equal(t, utils.AuthProtocolOidc, *mapper.Protocol)
	assert.NotNil(t, mapper.Config)
	assert.Equal(t, "uid", (*mapper.Config)["claim.name"])
	assert.Equal(t, "UID", (*mapper.Config)["groups.user.attribute"])
	assert.Equal(t, tenantId, (*mapper.Config)["tenant.id"])
	assert.Equal(t, "false", (*mapper.Config)["multivalued"])
	assert.Equal(t, "false", (*mapper.Config)["groups.intersection.enabled"])
	assert.False(t, *mapper.ConsentRequired)
}

func TestNewAccessKeyGIDProtocolMapper(t *testing.T) {
	t.Parallel()
	tenantId := "test-tenant-123"
	mapper := NewAccessKeyGIDProtocolMapper(tenantId)

	assert.Equal(t, "Access Key GID Mapper", *mapper.Name)
	assert.Equal(t, "runai-access-key-attribute-mapper", *mapper.ProtocolMapper)
	assert.Equal(t, utils.AuthProtocolOidc, *mapper.Protocol)
	assert.NotNil(t, mapper.Config)
	assert.Equal(t, "gid", (*mapper.Config)["claim.name"])
	assert.Equal(t, "GID", (*mapper.Config)["groups.user.attribute"])
	assert.Equal(t, tenantId, (*mapper.Config)["tenant.id"])
	assert.Equal(t, "false", (*mapper.Config)["multivalued"])
	assert.Equal(t, "false", (*mapper.Config)["groups.intersection.enabled"])
	assert.False(t, *mapper.ConsentRequired)
}

func TestNewAccessKeySupplementaryGroupsProtocolMapper(t *testing.T) {
	t.Parallel()
	tenantId := "test-tenant-123"
	mapper := NewAccessKeySupplementaryGroupsProtocolMapper(tenantId)

	assert.Equal(t, "Access Key SupplementaryGroups Mapper", *mapper.Name)
	assert.Equal(t, "runai-access-key-attribute-mapper", *mapper.ProtocolMapper)
	assert.Equal(t, utils.AuthProtocolOidc, *mapper.Protocol)
	assert.NotNil(t, mapper.Config)
	assert.Equal(t, "supplementarygroups", (*mapper.Config)["claim.name"])
	assert.Equal(t, "SUPPLEMENTARYGROUPS", (*mapper.Config)["groups.user.attribute"])
	assert.Equal(t, tenantId, (*mapper.Config)["tenant.id"])
	assert.Equal(t, "true", (*mapper.Config)["multivalued"])
	assert.Equal(t, "false", (*mapper.Config)["groups.intersection.enabled"])
	assert.False(t, *mapper.ConsentRequired)
}
