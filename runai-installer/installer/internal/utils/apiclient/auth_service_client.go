/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package apiclient

import (
	"context"
	"fmt"
	"net/http"

	authorization "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	idmclient "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
	runaiclient "github.com/dsx-ai-factory/run-ai/runai-installer/clients/pkg/clients"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
)

type UserCredentials struct {
	Email    string
	Password string
	UserID   string
}

type AuthServiceClient struct {
	client      *runaiclient.RunAIClient
	rolesClient *RolesHTTPClient
}

func NewAuthServiceClientFromRunAI(client *runaiclient.RunAIClient) *AuthServiceClient {
	return &AuthServiceClient{
		client: client,
	}
}

func NewAuthServiceClient(client *runaiclient.RunAIClient, rolesClient *RolesHTTPClient) *AuthServiceClient {
	return &AuthServiceClient{
		client:      client,
		rolesClient: rolesClient,
	}
}

func (c *AuthServiceClient) SetRolesHTTPClient(rolesClient *RolesHTTPClient) {
	c.rolesClient = rolesClient
}

func (c *AuthServiceClient) GetRoles(ctx context.Context) ([]Role, int, error) {
	logging.Infof("[AuthService] Fetching roles using HTTP client")

	if c.rolesClient == nil {
		return nil, 0, fmt.Errorf("roles HTTP client not initialized")
	}

	return c.rolesClient.GetRoles(ctx)
}

func (c *AuthServiceClient) GetRoleIDByName(ctx context.Context, roleName string) (int32, error) {
	logging.Infof("[AuthService] Getting role ID for role: %s", roleName)

	roles, _, err := c.GetRoles(ctx)
	if err != nil {
		return 0, err
	}

	for _, role := range roles {
		if role.Name == roleName {
			logging.Infof("[AuthService] Found role '%s' with ID: %d", roleName, role.ID)
			return role.ID, nil
		}
	}

	return 0, fmt.Errorf("role '%s' not found", roleName)
}

func (c *AuthServiceClient) CreateAccessRule(ctx context.Context, userEmail string, roleID int32, clusterUUID string, projectID string) (int, error) {
	logging.Infof("[AuthService] Creating access rule for user %s with role ID %d on project %s", userEmail, roleID, projectID)

	accessRuleRequest := authorization.AccessRuleCreationFields{
		SubjectId:   userEmail,
		SubjectType: authorization.SUBJECTTYPE_USER,
		RoleId:      roleID,
		ScopeId:     projectID,
		ScopeType:   authorization.SCOPETYPE_PROJECT,
		ClusterId:   &clusterUUID,
	}

	_, resp, err := c.client.AccessRulesAPI().CreateAccessRule(ctx).AccessRuleCreationFields(accessRuleRequest).Execute()
	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode
	}

	if err != nil {
		logging.Errorf("[AuthService] Failed to create access rule: %v", err)
		return statusCode, fmt.Errorf("failed to create access rule: %w", err)
	}

	logging.Infof("[AuthService] Successfully created access rule for user %s on project %s", userEmail, projectID)
	return statusCode, nil
}

func (c *AuthServiceClient) GetUserByEmail(ctx context.Context, email string) (*idmclient.User, error) {
	logging.Infof("[AuthService] Checking if user exists with email: %s", email)
	filterString := []string{fmt.Sprintf("username=@%s", email)}

	users, resp, err := c.client.IdmUsersAPI().GetUsers(ctx).FilterBy(filterString).Limit(1).Execute()
	if err != nil {
		statusCode := 0
		if resp != nil {
			statusCode = resp.StatusCode
		}
		logging.Errorf("[AuthService] Failed to get users from IDM: %v (status: %d)", err, statusCode)
		return nil, fmt.Errorf("failed to get users from IDM: %w", err)
	}

	for _, user := range users {
		if user.Username == email {
			logging.Infof("[AuthService] Found existing user with email: %s (ID: %s)", email, user.Id)
			return &user, nil
		}
	}

	logging.Infof("[AuthService] No user found with email: %s", email)
	return nil, nil
}

func (c *AuthServiceClient) AccessRuleExists(ctx context.Context, userEmail string, roleID int32, clusterUUID string, projectID string) (bool, error) {
	logging.Infof("[AuthService] Checking if access rule exists for user %s with role ID %d on project %s", userEmail, roleID, projectID)

	filterString := []string{fmt.Sprintf("subjectId=@%s", userEmail)}
	accessRulesResp, resp, err := c.client.AccessRulesAPI().GetAccessRules(ctx).
		FilterBy(filterString).Execute()

	if err != nil {
		statusCode := 0
		if resp != nil {
			statusCode = resp.StatusCode
		}
		logging.Errorf("[AuthService] Failed to get access rules: %v (status: %d)", err, statusCode)
		return false, fmt.Errorf("failed to get access rules: %w", err)
	}

	exists := len(accessRulesResp.GetAccessRules()) > 0
	if exists {
		logging.Infof("[AuthService] Access rule already exists for user %s on project %s", userEmail, projectID)
	} else {
		logging.Infof("[AuthService] No access rule found for user %s on project %s", userEmail, projectID)
	}
	return exists, nil
}

func (c *AuthServiceClient) CreateLocalUser(ctx context.Context, email string) (*UserCredentials, bool, error) {
	logging.Infof("[AuthService] Creating local user with email: %s via IDM service", email)

	userRequest := idmclient.NewUserCreationRequest(email)

	userResp, resp, err := c.client.IdmUsersAPI().CreateUser(ctx).UserCreationRequest(*userRequest).Execute()
	if err != nil {

		if resp != nil && resp.StatusCode == http.StatusConflict {
			logging.Infof("[AuthService] User %s already exists, continuing", email)
			return &UserCredentials{Email: email, Password: ""}, true, nil
		}
		logging.Errorf("[AuthService] Failed to create user via IDM: %v", err)
		return nil, false, fmt.Errorf("failed to create user via IDM: %w", err)
	}

	logging.Infof("[AuthService] Successfully created local user via IDM: %s (ID: %s)", email, userResp.Id)
	return &UserCredentials{
		Email:    userResp.Username,
		Password: userResp.TempPassword,
		UserID:   userResp.Id,
	}, false, nil
}
