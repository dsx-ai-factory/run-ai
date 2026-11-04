/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/errors"
	gocloakInterface "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/gocloak"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	runaicloak "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/runaicloak"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/utils"
)

const (
	ErrorGetSecret     = "failed getting client secret"
	MoreThan1Clients   = "found more than one client with the same name"
	MngClientId        = "realm-management"
	ExchangePolicyName = "oidc-exchange-policy"

	deleteClientTimeout = 30 * time.Second
)

type KeycloakClientRepository struct {
	auth             auth.Auth
	kcClient         gocloakInterface.GoCloak
	runaicloakClient runaicloak.RunAICloakClient
}

func NewKeycloakClientRepository(client *auth.KeycloakAuthenticator, runaicloakClient *runaicloak.Client) *KeycloakClientRepository {
	return &KeycloakClientRepository{
		auth:             client,
		kcClient:         client.KcClient,
		runaicloakClient: runaicloakClient,
	}
}

func (r *KeycloakClientRepository) CreateClient(ctx context.Context, realm string, client *models.Client) (string, error) {
	token, err := r.auth.GetAdminToken(ctx)
	if err != nil {
		return "", errors.HandleKeyCloakError(err)
	}
	client.SetCreatedAtNow()
	client.SetUpdatedAtNow()
	kcClient := fromClientModel(client)
	id, err := r.kcClient.CreateClient(ctx, *token, realm, *kcClient)
	return id, errors.HandleKeyCloakError(err)
}

func (k *KeycloakClientRepository) GetAccessKeyClientById(ctx context.Context, realm string, createdBy string, id string) (*models.Client, error) {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return nil, errors.HandleKeyCloakError(err)
	}
	if createdBy == "" || strings.ContainsAny(createdBy, " :") {
		return nil, fmt.Errorf("error getting access key by id %s. err: %w", id, errors.ErrInvalidInput)
	}
	accessKeyClients, err := k.getAccessKeysByCreator(ctx, *token, realm, createdBy)
	if err != nil {
		return nil, errors.HandleKeyCloakError(err)
	}
	if len(accessKeyClients) == 0 {
		return nil, fmt.Errorf("%s: %w", "access key is missing", errors.ErrNotFound)
	}
	for _, ak := range accessKeyClients {
		if validateAccessKey(ak, createdBy, id) {
			return toClientModel(ak), nil
		}
	}
	return nil, fmt.Errorf("%s: %w", "access key is missing", errors.ErrNotFound)
}

func validateAccessKey(client *gocloak.Client, createdBy, id string) bool {
	return client.ID != nil && *client.ID == id && client.Attributes != nil && models.IsAccessKeyClientType((*client.Attributes)[utils.ClientTypeAttr]) && (*client.Attributes)[utils.CreatedBy] == createdBy
}

func (r *KeycloakClientRepository) SetEmailScopeOptional(ctx context.Context, realm string, idOfClient string) error {
	return r.setScopeOptional(ctx, realm, idOfClient, utils.EmailScope)
}

func (r *KeycloakClientRepository) SetEmailScopeDefault(ctx context.Context, realm string, idOfClient string) error {
	token, err := r.auth.GetAdminToken(ctx)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	defaultScopes, err := r.kcClient.GetClientsDefaultScopes(ctx, *token, realm, idOfClient)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	for _, s := range defaultScopes {
		if s.Name != nil && *s.Name == utils.EmailScope {
			return nil
		}
	}
	scopeID, err := r.getScopeIdByName(ctx, realm, utils.EmailScope)
	if err != nil {
		return err
	}
	if err := r.kcClient.RemoveOptionalScopeFromClient(ctx, *token, realm, idOfClient, scopeID); err != nil {
		return errors.HandleKeyCloakError(err)
	}
	return errors.HandleKeyCloakError(r.kcClient.AddDefaultScopeToClient(ctx, *token, realm, idOfClient, scopeID))
}

