/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package protocol_mapper

import (
	"github.com/Nerzal/gocloak/v13"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/utils"
)

func NewRunaiGroupsCountProtocolMapper(idTokenGroupsClaim bool) gocloak.ProtocolMapperRepresentation {
	config := buildBaseTokenClaimsConfig("groups_count", idTokenGroupsClaim)
	config["list.user.attribute"] = utils.Groups

	return gocloak.ProtocolMapperRepresentation{
		Name:            gocloak.StringP("Runai Groups Count Mapper"),
		Protocol:        gocloak.StringP(utils.AuthProtocolOidc),
		ProtocolMapper:  gocloak.StringP("runai-counter-mapper"),
		ConsentRequired: gocloak.BoolP(false),
		Config:          &config,
	}
}
