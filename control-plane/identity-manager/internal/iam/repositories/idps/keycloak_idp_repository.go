/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package idps

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"

	"k8s.io/utils/ptr"

	"github.com/Nerzal/gocloak/v13"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/auth"
	iamError "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/errors"
	gocloakInterface "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/gocloak"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/runaicloak"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/config"
)

type KeycloakIdpRepository struct {
	auth             auth.Auth
	kcClient         gocloakInterface.GoCloak
	runaicloakClient runaicloak.RunAICloakClient
}

func NewKeycloakIdpRepository(client *auth.KeycloakAuthenticator, runaicloakClient *runaicloak.Client) *KeycloakIdpRepository {
	return &KeycloakIdpRepository{
		auth:             client,
		kcClient:         client.KcClient,
		runaicloakClient: runaicloakClient,
	}
}

func (k *KeycloakIdpRepository) GetIdp(ctx context.Context, realm string, idpAlias string) (*models.Idp, error) {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return nil, iamError.HandleKeyCloakError(err)
	}
	kcIdp, err := k.kcClient.GetIdentityProvider(ctx, *token, realm, idpAlias)
	if err != nil {
		return nil, iamError.HandleKeyCloakError(err)
	}
	return k.toIdpModel(kcIdp), nil
}

func (k *KeycloakIdpRepository) CreateIdp(ctx context.Context, realm string, idp models.Idp) (*string, error) {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return nil, iamError.HandleKeyCloakError(err)
	}
	kcIdp := k.fromIdpModel(&idp, nil)
	_, err = k.kcClient.CreateIdentityProvider(ctx, *token, realm, *kcIdp)
	if err != nil {
		return nil, iamError.HandleKeyCloakError(err)
	}
	if idp.SupportTokenExchange != nil && *idp.SupportTokenExchange {
		err = k.runaicloakClient.ToggleManagementPermissionsForIdp(ctx, realm, *token, *kcIdp.Alias, true)
		if err != nil {
			return nil, iamError.HandleKeyCloakError(err)
		}
	}
	return idp.Alias, nil
}

func (k *KeycloakIdpRepository) UpdateIdp(ctx context.Context, realm string, idp models.Idp) error {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return iamError.HandleKeyCloakError(err)
	}
	kcIdp, err := k.kcClient.GetIdentityProvider(ctx, *token, realm, *idp.Alias)
	if err != nil {
		return fmt.Errorf("failed to get idp %s in realm %s: %w", *idp.Alias, realm, iamError.HandleKeyCloakError(err))
	}
	kcIdp = k.fromIdpModel(&idp, kcIdp)
	err = k.kcClient.UpdateIdentityProvider(ctx, *token, realm, *idp.Alias, *kcIdp)
	if err != nil {
		return iamError.HandleKeyCloakError(err)
	}
	return nil
}

func (k *KeycloakIdpRepository) DeleteIdp(ctx context.Context, realm string, idpAlias string) error {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return iamError.HandleKeyCloakError(err)
	}

	if idpAlias != "saml" && idpAlias != "openshift-v4" && !strings.HasPrefix(idpAlias, "aliastest") {
		_ = k.runaicloakClient.ToggleManagementPermissionsForIdp(ctx, realm, *token, idpAlias, false)
	}

	err = k.kcClient.DeleteIdentityProvider(ctx, *token, realm, idpAlias)
	if err != nil {
		return iamError.HandleKeyCloakError(err)
	}
	return nil
}

func (k *KeycloakIdpRepository) GetIdps(ctx context.Context, realm string) ([]*models.Idp, error) {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return nil, iamError.HandleKeyCloakError(err)
	}
	kcIdps, err := k.kcClient.GetIdentityProviders(ctx, *token, realm)
	if err != nil {
		return nil, iamError.HandleKeyCloakError(err)
	}
	var result []*models.Idp
	for _, idp := range kcIdps {
		result = append(result, k.toIdpModel(idp))
	}
	return result, nil
}

func (k *KeycloakIdpRepository) GetIdpAliasForExchangeToken(ctx context.Context, realm string) (*string, error) {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return nil, iamError.HandleKeyCloakError(err)
	}
	kcIdps, err := k.kcClient.GetIdentityProviders(ctx, *token, realm)
	if err != nil {
		return nil, iamError.HandleKeyCloakError(err)
	}
	for _, idp := range kcIdps {
		if idp.ProviderID != nil && *idp.ProviderID == "oidc" && k.runaicloakClient.DoesIdpHaveExchangePermissions(ctx, realm, *token, *idp.Alias) {
			return idp.Alias, nil
		}
	}
	return nil, fmt.Errorf("failed to find idp with token-exchange permissions %w", iamError.ErrNotFound)
}

