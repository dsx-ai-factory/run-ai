/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package identity_provider

import (
	"context"

	iamModels "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/models"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/realm/identity_provider/mappers"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

type ocpIdentityProviderRepresentation struct {
	base         IdentityProviderRepresentation
	mapperMapper mappers.UserAttributeMapper
}

func NewOcpIdentityProvider() IdpRepresentationer {
	base := IdentityProviderRepresentation{
		ProviderID:                models.OCP,
		FirstBrokerLoginFlowAlias: utils.LOGIN_FLOW_NAME,
		AddReadTokenRoleOnCreate:  true,
		hardCodedMappers:          models.ExternalUser,
		attributeMappers:          []models.IdpAttributeMapper{models.OCPEmail, models.OCPGroups, models.OCPUid, models.OCPGid, models.OCPSupGroup},
	}
	return &ocpIdentityProviderRepresentation{
		base:         base,
		mapperMapper: mappers.NewOcpIdpMapperMapper(),
	}
}

func (oip *ocpIdentityProviderRepresentation) GetRepresentation(ctx context.Context, data *models.IdpData, config *map[string]string) *iamModels.Idp {
	if data.Idp != oip.base.ProviderID {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("cannot update Idp type from %s to %s", string(data.Idp), oip.base.ProviderID)
		return nil
	}

	if config != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("shouldn't happen in ocp")
	}
	if data.Config != nil {
		if (*data.Config)["clientId"] != "" {
			(*config)["clientId"] = (*data.Config)["clientId"]
		}

		if (*data.Config)["clientSecret"] != "" {
			(*config)["clientSecret"] = (*data.Config)["clientSecret"]
		}

		if (*data.Config)["baseUrl"] != "" {
			(*config)["baseUrl"] = (*data.Config)["baseUrl"]
		}
	}

	identityProvider := basicIdpRep(data.Alias, &oip.base, config)

	return &identityProvider
}

func (oip *ocpIdentityProviderRepresentation) GetMappersRepresentation(ctx context.Context, alias string, values map[string]string) []*iamModels.IdpMapper {
	idpMappers := []*iamModels.IdpMapper{
		mappers.GenerateHardCodedMapper(alias, oip.base.hardCodedMappers),
	}
	return append(idpMappers, mappers.GenerateUserAttributeMappers(alias, oip.base.attributeMappers, values, oip.mapperMapper)...)
}
