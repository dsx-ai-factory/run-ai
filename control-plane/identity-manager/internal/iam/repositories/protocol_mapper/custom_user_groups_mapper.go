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

func NewRunaiGroupsProtocolMapper(idTokenGroupsClaim bool, tenantId string) gocloak.ProtocolMapperRepresentation {
	config := buildGroupsMapperConfig("groups", tenantId)
	config["id.token.claim"] = strconv.FormatBool(idTokenGroupsClaim)

	return gocloak.ProtocolMapperRepresentation{
		Name:            gocloak.StringP("Runai Groups Mapper"),
		Protocol:        gocloak.StringP(utils.AuthProtocolOidc),
		ProtocolMapper:  gocloak.StringP("runai-oidc-custom-groups-mapper"),
		ConsentRequired: gocloak.BoolP(false),
		Config:          &config,
	}
}
