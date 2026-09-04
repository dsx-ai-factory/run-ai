/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package realms

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/errors"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/runaicloak"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/utils"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/auth"
	gocloakInterface "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/gocloak"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	helpers "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/repositories/realms/helpers"
	"github.com/Nerzal/gocloak/v13"
)

type KeycloakRealmRepository struct {
	auth             auth.Auth
	kcClient         gocloakInterface.GoCloak
	runaicloakClient runaicloak.RunAICloakClient
}

func NewKeycloakRealmRepository(client *auth.KeycloakAuthenticator, runaicloakClient *runaicloak.Client) *KeycloakRealmRepository {
	return &KeycloakRealmRepository{
		auth:             client,
		kcClient:         client.KcClient,
		runaicloakClient: runaicloakClient,
	}
}

func (k *KeycloakRealmRepository) CreateRealm(ctx context.Context, realm *models.Realm) (string, error) {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return "", errors.HandleKeyCloakError(err)
	}
	if realm.Name == nil {
		return "", fmt.Errorf("failed to create realm. realm name is missing err: %s", errors.ErrInvalidInput)
	}
	kcRealm := helpers.FromRealmModel(realm, &gocloak.RealmRepresentation{})
	_, err = k.kcClient.CreateRealm(ctx, *token, *kcRealm)
	if err != nil {
		return "", errors.HandleKeyCloakError(err)
	}
	err = k.updateRealmUserProfileMetadata(ctx, *realm.Name, *token, k.auth.GetHost())
	if err != nil {
		return "", fmt.Errorf("failed to create realm. failed creating user profile attributes. err: %s", err.Error())
	}
	return *realm.Name, nil
}

func (k *KeycloakRealmRepository) DisableRealm(ctx context.Context, realmName string) error {
	realm, err := k.GetRealm(ctx, realmName)
	if err != nil {
		return fmt.Errorf("failed to get realm. err: %s", err.Error())
	}
	realm.Enabled = gocloak.BoolP(false)
	err = k.UpdateRealm(ctx, *realm)
	if err != nil {
		return fmt.Errorf("failed to disable realm. err: %s", err.Error())
	}
	return nil
}

func (k *KeycloakRealmRepository) GetRealm(ctx context.Context, realmName string) (*models.Realm, error) {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return nil, errors.HandleKeyCloakError(err)
	}
	realm, err := k.kcClient.GetRealm(ctx, *token, realmName)
	if err != nil {
		if keycloakError, ok := err.(*gocloak.APIError); ok && keycloakError.Code == http.StatusNotFound {
			return nil, fmt.Errorf("realm with name %q: %w", realmName, errors.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to get realm. err: %s", err.Error())
	}
	return helpers.ToRealmModel(realm), nil
}

func (k *KeycloakRealmRepository) GetRealms(ctx context.Context) ([]*models.Realm, error) {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return nil, errors.HandleKeyCloakError(err)
	}
	realms, err := k.kcClient.GetRealms(ctx, *token)
	if err != nil {
		return nil, fmt.Errorf("failed to get realms. err: %s", err.Error())
	}
	var result []*models.Realm
	for _, realm := range realms {
		result = append(result, helpers.ToRealmModel(realm))
	}
	return result, nil
}

func (k *KeycloakRealmRepository) UpdateRealm(ctx context.Context, realm models.Realm) error {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	currentRealm, err := k.kcClient.GetRealm(ctx, *token, *realm.Name)
	if err != nil {
		return fmt.Errorf("failed to get realm. err: %s", err.Error())
	}

	updatedRealm := helpers.FromRealmModel(&realm, currentRealm)
	err = k.kcClient.UpdateRealm(ctx, *token, *updatedRealm)
	if err != nil {
		return fmt.Errorf("failed to update realm. err: %s", err.Error())
	}
	err = k.updateRealmUserProfileMetadata(ctx, *realm.Name, *token, k.auth.GetHost())
	if err != nil {
		return fmt.Errorf("failed to updating realm. failed updating user profile attributes. err: %s", err.Error())
	}
	return nil
}

func (k *KeycloakRealmRepository) DeleteRealm(ctx context.Context, realmName string) error {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	err = k.kcClient.DeleteRealm(ctx, *token, realmName)
	if err != nil {
		return fmt.Errorf("failed to delete realm. err: %s", err.Error())
	}
	return nil
}

