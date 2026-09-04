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

func NewSamlIdpMapperMapper() UserAttributeMapper {
	return SamlIdpMapperMapper{}
}

type SamlIdpMapperMapper struct{}

func (s SamlIdpMapperMapper) GenerateUserAttributeMapper(alias string, mapper models.IdpAttributeMapper, value *string) *iamModels.IdpMapper {
	if value == nil || *value == "" {
		value = &mapper.Value
	}
	config := samlUserAttributeConfig(mapper.Attribute, *value)

	return &iamModels.IdpMapper{
		Name:                   &mapper.Name,
		IdentityProviderAlias:  &alias,
		IdentityProviderMapper: ptr.To(samlMapper),
		Config:                 config,
	}
}
