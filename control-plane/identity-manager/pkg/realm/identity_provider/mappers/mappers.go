/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mappers

import (
	iamModels "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/models"
	"k8s.io/utils/ptr"
)

const (
	hardcodedMapper string = "hardcoded-attribute-idp-mapper"
	oidcMapper      string = "runai-oidc-user-attribute-idp-mapper"
	samlMapper      string = "saml-user-attribute-idp-mapper"
	ocpMapper       string = "openshift-v4-user-attribute-mapper"
)

type UserAttributeMapper interface {
	GenerateUserAttributeMapper(alias string, attributeMapper models.IdpAttributeMapper, value *string) *iamModels.IdpMapper
}

func GenerateHardCodedMapper(alias string, mapper models.IdpAttributeMapper) *iamModels.IdpMapper {
	config := map[string]string{
		"syncMode":        "FORCE",
		"jsonField":       mapper.Attribute,
		"attribute":       mapper.Attribute,
		"attribute.value": mapper.Value,
	}
	return &iamModels.IdpMapper{
		Name:                   &mapper.Name,
		IdentityProviderAlias:  &alias,
		IdentityProviderMapper: ptr.To(hardcodedMapper),
		Config:                 &config,
	}
}

func GenerateUserAttributeMappers(alias string, attributeMappers []models.IdpAttributeMapper, values map[string]string, mapper UserAttributeMapper) []*iamModels.IdpMapper {
	var mappers []*iamModels.IdpMapper
	for _, idpMapper := range attributeMappers {
		var attributeVal *string
		if val, ok := values[idpMapper.Attribute]; ok {
			attributeVal = &val
		}
		mappers = append(mappers, mapper.GenerateUserAttributeMapper(alias, idpMapper, attributeVal))
	}
	return mappers
}