func (k *KeycloakIdpRepository) IsIdpConfigured(ctx context.Context, realm string) (bool, error) {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return false, iamError.HandleKeyCloakError(err)
	}
	kcIdps, err := k.kcClient.GetIdentityProviders(ctx, *token, realm)
	if err != nil {
		return false, fmt.Errorf("failed to get idps. err: %s", err.Error())
	}
	return len(kcIdps) > 0, nil
}

func (k *KeycloakIdpRepository) GetIdpMappers(ctx context.Context, realm, alias string) ([]*models.IdpMapper, error) {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return nil, iamError.HandleKeyCloakError(err)
	}
	kcMappers, err := k.kcClient.GetIdentityProviderMappers(ctx, *token, realm, alias)
	if err != nil {
		return nil, iamError.HandleKeyCloakError(err)
	}
	var result []*models.IdpMapper
	for _, mapper := range kcMappers {
		result = append(result, k.toIdpMapperModel(mapper))
	}
	return result, nil
}

func (k *KeycloakIdpRepository) UpdateIdpMappers(ctx context.Context, realm, alias string, mappers []*models.IdpMapper) error {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return iamError.HandleKeyCloakError(err)
	}
	var mapperErr error
	for _, mapper := range mappers {
		kcMapper := k.fromIdpMapperModel(mapper)
		if mapper.ID != nil {
			err = k.kcClient.UpdateIdentityProviderMapper(ctx, *token, realm, alias, *kcMapper)
		} else {
			_, err = k.kcClient.CreateIdentityProviderMapper(ctx, *token, realm, alias, *kcMapper)
		}
		mapperErr = errors.Join(mapperErr, err)
	}
	if mapperErr != nil {
		return fmt.Errorf("failed to update idp mappers. err: %s", mapperErr.Error())
	}
	return mapperErr
}

func (k *KeycloakIdpRepository) GetIdpConfig(ctx context.Context, realm string, url string, providerId string) (map[string]string, error) {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return nil, iamError.HandleKeyCloakError(err)
	}

	config, err := k.kcClient.ImportIdentityProviderConfig(ctx, *token, realm, url, providerId)
	if err != nil {
		logging.Errorf("failed to get identityProvider from url. err: %+v", err)
	}
	return config, iamError.HandleKeyCloakError(err)
}

func (k *KeycloakIdpRepository) GetIdpConfigFromFile(ctx context.Context, realm string, providerId string, fileName string, fileBody io.Reader) (map[string]string, error) {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return nil, iamError.HandleKeyCloakError(err)
	}

	config, err := k.kcClient.ImportIdentityProviderConfigFromFile(ctx, *token, realm, providerId, fileName, fileBody)
	if err != nil {
		logging.Errorf("failed to get identityProvider from url. err: %+v", err)
	}
	return config, iamError.HandleKeyCloakError(err)
}

func (k *KeycloakIdpRepository) SetBrowserRedirector(ctx context.Context, realm string, alias string) error {
	flowName := "browser"
	providerId := "identity-provider-redirector"

	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return iamError.HandleKeyCloakError(err)
	}
	flows, err := k.kcClient.GetAuthenticationExecutions(ctx, *token, realm, flowName)
	if err != nil {
		return fmt.Errorf("setBrowserRedirector: failed to get login flow. err: %s", err.Error())
	}

	for _, flow := range flows {
		if *flow.ProviderID == providerId {
			err := k.updateIDPExecutionConfig(ctx, realm, *flow.ID, alias)
			if err != nil {
				return fmt.Errorf("setBrowserRedirector: failed to update idp execution config. err: %s", err.Error())
			}
			return nil
		}
	}
	return nil
}

func (k *KeycloakIdpRepository) SetMandatotyClaimMapper(ctx context.Context, realm, alias string, mapper *models.IdpMapper) error {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return iamError.HandleKeyCloakError(err)
	}
	mandatoryClaimMapper, err := k.getIdpMapperByName(ctx, *token, realm, alias, *mapper.Name)
	if err != nil {
		return iamError.HandleKeyCloakError(err)
	}
	kcMapper := k.fromIdpMapperModel(mapper)
	if mandatoryClaimMapper == nil {
		_, err = k.kcClient.CreateIdentityProviderMapper(ctx, *token, realm, alias, *kcMapper)
	} else {
		kcMapper.ID = mandatoryClaimMapper.ID
		err = k.kcClient.UpdateIdentityProviderMapper(ctx, *token, realm, alias, *kcMapper)
	}
	if err != nil {
		return iamError.HandleKeyCloakError(err)
	}
	return nil
}