func (r *KeycloakClientRepository) isScopeOptional(ctx context.Context, realm string, idOfClient, scopeName string) (bool, error) {
	token, err := r.auth.GetAdminToken(ctx)
	if err != nil {
		return false, errors.HandleKeyCloakError(err)
	}
	optionalScopes, err := r.kcClient.GetClientsOptionalScopes(ctx, *token, realm, idOfClient)
	if err != nil {
		return false, errors.HandleKeyCloakError(err)
	}
	for _, scope := range optionalScopes {
		if scope.Name != nil && *scope.Name == scopeName {
			return true, nil
		}
	}
	return false, nil
}

func (r *KeycloakClientRepository) getScopeIdByName(ctx context.Context, realm string, scopeName string) (string, error) {
	token, err := r.auth.GetAdminToken(ctx)
	if err != nil {
		return "", errors.HandleKeyCloakError(err)
	}
	clientScopes, err := r.kcClient.GetClientScopes(ctx, *token, realm)
	if err != nil {
		return "", errors.HandleKeyCloakError(err)
	}
	for _, scope := range clientScopes {
		if scope.Name != nil && *scope.Name == scopeName {
			return *scope.ID, nil
		}
	}
	return "", fmt.Errorf("scope not found %s", scopeName)
}

func (r *KeycloakClientRepository) updateScopeToOptional(ctx context.Context, realm string, idOfClient string, scopeID string) error {
	token, err := r.auth.GetAdminToken(ctx)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	err = r.kcClient.RemoveDefaultScopeFromClient(ctx, *token, realm, idOfClient, scopeID)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	err = r.kcClient.AddOptionalScopeToClient(ctx, *token, realm, idOfClient, scopeID)
	return errors.HandleKeyCloakError(err)
}

func (r *KeycloakClientRepository) setScopeOptional(ctx context.Context, realm string, idOfClient, scopeName string) error {
	isOptional, err := r.isScopeOptional(ctx, realm, idOfClient, scopeName)
	if err != nil {
		return err
	}
	if isOptional {
		return nil
	}
	scopeID, err := r.getScopeIdByName(ctx, realm, scopeName)
	if err != nil {
		return err
	}
	return r.updateScopeToOptional(ctx, realm, idOfClient, scopeID)
}

func (r *KeycloakClientRepository) UpdateClient(ctx context.Context, realm string, client *models.Client) error {
	token, err := r.auth.GetAdminToken(ctx)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	client.SetUpdatedAtNow()
	kcClient := fromClientModel(client)
	err = r.kcClient.UpdateClient(ctx, *token, realm, *kcClient)
	return errors.HandleKeyCloakError(err)
}

func (r *KeycloakClientRepository) GetClient(ctx context.Context, realm string, ID string) (*models.Client, error) {
	token, err := r.auth.GetAdminToken(ctx)
	if err != nil {
		return nil, errors.HandleKeyCloakError(err)
	}
	kcClient, err := r.kcClient.GetClient(ctx, *token, realm, ID)
	if err != nil {
		return nil, errors.HandleKeyCloakError(err)
	}
	return toClientModel(kcClient), nil
}

func (r *KeycloakClientRepository) UpdateRedirectUris(ctx context.Context, realm string, ID string, redirectUris []string) error {
	token, err := r.auth.GetAdminToken(ctx)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	kcClient, err := r.kcClient.GetClient(ctx, *token, realm, ID)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	kcClient.RedirectURIs = &redirectUris
	err = r.kcClient.UpdateClient(ctx, *token, realm, *kcClient)
	return errors.HandleKeyCloakError(err)
}

func (r *KeycloakClientRepository) DeleteClient(ctx context.Context, realm string, ID string) error {
	token, err := r.auth.GetAdminToken(ctx)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	err = r.kcClient.DeleteClient(ctx, *token, realm, ID)
	return errors.HandleKeyCloakError(err)
}

