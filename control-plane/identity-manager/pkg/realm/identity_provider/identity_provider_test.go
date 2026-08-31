/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package identity_provider

import (
	"context"
	"testing"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestManager_SamlIds(t *testing.T) {
	entityId := "default/runai"

	sip := NewSamlIdentityProvider(entityId)
	req := api.IdpCreationRequest{
		Type: "saml",
		SamlData: &api.SamlCreationData{
			MetadataXmlUrl: "https://www.example.com/",
		},
	}

	data, err := models.GenerateIdpData("test", req, false)
	assert.Nil(t, err)
	rep := sip.GetRepresentation(context.Background(), data, &map[string]string{})
	assert.Equal(t, entityId, (*rep.Config)["entityId"])
}

func TestManager_SamlNewEntityId(t *testing.T) {
	entityId := "default/runai"
	newentityId := "newentityId"

	sip := NewSamlIdentityProvider(entityId)
	req := api.IdpCreationRequest{
		Type: "saml",
		SamlData: &api.SamlCreationData{
			MetadataXmlUrl: "https://www.example.com/",
			EntityId:       newentityId,
		},
	}

	data, err := models.GenerateIdpData("test", req, false)
	assert.Nil(t, err)
	rep := sip.GetRepresentation(context.Background(), data, &map[string]string{})
	assert.Equal(t, newentityId, (*rep.Config)["entityId"])
}
