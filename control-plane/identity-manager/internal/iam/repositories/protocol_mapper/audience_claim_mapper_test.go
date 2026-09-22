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

func TestNewRunaiAudienceProtocolMapper(t *testing.T) {
	t.Parallel()
	mapper := NewRunaiAudienceProtocolMapper()

	assert.Equal(t, "Audience Mapper", *mapper.Name)
	assert.Equal(t, "oidc-audience-mapper", *mapper.ProtocolMapper)
	assert.Equal(t, utils.AuthProtocolOidc, *mapper.Protocol)
	assert.NotNil(t, mapper.Config)
	assert.Equal(t, "runai", (*mapper.Config)["included.custom.audience"])
	assert.Equal(t, "true", (*mapper.Config)["id.token.claim"])
	assert.Equal(t, "true", (*mapper.Config)["access.token.claim"])
	assert.False(t, *mapper.ConsentRequired)
}

func TestNewRunaiDummyAudienceProtocolMapper(t *testing.T) {
	t.Parallel()
	mapper := NewRunaiDummyAudienceProtocolMapper()

	assert.Equal(t, "Dummy Audience Mapper", *mapper.Name)
	assert.Equal(t, "oidc-audience-mapper", *mapper.ProtocolMapper)
	assert.Equal(t, utils.AuthProtocolOidc, *mapper.Protocol)
	assert.NotNil(t, mapper.Config)
	assert.Equal(t, "runai-dummy", (*mapper.Config)["included.custom.audience"])
	assert.Equal(t, "true", (*mapper.Config)["id.token.claim"])
	assert.Equal(t, "true", (*mapper.Config)["access.token.claim"])
	assert.False(t, *mapper.ConsentRequired)
}
