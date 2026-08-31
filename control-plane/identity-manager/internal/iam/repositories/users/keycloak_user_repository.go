/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package users

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/errors"

	"github.com/Nerzal/gocloak/v13"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/auth"
	gocloakInterface "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/gocloak"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	helpers "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/repositories/users/helpers"
	runaicloak "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/runaicloak"
)

const (
	Username       = "username"
	Exact          = "exact"
	Q              = "q"
	Max            = "max"
	Offset         = "offset"
	SortBy         = "sortBy"
	SortOrder      = "sortOrder"
	CreationFilter = "creationFilter"
	LoginFilter    = "loginFilter"
	UpdateFilter   = "updateFilter"
	CreatedBy      = "createdBy"
	Search         = "search"
	MinimalSearch  = "minimalSearch"
	FullName       = "fullName"
)

const (
	UnmanagedAttributesKey = "unmanagedAttributePolicy"
)

type UsersResponse struct {
	Count int
	Users []map[string]string
}

type KeycloakUserRepository struct {
	auth       auth.Auth
	kcClient   gocloakInterface.GoCloak
	runaicloak runaicloak.RunAICloakClient
}

func NewKeycloakUserRepository(client *auth.KeycloakAuthenticator, runaicloak runaicloak.RunAICloakClient) *KeycloakUserRepository {
	return &KeycloakUserRepository{
		auth:       client,
		kcClient:   client.KcClient,
		runaicloak: runaicloak,
	}
}

func (k *KeycloakUserRepository) GetUsers(ctx context.Context, realm string, userFilters models.UserFilters) ([]*models.User, error) {
	params := setQueryParamMap(userFilters)
	usersResponse, err := k.getCustomUsers(ctx, realm, params)
	if err != nil {
		return nil, err
	}
	var users []*models.User
	for _, user := range usersResponse.Users {
		if user != nil {
			users = append(users, helpers.ToUserModelFromCustomUser(user))
		}
	}
	return users, nil
}

func (k *KeycloakUserRepository) CreateUser(ctx context.Context, realm string, user *models.User) (string, error) {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return "", errors.HandleKeyCloakError(err)
	}
	kcUser := &gocloak.User{}
	kcUser = helpers.FromUserModel(user, kcUser)
	createdUserID, err := k.kcClient.CreateUser(ctx, *token, realm, *kcUser)
	if err != nil {
		return "", errors.HandleKeyCloakError(err)
	}
	return createdUserID, nil
}

func (k *KeycloakUserRepository) GetUserByID(ctx context.Context, realm string, userID string) (*models.User, error) {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return nil, errors.HandleKeyCloakError(err)
	}
	internalUser, err := k.kcClient.GetUserByID(ctx, *token, realm, userID)
	if err != nil {
		return nil, errors.HandleKeyCloakError(err)
	}
	return helpers.ToUserModel(*internalUser), nil
}

func (k *KeycloakUserRepository) GetUserByEmail(ctx context.Context, realm string, email string) (*models.User, error) {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return nil, errors.HandleKeyCloakError(err)
	}
	params := gocloak.GetUsersParams{
		Email: &email,
	}
	userNotFoundErr := fmt.Errorf("user with email %s not found", email)
	kcUser, err := k.kcClient.GetUsers(ctx, *token, realm, params)
	if err != nil {
		return nil, errors.HandleKeyCloakError(err)
	}
	if len(kcUser) == 0 {
		return nil, userNotFoundErr
	}
	for _, user := range kcUser {
		if user.Email != nil {
			if *user.Email == email {
				return helpers.ToUserModel(*user), nil
			}
		}
	}
	return nil, userNotFoundErr
}

func (k *KeycloakUserRepository) UpdateUser(ctx context.Context, realm string, userID string, user *models.User) error {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	existing, err := k.kcClient.GetUserByID(ctx, *token, realm, userID)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	existing = helpers.ApplyUserPatch(existing, user)
	err = k.kcClient.UpdateUser(ctx, *token, realm, *existing)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	return nil
}

func (k *KeycloakUserRepository) DeleteUser(ctx context.Context, realm string, userID string) error {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	err = k.kcClient.DeleteUser(ctx, *token, realm, userID)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	return nil
}

func (k *KeycloakUserRepository) LogoutUser(ctx context.Context, realm string, userID string) error {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	err = k.kcClient.LogoutAllSessions(ctx, *token, realm, userID)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	return nil
}

func (k *KeycloakUserRepository) LogoutUserBySessionID(ctx context.Context, realm string, sessionID string) error {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	err = k.kcClient.LogoutUserSession(ctx, *token, realm, sessionID)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	return nil
}

func (k *KeycloakUserRepository) ChangeUserPassword(ctx context.Context, realm string, user *models.User, currentPassword string, newPassword string) error {
	jwt, err := k.kcClient.Login(ctx, "runai-cli", "", realm, *user.Username, currentPassword)
	if jwt == nil || err != nil {
		return errors.HandleKeyCloakError(err)
	}
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return err
	}
	err = k.kcClient.SetPassword(ctx, *token, *user.ID, realm, newPassword, false)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	return nil
}

