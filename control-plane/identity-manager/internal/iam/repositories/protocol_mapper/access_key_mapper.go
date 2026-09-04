/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package protocol_mapper

import (
	"github.com/Nerzal/gocloak/v13"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/utils"
)

const accessKeyAttributeProviderID = "runai-access-key-attribute-mapper"

func newAccessKeyAttributeProtocolMapper(name, claimName, userAttribute, tenantId string, multivalued, intersectionEnabled bool) gocloak.ProtocolMapperRepresentation {
	config := buildAccessKeyAttributeMapperConfig(claimName, userAttribute, tenantId, multivalued, intersectionEnabled)

	return gocloak.ProtocolMapperRepresentation{
		Name:            gocloak.StringP(name),
		Protocol:        gocloak.StringP(utils.AuthProtocolOidc),
		ProtocolMapper:  gocloak.StringP(accessKeyAttributeProviderID),
		ConsentRequired: gocloak.BoolP(false),
		Config:          &config,
	}
}

func NewAccessKeyGroupsProtocolMapper(tenantId string) gocloak.ProtocolMapperRepresentation {
	return newAccessKeyAttributeProtocolMapper("Access Key Groups Mapper", "groups", utils.Groups, tenantId, true, true)
}

func NewAccessKeyUIDProtocolMapper(tenantId string) gocloak.ProtocolMapperRepresentation {
	return newAccessKeyAttributeProtocolMapper("Access Key UID Mapper", "uid", utils.UID, tenantId, false, false)
}

func NewAccessKeyGIDProtocolMapper(tenantId string) gocloak.ProtocolMapperRepresentation {
	return newAccessKeyAttributeProtocolMapper("Access Key GID Mapper", "gid", utils.GID, tenantId, false, false)
}

func NewAccessKeySupplementaryGroupsProtocolMapper(tenantId string) gocloak.ProtocolMapperRepresentation {
	return newAccessKeyAttributeProtocolMapper("Access Key SupplementaryGroups Mapper", "supplementarygroups", utils.SupplementaryGroups, tenantId, true, false)
}
