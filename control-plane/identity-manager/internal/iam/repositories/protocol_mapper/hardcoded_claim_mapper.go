/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package protocol_mapper

import (
	"strconv"

	"github.com/Nerzal/gocloak/v13"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/utils"
)

type HardcodedClaimMapperConfig struct {
	MapperName        string
	ClaimName         string
	ClaimType         ClaimType
	ClaimValue        string
	ShowInUserInfo    bool
	ShowInIdToken     bool
	ShowInAccessToken bool
}

func NewTenantIdProtocolMapper(tenantId string) gocloak.ProtocolMapperRepresentation {
	return newHardcodedClaimProtocolMapper(HardcodedClaimMapperConfig{
		MapperName:        "TenantId Mapper",
		ClaimName:         "tenant_id",
		ClaimType:         IntClaim,
		ClaimValue:        tenantId,
		ShowInUserInfo:    true,
		ShowInIdToken:     true,
		ShowInAccessToken: true,
	})
}

func NewClientTypeProtocolMapper(clientType string) gocloak.ProtocolMapperRepresentation {
	return newHardcodedClaimProtocolMapper(HardcodedClaimMapperConfig{
		MapperName:        "ClientType Mapper",
		ClaimName:         "client_type",
		ClaimType:         StringClaim,
		ClaimValue:        clientType,
		ShowInUserInfo:    true,
		ShowInIdToken:     true,
		ShowInAccessToken: true,
	})
}

func NewCreatedByToSubProtocolMapper(createdBy string) gocloak.ProtocolMapperRepresentation {
	return newHardcodedClaimProtocolMapper(HardcodedClaimMapperConfig{
		MapperName:        "CreatedByToSub Mapper",
		ClaimName:         "sub",
		ClaimType:         StringClaim,
		ClaimValue:        createdBy,
		ShowInUserInfo:    true,
		ShowInIdToken:     true,
		ShowInAccessToken: true,
	})
}

func NewCreatedByToEmailProtocolMapper(createdBy string) gocloak.ProtocolMapperRepresentation {
	return newHardcodedClaimProtocolMapper(HardcodedClaimMapperConfig{
		MapperName:        "CreatedByToEmail Mapper",
		ClaimName:         "email",
		ClaimType:         StringClaim,
		ClaimValue:        createdBy,
		ShowInUserInfo:    true,
		ShowInIdToken:     true,
		ShowInAccessToken: true,
	})
}

func newHardcodedClaimProtocolMapper(mapperConfig HardcodedClaimMapperConfig) gocloak.ProtocolMapperRepresentation {
	config := map[string]string{
		"claim.name":           mapperConfig.ClaimName,
		"claim.value":          mapperConfig.ClaimValue,
		"jsonType.label":       mapperConfig.ClaimType,
		"id.token.claim":       strconv.FormatBool(mapperConfig.ShowInIdToken),
		"userinfo.token.claim": strconv.FormatBool(mapperConfig.ShowInUserInfo),
		"access.token.claim":   strconv.FormatBool(mapperConfig.ShowInAccessToken),
	}

	return gocloak.ProtocolMapperRepresentation{
		Name:            gocloak.StringP(mapperConfig.MapperName),
		Protocol:        gocloak.StringP(utils.AuthProtocolOidc),
		ProtocolMapper:  gocloak.StringP("oidc-hardcoded-claim-mapper"),
		ConsentRequired: gocloak.BoolP(false),
		Config:          &config,
	}
}