// DisableClient disables the client and its direct access grants (password grant) without deleting it.
func (r *KeycloakClientRepository) DisableClient(ctx context.Context, realm string, ID string) error {
	token, err := r.auth.GetAdminToken(ctx)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	kcClient, err := r.kcClient.GetClient(ctx, *token, realm, ID)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	kcClient.Enabled = gocloak.BoolP(false)
	kcClient.DirectAccessGrantsEnabled = gocloak.BoolP(false)
	err = r.kcClient.UpdateClient(ctx, *token, realm, *kcClient)
	return errors.HandleKeyCloakError(err)
}

func (r *KeycloakClientRepository) GetClients(ctx context.Context, realm string, filter *models.ClientFilter) ([]models.Client, error) {
	token, err := r.auth.GetAdminToken(ctx)
	if err != nil {
		return nil, errors.HandleKeyCloakError(err)
	}
	kcClients, err := r.getClientsAllTypeFormats(ctx, *token, realm, filter)
	if err != nil {
		return nil, errors.HandleKeyCloakError(err)
	}
	clients := make([]models.Client, 0)
	for _, kcClient := range kcClients {
		clients = append(clients, *toClientModel(kcClient))
	}
	return clients, nil
}

func (r *KeycloakClientRepository) GetClientSecret(ctx context.Context, realm string, ID string) (string, error) {
	token, err := r.auth.GetAdminToken(ctx)
	if err != nil {
		return "", errors.HandleKeyCloakError(err)
	}
	credentialData, err := r.kcClient.GetClientSecret(ctx, *token, realm, ID)
	if err != nil {
		return "", errors.HandleKeyCloakError(err)
	}
	return *credentialData.Value, nil
}

func (r *KeycloakClientRepository) GetClientByClientId(ctx context.Context, realm string, clientId string) (*models.Client, error) {
	token, err := r.auth.GetAdminToken(ctx)
	if err != nil {
		return nil, errors.HandleKeyCloakError(err)
	}
	kcClients, err := r.kcClient.GetClients(ctx, *token, realm, gocloak.GetClientsParams{ClientID: &clientId})
	if err != nil {
		return nil, errors.HandleKeyCloakError(err)
	}
	if len(kcClients) == 0 {
		return nil, nil
	}
	if len(kcClients) > 1 {
		return nil, fmt.Errorf("%s: %w", MoreThan1Clients, errors.ErrInternal)
	}
	return toClientModel(kcClients[0]), nil
}

func (k *KeycloakClientRepository) CreateExchangePolicy(ctx context.Context, realm string, idpAlias, targetClientID string) error {
	adminToken, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return err
	}

	realmManagement, err := k.GetClientByClientId(ctx, realm, MngClientId)
	if err != nil {
		return err
	}

	err = k.runaicloakClient.CreatePolicyExchangeRequest(ctx, realm, *adminToken, targetClientID, *realmManagement.ID, ExchangePolicyName)
	if err != nil {
		return err
	}

	err = k.setExchangePolicyRequest(ctx, realm, idpAlias, realmManagement.ID)
	if err != nil {
		return err
	}

	return nil
}

func (k *KeycloakClientRepository) setExchangePolicyRequest(ctx context.Context, realm string, idpAlias string, mngIdOfClient *string) error {
	adminToken, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}

	permissionresourceID, exchangeScopePermissionId, err := k.runaicloakClient.GetPermissionsResourceIds(ctx, realm, *adminToken, idpAlias)
	if err != nil {
		return err
	}

	exchangePolicyScopeId, err := k.runaicloakClient.GetPolicyIdForTokenExchange(ctx, realm, *adminToken, *mngIdOfClient)
	if err != nil {
		return err
	}
	exchangePolicyId, err := k.getExchangePolicyId(ctx, realm, *adminToken, *mngIdOfClient)
	if err != nil {
		return err
	}
	exchangePolicyName, err := k.runaicloakClient.GetPolicyNameByScope(ctx, realm, *adminToken, *mngIdOfClient, exchangeScopePermissionId)
	if err != nil {
		return err
	}

	return k.runaicloakClient.SetExchangePolicy(ctx, realm, *adminToken, *mngIdOfClient, permissionresourceID, exchangeScopePermissionId, exchangePolicyScopeId, exchangePolicyId, exchangePolicyName)
}

