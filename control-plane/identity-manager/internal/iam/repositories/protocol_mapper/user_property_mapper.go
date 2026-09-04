/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package protocol_mapper

import (
	"strconv"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/utils"
	"github.com/Nerzal/gocloak/v13"
)

type UserPropertyMapperConfig struct {
	MapperName        string
	ClaimName         string
	ClaimType         ClaimType
	UserAttribute     string
	ShowInIdToken     bool
	ShowInUserInfo    bool
	ShowInAccessToken bool
}

func NewSubjectUidProtocolMapper() gocloak.ProtocolMapperRepresentation {
	return newUserPropertyMapper(UserPropertyMapperConfig{
		MapperName:        "SubjectUuid Mapper",
		ClaimName:         "subject_uuid",
		ClaimType:         StringClaim,
		UserAttribute:     "id",
		ShowInUserInfo:    true,
		ShowInIdToken:     true,
		ShowInAccessToken: true,
	})
}

func NewSubToEmailProtocolMapper() gocloak.ProtocolMapperRepresentation {
	return newUserPropertyMapper(UserPropertyMapperConfig{
		MapperName:        "SubToEmail Mapper",
		ClaimName:         "sub",
		ClaimType:         StringClaim,
		UserAttribute:     "email",
		ShowInUserInfo:    true,
		ShowInIdToken:     true,
		ShowInAccessToken: true,
	})
}

func newUserPropertyMapper(mapperConfig UserPropertyMapperConfig) gocloak.ProtocolMapperRepresentation {
	config := map[string]string{
		"claim.name":           mapperConfig.ClaimName,
		"user.attribute":       mapperConfig.UserAttribute,
		"jsonType.label":       mapperConfig.ClaimType,
		"id.token.claim":       strconv.FormatBool(mapperConfig.ShowInIdToken),
		"userinfo.token.claim": strconv.FormatBool(mapperConfig.ShowInUserInfo),
		"access.token.claim":   strconv.FormatBool(mapperConfig.ShowInAccessToken),
	}

	return gocloak.ProtocolMapperRepresentation{
		Name:            gocloak.StringP(mapperConfig.MapperName),
		Protocol:        gocloak.StringP(utils.AuthProtocolOidc),
		ProtocolMapper:  gocloak.StringP("oidc-usermodel-property-mapper"),
		ConsentRequired: gocloak.BoolP(false),
		Config:          &config,
	}
}