func (k *KeycloakUserRepository) ResetUserPassword(ctx context.Context, realm string, userID string, password string) error {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	err = k.kcClient.SetPassword(ctx, *token, userID, realm, password, true)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	return nil
}

func (k *KeycloakUserRepository) SendPasswordResetEmail(ctx context.Context, realm string, userID string, expirationSeconds int, redirectURI string, clientId string) error {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}

	params := gocloak.ExecuteActionsEmail{
		UserID:      &userID,
		Actions:     &[]string{helpers.UPDATE_PASSWORD_ACTION},
		Lifespan:    gocloak.IntP(expirationSeconds),
		ClientID:    gocloak.StringP(clientId),
		RedirectURI: gocloak.StringP(redirectURI),
	}

	err = k.kcClient.ExecuteActionsEmail(ctx, *token, realm, params)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	return nil
}

func (k *KeycloakUserRepository) ToggleRequiredAction(ctx context.Context, realm string, actionAlias models.RequiredAction, enabled bool) error {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	action, err := k.kcClient.GetRequiredAction(ctx, *token, realm, string(actionAlias))
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	action.Enabled = &enabled
	err = k.kcClient.UpdateRequiredAction(ctx, *token, realm, *action)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	return nil
}

func (k *KeycloakUserRepository) GetRequiredAction(ctx context.Context, realm string, actionAlias string) (*models.RequiredUserAction, error) {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return nil, errors.HandleKeyCloakError(err)
	}
	action, err := k.kcClient.GetRequiredAction(ctx, *token, realm, actionAlias)
	if err != nil {
		return nil, errors.HandleKeyCloakError(err)
	}
	return &models.RequiredUserAction{
		Alias:         action.Alias,
		Config:        action.Config,
		DefaultAction: action.DefaultAction,
		Enabled:       action.Enabled,
		Name:          action.Name,
		Priority:      action.Priority,
		ProviderID:    action.ProviderID,
	}, nil
}

func (k *KeycloakUserRepository) ToggleUnmanagedUsersAttributes(ctx context.Context, realm string, enabled bool) error {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return err
	}
	attributesJson, err := k.runaicloak.GetRealmUserAttributes(ctx, realm, *token)
	if err != nil {
		return err
	}
	return k.runaicloak.ToggleUserAttribute(ctx, realm, *token, attributesJson, UnmanagedAttributesKey, enabled)
}

func (k *KeycloakUserRepository) getCustomUsers(ctx context.Context, realm string, queryParams map[string]string) (*UsersResponse, error) {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return nil, errors.HandleKeyCloakError(err)
	}
	url := fmt.Sprintf("%s/admin/realms/%s/runaienteties/users", k.auth.GetHost(), realm)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if token != nil {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *token))
	}

	query := req.URL.Query()
	for key, value := range queryParams {
		query.Add(key, value)
	}
	req.URL.RawQuery = query.Encode()

	httpClient := http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	var usersResponse UsersResponse
	if err != nil {
		return nil, err
	} else {
		if err := json.Unmarshal(body, &usersResponse); err != nil {
			return nil, err
		}
	}
	return &usersResponse, nil
}

func setQueryParamMap(userFilters models.UserFilters) map[string]string {
	queryMap := make(map[string]string)
	if userFilters.Username != nil {
		queryMap[Username] = *userFilters.Username
		if userFilters.UsernameExact != nil && *userFilters.UsernameExact {
			queryMap[Exact] = "true"
		}
	}
	if userFilters.Q != nil {
		queryMap[Q] = *userFilters.Q
	}
	if userFilters.Max != nil {
		queryMap[Max] = strconv.Itoa(*userFilters.Max)
	}
	if userFilters.Offset != nil {
		queryMap[Offset] = strconv.Itoa(*userFilters.Offset)
	}
	if userFilters.SortBy != nil {
		queryMap[SortBy] = *userFilters.SortBy
	}
	if userFilters.SortOrder != nil {
		queryMap[SortOrder] = *userFilters.SortOrder
	}
	if userFilters.CreationFilter != nil {
		queryMap[CreationFilter] = *userFilters.CreationFilter
	}
	if userFilters.LoginFilter != nil {
		queryMap[LoginFilter] = *userFilters.LoginFilter
	}
	if userFilters.UpdateFilter != nil {
		queryMap[UpdateFilter] = *userFilters.UpdateFilter
	}
	if userFilters.CreatedBy != nil {
		queryMap[CreatedBy] = *userFilters.CreatedBy
	}
	if userFilters.Search != nil {
		queryMap[Search] = *userFilters.Search
	}
	if userFilters.MinimalSearch != nil && *userFilters.MinimalSearch {
		queryMap[MinimalSearch] = "true"
	}
	if userFilters.FullName != nil {
		queryMap[FullName] = *userFilters.FullName
	}
	return queryMap
}