func (k *KeycloakIdpRepository) DeleteMandatoryClaimMapper(ctx context.Context, realm, alias, name string) error {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return iamError.HandleKeyCloakError(err)
	}
	mandatoryClaimMapper, err := k.getIdpMapperByName(ctx, *token, realm, alias, name)
	if err != nil {
		return err
	}
	if mandatoryClaimMapper != nil {
		return k.kcClient.DeleteIdentityProviderMapper(ctx, *token, realm, alias, *mandatoryClaimMapper.ID)
	}
	return nil
}

func (k *KeycloakIdpRepository) getIdpMapperByName(ctx context.Context, token, realm, alias, name string) (*gocloak.IdentityProviderMapper, error) {
	mappers, err := k.kcClient.GetIdentityProviderMappers(ctx, token, realm, alias)
	if err != nil {
		return nil, iamError.HandleKeyCloakError(err)
	}
	for _, m := range mappers {
		if *m.Name == name {
			return m, nil
		}
	}
	return nil, nil
}

func (k *KeycloakIdpRepository) updateIDPExecutionConfig(ctx context.Context, realm string, executionId string, alias string) error {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return iamError.HandleKeyCloakError(err)
	}
	return k.runaicloakClient.UpdateIDPExecutionConfig(ctx, realm, *token, executionId, alias)
}

func (k *KeycloakIdpRepository) toIdpMapperModel(kcIdpMapper *gocloak.IdentityProviderMapper) *models.IdpMapper {
	return &models.IdpMapper{
		ID:                     kcIdpMapper.ID,
		Name:                   kcIdpMapper.Name,
		IdentityProviderMapper: kcIdpMapper.IdentityProviderMapper,
		IdentityProviderAlias:  kcIdpMapper.IdentityProviderAlias,
		Config:                 kcIdpMapper.Config,
	}
}

func (k *KeycloakIdpRepository) fromIdpMapperModel(idpMapper *models.IdpMapper) *gocloak.IdentityProviderMapper {
	return &gocloak.IdentityProviderMapper{
		ID:                     idpMapper.ID,
		Name:                   idpMapper.Name,
		IdentityProviderMapper: idpMapper.IdentityProviderMapper,
		IdentityProviderAlias:  idpMapper.IdentityProviderAlias,
		Config:                 idpMapper.Config,
	}
}

func (k *KeycloakIdpRepository) toIdpModel(kcIdp *gocloak.IdentityProviderRepresentation) *models.Idp {
	return &models.Idp{
		Alias:       kcIdp.Alias,
		ProviderID:  kcIdp.ProviderID,
		DisplayName: kcIdp.DisplayName,
		Enabled:     kcIdp.Enabled,
		Config:      kcIdp.Config,
	}
}

func (k *KeycloakIdpRepository) fromIdpModel(idp *models.Idp, kcIdp *gocloak.IdentityProviderRepresentation) *gocloak.IdentityProviderRepresentation {
	if kcIdp == nil {
		kcIdp = &gocloak.IdentityProviderRepresentation{}
	}
	kcIdp.Alias = idp.Alias
	kcIdp.ProviderID = idp.ProviderID
	kcIdp.DisplayName = idp.DisplayName
	kcIdp.Enabled = idp.Enabled
	kcIdp.Config = idp.Config
	kcIdp.AddReadTokenRoleOnCreate = idp.AddReadTokenRoleOnCreate
	kcIdp.FirstBrokerLoginFlowAlias = idp.FirstBrokerLoginFlowAlias
	kcIdp.InternalID = idp.InternalID
	kcIdp.LinkOnly = idp.LinkOnly
	kcIdp.PostBrokerLoginFlowAlias = idp.PostBrokerLoginFlowAlias
	kcIdp.StoreToken = idp.SupportTokenExchange
	kcIdp.TrustEmail = ptr.To(true)
	return kcIdp
}

func (k *KeycloakIdpRepository) GetSPSamlMetadataUrl(ctx context.Context, realm string, idpAlias string) string {
	tenantRealm := *runai_config.Get[*config.TenantRealm]()
	return fmt.Sprintf("%s/auth/realms/%s/broker/%s/endpoint/descriptor", tenantRealm.AppHost, realm, idpAlias)
}