func (k *KeycloakClientRepository) getExchangePolicyId(ctx context.Context, realm string, adminToken string, mngIdOfClient string) (string, error) {
	params := gocloak.GetPolicyParams{Name: gocloak.StringP(ExchangePolicyName)}
	exchangePolicy, resError := k.kcClient.GetPolicies(ctx, adminToken, realm, mngIdOfClient, params)
	if resError != nil {
		return "", fmt.Errorf("failed to find oidc-exchange-policy policy id. err: %s. %w", resError.Error(), errors.ErrInternal)
	}
	if len(exchangePolicy) != 1 {
		return "", fmt.Errorf("failed to find oidc-exchange-policy policy id %w", errors.ErrInternal)
	}

	return *exchangePolicy[0].ID, nil
}

func (k *KeycloakClientRepository) GetClientsCount(ctx context.Context, realm string, filter *models.ClientFilter) (int, error) {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return 0, errors.HandleKeyCloakError(err)
	}
	kcClients, err := k.getClientsAllTypeFormats(ctx, *token, realm, filter)
	if err != nil {
		return 0, errors.HandleKeyCloakError(err)
	}
	return len(kcClients), nil
}

// getClientsAllTypeFormats runs the query once per stored format of the filtered client type (current and legacy)
// and merges the results. Keycloak attribute search is exact-match and cannot express OR, and the reconcile that
// rewrites legacy values can run more than once, so a lookup by type must see clients in both formats.
func (k *KeycloakClientRepository) getClientsAllTypeFormats(ctx context.Context, token, realm string, filter *models.ClientFilter) ([]*gocloak.Client, error) {
	filters := []*models.ClientFilter{filter}
	if filter != nil && filter.ClientType != nil {
		if legacy, ok := models.LegacyClientType(*filter.ClientType); ok {
			legacyFilter := *filter
			legacyFilter.ClientType = &legacy
			filters = append(filters, &legacyFilter)
		}
	}
	var merged []*gocloak.Client
	seen := map[string]struct{}{}
	for _, f := range filters {
		kcClients, err := k.kcClient.GetClients(ctx, token, realm, *toKcFilters(f))
		if err != nil {
			return nil, err
		}
		for _, c := range kcClients {
			if c.ID != nil {
				if _, dup := seen[*c.ID]; dup {
					continue
				}
				seen[*c.ID] = struct{}{}
			}
			merged = append(merged, c)
		}
	}
	if filter != nil && filter.MaxResults != nil && len(merged) > *filter.MaxResults {
		merged = merged[:*filter.MaxResults]
	}
	return merged, nil
}

// getAccessKeysByCreator returns the access keys created by createdBy, whether stored with the current or the
// legacy client type value.
func (k *KeycloakClientRepository) getAccessKeysByCreator(ctx context.Context, token, realm, createdBy string) ([]*gocloak.Client, error) {
	var merged []*gocloak.Client
	seen := map[string]struct{}{}
	for _, clientType := range []string{models.ClientTypeAccessKey, models.LegacyClientTypeAccessKey} {
		clientsFilter := gocloak.StringP(fmt.Sprintf("%s:%s %s:%s", utils.CreatedBy, createdBy, utils.ClientTypeAttr, clientType))
		kcClients, err := k.kcClient.GetClients(ctx, token, realm, gocloak.GetClientsParams{
			Search:               gocloak.BoolP(true),
			SearchableAttributes: clientsFilter,
		})
		if err != nil {
			return nil, err
		}
		for _, c := range kcClients {
			if c.ID != nil {
				if _, dup := seen[*c.ID]; dup {
					continue
				}
				seen[*c.ID] = struct{}{}
			}
			merged = append(merged, c)
		}
	}
	return merged, nil
}

