/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mappers

import (
	iamModels "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/models"
	"k8s.io/utils/ptr"
)

func NewOcpIdpMapperMapper() UserAttributeMapper {
	return OcpIdpMapperMapper{}
}

type OcpIdpMapperMapper struct{}

func (o OcpIdpMapperMapper) GenerateUserAttributeMapper(alias string, mapper models.IdpAttributeMapper, value *string) *iamModels.IdpMapper {
	if value == nil || *value == "" {
		value = &mapper.Value
	}
	config := ocpUserAttributeConfig(mapper.Attribute, *value)

	return &iamModels.IdpMapper{
		Name:                   &mapper.Name,
		IdentityProviderAlias:  &alias,
		IdentityProviderMapper: ptr.To(ocpMapper),
		Config:                 config,
	}
}
