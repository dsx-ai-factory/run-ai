/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package identity_provider

import (
	"context"
	"fmt"
	"net/url"

	runai_config "github.com/run-ai/runai/runai-common-packages/config"

	iamModels "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/config"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/models"
	"k8s.io/utils/ptr"
)

type IdentityProviderRepresentation struct {
	ProviderID                models.ProviderID
	FirstBrokerLoginFlowAlias string
	AddReadTokenRoleOnCreate  bool
	hardCodedMappers          models.IdpAttributeMapper
	attributeMappers          []models.IdpAttributeMapper
}

func basicIdpRep(alias string, base *IdentityProviderRepresentation, config *map[string]string) iamModels.Idp {
	identityProvider := iamModels.Idp{}

	identityProvider.Alias = &alias
	identityProvider.DisplayName = &alias
	identityProvider.ProviderID = ptr.To(string(base.ProviderID))
	identityProvider.SupportTokenExchange = ptr.To(base.ProviderID == models.OIDC)
	identityProvider.AddReadTokenRoleOnCreate = &base.AddReadTokenRoleOnCreate
	identityProvider.FirstBrokerLoginFlowAlias = &base.FirstBrokerLoginFlowAlias
	identityProvider.Config = config

	return identityProvider
}

type IdpRepresentationFactory struct {
	idpRepresentation map[models.ProviderID]IdpRepresentationer
}

func (t IdpRepresentationFactory) GetIdpRepresentationer(ctx context.Context, provider models.ProviderID) (IdpRepresentationer, error) {
	rep, ok := t.idpRepresentation[provider]
	if ok {
		return rep, nil
	}
	return nil, fmt.Errorf("unsupported provider id: %s", string(provider))
}

func NewIdpRepresentationFactory() IdpRepresentationFactory {
	tenantRealm := *runai_config.Get[*config.TenantRealm]()
	parsedURL, err := url.Parse(tenantRealm.AppHost)
	if err != nil {
		return IdpRepresentationFactory{}
	}
	idpRep := map[models.ProviderID]IdpRepresentationer{}
	idpRep[models.SAML] = NewSamlIdentityProvider(fmt.Sprintf("runai-%s/runai", parsedURL.Hostname()))
	idpRep[models.OCP] = NewOcpIdentityProvider()
	idpRep[models.OIDC] = NewOidcIdentityProvider()

	return IdpRepresentationFactory{
		idpRepresentation: idpRep,
	}
}