func (k *KeycloakRealmRepository) UpdateAutoRedirectSSO(ctx context.Context, realm string, idpAlias *string, isEnabled bool) error {
	authFlow, err := k.getAuthFlowByProviderID(ctx, realm, utils.IdpProviderID)
	if err != nil {
		return err
	}
	if isEnabled && idpAlias != nil {
		return k.enableAutoRedirectSSO(ctx, realm, *authFlow.ID, *idpAlias)
	}
	if !isEnabled {
		return k.disableAutoRedirectSSO(ctx, realm, authFlow.AuthenticationConfig)
	}
	return fmt.Errorf("invalid settings fields")
}

func (k *KeycloakRealmRepository) enableAutoRedirectSSO(ctx context.Context, realm, authFlowID, idpAlias string) error {
	return k.updateAuthFlowExecutionConfig(ctx, realm, authFlowID, idpAlias)
}

func (k *KeycloakRealmRepository) disableAutoRedirectSSO(ctx context.Context, realm string, authFlowConfig *string) error {
	if authFlowConfig == nil {
		return nil
	}
	return k.deleteAuthFlowExecutionConfig(ctx, realm, *authFlowConfig)
}

func (k *KeycloakRealmRepository) updateAuthFlowExecutionConfig(ctx context.Context, realm string, executionID string, alias string) error {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}

	err = k.runaicloakClient.UpdateAuthFlowExecutionConfig(ctx, realm, *token, executionID, alias)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	return nil
}

func (k *KeycloakRealmRepository) deleteAuthFlowExecutionConfig(ctx context.Context, realm string, configID string) error {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	err = k.runaicloakClient.DeleteAuthFlowExecutionConfig(ctx, realm, *token, configID)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	return nil
}

func (k *KeycloakRealmRepository) getAuthFlowByProviderID(ctx context.Context, realm, authFlowProviderID string) (*models.ModifyAuthenticationExecutionRepresentation, error) {
	authFlows, err := k.getAuthExecutions(ctx, realm, utils.BrowserFlowAlias)
	if err != nil {
		return nil, err
	}
	for _, flow := range authFlows {
		if flow.ProviderID != nil && *flow.ProviderID == authFlowProviderID {
			return flow, nil
		}
	}
	return nil, fmt.Errorf("couldn't find auth flow by provider id: %s", authFlowProviderID)
}

func (k *KeycloakRealmRepository) getAuthExecutions(ctx context.Context, realm string, flowAlias string) ([]*models.ModifyAuthenticationExecutionRepresentation, error) {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return nil, err
	}
	executions, err := k.kcClient.GetAuthenticationExecutions(ctx, *token, realm, flowAlias)
	if err != nil {
		return nil, errors.HandleKeyCloakError(err)
	}
	var modifyAuthRep []*models.ModifyAuthenticationExecutionRepresentation
	for _, exec := range executions {
		modifyAuthRep = append(modifyAuthRep, helpers.ToModifyAutheExecutionRepModel(exec))
	}
	return modifyAuthRep, nil
}