func (k *KeycloakClientRepository) RegenerateClientSecret(ctx context.Context, realm string, ID string) (string, error) {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return "", errors.HandleKeyCloakError(err)
	}
	credentialData, err := k.kcClient.RegenerateClientSecret(ctx, *token, realm, ID)
	if err != nil {
		return "", errors.HandleKeyCloakError(err)
	}
	if credentialData == nil || credentialData.Value == nil {
		return "", fmt.Errorf("%s: %w", ErrorGetSecret, errors.ErrInternal)
	}
	return *credentialData.Value, nil
}

func (k *KeycloakClientRepository) ToggleClient(ctx context.Context, realm string, ID string, enabled bool) error {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	kcClient, err := k.kcClient.GetClient(ctx, *token, realm, ID)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	kcClient.Enabled = &enabled
	return k.kcClient.UpdateClient(ctx, *token, realm, *kcClient)
}

func (k *KeycloakClientRepository) UpdatePostLogoutRedirectUris(ctx context.Context, realm string, clientId string, customUri string) error {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	kcClients, err := k.kcClient.GetClients(ctx, *token, realm, gocloak.GetClientsParams{
		ClientID: gocloak.StringP(clientId),
	})
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	if len(kcClients) == 0 {
		return fmt.Errorf("client not found: %w", errors.ErrInternal)
	}
	if len(kcClients) > 1 {
		return fmt.Errorf("%s: %w", MoreThan1Clients, errors.ErrInternal)
	}
	kcClient := kcClients[0]
	if kcClient.Attributes == nil {
		kcClient.Attributes = &map[string]string{}
	}
	(*kcClient.Attributes)[utils.LogoutRedirectUris] = utils.BuildLogoutRedirectUrisValue(customUri)
	err = k.kcClient.UpdateClient(ctx, *token, realm, *kcClient)
	return errors.HandleKeyCloakError(err)
}

func (k *KeycloakClientRepository) UpdateClientStatus(ctx context.Context, realm string, ID string, isEnabled bool) error {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	kcClient, err := k.kcClient.GetClient(ctx, *token, realm, ID)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	kcClient.Enabled = &isEnabled
	if kcClient.Attributes == nil {
		kcClient.Attributes = &map[string]string{}
	}
	(*kcClient.Attributes)[utils.UpdateTime] = utils.UnixTimestampToString()
	err = k.kcClient.UpdateClient(ctx, *token, realm, *kcClient)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	return nil
}

func (k *KeycloakClientRepository) DeleteAccessKeysByEmail(ctx context.Context, realm string, email string) error {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	if email == "" {
		return fmt.Errorf("error deleting user: %s access keys. err: %w", email, errors.ErrInvalidInput)
	}
	userAccessKeys, err := k.getAccessKeysByCreator(ctx, *token, realm, email)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	if len(userAccessKeys) > 0 {
		err = k.deleteUserAccessKeys(ctx, realm, token, userAccessKeys)
		if err != nil {
			return fmt.Errorf("failed to delete one or more access keys for user: %s. err: %w. %w", email, err, errors.ErrInternal)
		}
	}
	return nil
}

func (k *KeycloakClientRepository) deleteUserAccessKeys(ctx context.Context, realm string, token *string, userAccessKeys []*gocloak.Client) error {
	errChan := make(chan error, len(userAccessKeys))
	var wg sync.WaitGroup
	for _, ak := range userAccessKeys {
		ak := ak
		wg.Add(1)
		go func() {
			defer wg.Done()
			deleteCtx, cancel := context.WithTimeout(ctx, deleteClientTimeout)
			defer cancel()
			err := k.kcClient.DeleteClient(deleteCtx, *token, realm, *ak.ID)
			if err != nil {
				errChan <- err
			}
		}()
	}
	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}
	return nil
}
