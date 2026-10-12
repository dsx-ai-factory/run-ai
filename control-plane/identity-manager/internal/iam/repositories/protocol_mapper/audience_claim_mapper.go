/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package protocol_mapper

import (
	"github.com/Nerzal/gocloak/v13"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/utils"
)

func NewRunaiAudienceProtocolMapper() gocloak.ProtocolMapperRepresentation {
	return NewAudienceProtocolMapper("Audience Mapper", "runai")
}

// NewRunaiDummyAudienceProtocolMapper remove once researcher service + cluster api know to parse single audience
func NewRunaiDummyAudienceProtocolMapper() gocloak.ProtocolMapperRepresentation {
	return NewAudienceProtocolMapper("Dummy Audience Mapper", "runai-dummy")
}

func NewAudienceProtocolMapper(name, aud string) gocloak.ProtocolMapperRepresentation {
	config := map[string]string{
		"included.custom.audience": aud,
		"id.token.claim":           "true",
		"access.token.claim":       "true",
	}

	return gocloak.ProtocolMapperRepresentation{
		Name:            gocloak.StringP(name),
		Protocol:        gocloak.StringP(utils.AuthProtocolOidc),
		ProtocolMapper:  gocloak.StringP("oidc-audience-mapper"),
		ConsentRequired: gocloak.BoolP(false),
		Config:          &config,
	}
}