func (k *KeycloakRealmRepository) updateRealmUserProfileMetadata(ctx context.Context, realmName, accessToken, authURL string) error {
	userProfileMetadata, err := helpers.GetUserProfileMetadata(ctx, realmName, accessToken, authURL)
	if err != nil {
		return err
	}
	if userProfileMetadata != nil {
		upAttributesCopy := userProfileMetadata.Attributes
		upAttributesCopy = enrichUserProfileAttributes(upAttributesCopy)
		if len(userProfileMetadata.Attributes) < len(upAttributesCopy) {
			userProfileMetadata.Attributes = upAttributesCopy
			err := helpers.UpdateUserProfileMetadata(ctx, realmName, accessToken, authURL, *userProfileMetadata)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (k *KeycloakRealmRepository) SetRealmEmailScopeOptional(ctx context.Context, realm string) error {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	optionalScopes, err := k.kcClient.GetDefaultOptionalClientScopes(ctx, *token, realm)
	if err != nil {
		return fmt.Errorf("failed to get realm optional scopes: %w", err)
	}
	for _, s := range optionalScopes {
		if s.Name != nil && *s.Name == utils.EmailScope {
			return nil
		}
	}
	allScopes, err := k.runaicloakClient.GetClientScopes(ctx, realm, *token)
	if err != nil {
		return fmt.Errorf("failed to get realm client scopes: %w", err)
	}
	emailScopeID := ""
	for _, s := range allScopes {
		if name, ok := s["name"].(string); ok && name == utils.EmailScope {
			if id, ok := s["id"].(string); ok {
				emailScopeID = id
			}
			break
		}
	}
	if emailScopeID == "" {
		return fmt.Errorf("email scope not found in realm %s", realm)
	}
	if err := k.runaicloakClient.DeleteDefaultClientScopeById(ctx, realm, *token, emailScopeID); err != nil {
		return fmt.Errorf("failed to remove email from realm default scopes: %w", err)
	}
	if err := k.runaicloakClient.AddOptionalClientScopeById(ctx, realm, *token, emailScopeID); err != nil {
		return fmt.Errorf("failed to add email to realm optional scopes: %w", err)
	}
	return nil
}

func enrichUserProfileAttributes(userProfileAttributes []models.UserProfileAttributeMetadata) []models.UserProfileAttributeMetadata {
	existingAttributesMap := getAttributeNameMap(userProfileAttributes)
	customAttributes := getCustomAttributes()
	for _, customAttribute := range customAttributes {
		if shouldAddCustomAttribute(customAttribute, existingAttributesMap) {
			userProfileAttributes = append(userProfileAttributes, customAttribute)
		}
	}
	return userProfileAttributes
}

func getAttributeNameMap(userProfileAttributes []models.UserProfileAttributeMetadata) map[string]struct{} {
	attributeNames := map[string]struct{}{}
	for _, attribute := range userProfileAttributes {
		if attribute.Name != nil {
			attributeNames[*attribute.Name] = struct{}{}
		}
	}
	return attributeNames
}

func isAttributeAlreadyExists(attName string, attributeNamesMap map[string]struct{}) bool {
	_, exists := attributeNamesMap[attName]
	return exists
}

func shouldAddCustomAttribute(customAttribute models.UserProfileAttributeMetadata, existingAttributesMap map[string]struct{}) bool {
	return customAttribute.Name != nil && !isAttributeAlreadyExists(*customAttribute.Name, existingAttributesMap)
}

func getCustomAttributes() []models.UserProfileAttributeMetadata {
	userProfileAttributePermission := getUserProfileAttributePermission()
	return []models.UserProfileAttributeMetadata{
		{
			Name:        gocloak.StringP(utils.IsLocalUser),
			DisplayName: gocloak.StringP(utils.IsLocalUser),
			Multivalued: gocloak.BoolP(false),
			Permissions: userProfileAttributePermission,
		},
		{
			Name:        gocloak.StringP(utils.LastLoginTime),
			DisplayName: gocloak.StringP(utils.LastLoginTime),
			Multivalued: gocloak.BoolP(false),
			Permissions: userProfileAttributePermission,
		},
		{
			Name:        gocloak.StringP(utils.CreationTime),
			DisplayName: gocloak.StringP(utils.CreationTime),
			Multivalued: gocloak.BoolP(false),
			Permissions: userProfileAttributePermission,
		},
		{
			Name:        gocloak.StringP(utils.UpdateTime),
			DisplayName: gocloak.StringP(utils.UpdateTime),
			Multivalued: gocloak.BoolP(false),
			Permissions: userProfileAttributePermission,
		},
		{
			Name:        gocloak.StringP(utils.Upn),
			DisplayName: gocloak.StringP(utils.Upn),
			Multivalued: gocloak.BoolP(false),
			Permissions: userProfileAttributePermission,
		},
		{
			Name:        gocloak.StringP(utils.UID),
			DisplayName: gocloak.StringP(utils.UID),
			Multivalued: gocloak.BoolP(false),
			Permissions: userProfileAttributePermission,
		},
		{
			Name:        gocloak.StringP(utils.GID),
			DisplayName: gocloak.StringP(utils.GID),
			Multivalued: gocloak.BoolP(false),
			Permissions: userProfileAttributePermission,
		},
		{
			Name:        gocloak.StringP(utils.Groups),
			DisplayName: gocloak.StringP(utils.Groups),
			Multivalued: gocloak.BoolP(true),
			Permissions: userProfileAttributePermission,
		},
		{
			Name:        gocloak.StringP(utils.SupplementaryGroups),
			DisplayName: gocloak.StringP(utils.SupplementaryGroups),
			Multivalued: gocloak.BoolP(true),
			Permissions: userProfileAttributePermission,
		},
		{
			Name:        gocloak.StringP(utils.CreatedBy),
			DisplayName: gocloak.StringP(utils.CreatedBy),
			Multivalued: gocloak.BoolP(false),
			Permissions: userProfileAttributePermission,
		},
	}
}

func getUserProfileAttributePermission() models.UserProfileAttributePermissions {
	return models.UserProfileAttributePermissions{
		View: []string{
			"admin", "user",
		},
		Edit: []string{
			"admin", "user",
		},
	}
}
