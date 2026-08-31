/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	"fmt"
	"strings"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/run-ai/runai/control-plane/authorization-migrator/keycloak"
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm/clause"
)

const (
	EntityTypeApp         = "app"
	EntityTypeGroup       = "group"
	EntityTypeRegularUser = "regular-user"
	EntityTypeSSOUser     = "sso-user"
	RunaiAgentPrefix      = "runai_agent_"
)

// keycloak attributes
const (
	IS_LOCAL_USER   = "runai.is_local"
	IS_EXTERNAL_APP = "runai.is_external"
	CREATION_TIME   = "runai.creation.time"
	UPDATE_TIME     = "runai.update.time"
	CREATED_BY      = "runai.created.by"
)

// string forms of boolean keycloak attribute values
const (
	attrTrue  = "true"
	attrFalse = "false"
)

type RoleAssignmentRow struct {
	UserId                string
	ClientId              *string
	TenantId              int
	AuthRealm             *string
	EntityType            string
	CreatedBy             *string
	CreatedAt             *time.Time
	UpdatedAt             *time.Time
	AllClustersPermitted  *bool
	Role                  *string
	ClusterUuid           *string
	DepartmentId          *int
	ProjectId             *int
	ClusterName           *string
	DepartmentName        *string
	DepartmentClusterUuid *string
	ProjectName           *string
	ProjectClusterUuid    *string
	TenantName            *string
}

type TenantRow struct {
	TenantId   int
	TenantName string
}

func CreateAccessRule(db orm.DB, tenantId int, subjectId string, entityType string, roleName string, scopeId string, scopeType api.ScopeType, scopeName string, clusterUuid string) {
	accessRuleSpec := models.AccessRule{
		SubjectId:   subjectId,
		SubjectType: api.SubjectTypeExtended(getSubjectType(entityType)),
		RoleID:      getRoleId(db, roleName),
		ScopeId:     scopeId,
		ScopeType:   scopeType,
		ScopeName:   scopeName,
		TenantId:    uint(tenantId),
		ClusterUuid: clusterUuid,
	}

	if _, err := db.
		Clauses(clause.OnConflict{
			DoNothing: true,
		}).
		Create(&accessRuleSpec); err != nil {
		logging.Errorf("Failed to create access rule. err: %+v", err)
		panic(err)
	}
}

func getRoleId(db orm.DB, name string) uint {
	role := &models.Role{}
	if err := db.Select("id").Where("name = ?", name).First(role); err != nil {
		logging.Errorf(fmt.Sprintf("Failed to get role id of role %s", name), err)
		panic(err)
	}

	return role.ID
}

func getSubjectType(entityType string) api.SubjectType {
	switch entityType {
	case EntityTypeApp:
		return api.SUBJECTTYPE_APP
	case EntityTypeGroup:
		return api.SUBJECTTYPE_GROUP
	default:
		return api.SUBJECTTYPE_USER
	}
}

func MigrateLocalUser(row RoleAssignmentRow, authRealm string) (*gocloak.User, error) {
	user, err := keycloak.GetKeycloakUserById(row.UserId, authRealm)
	if err != nil {
		logging.Errorf("failed getting user with id %s from keycloak, err: %+v", row.UserId, err)
		return nil, err
	}
	updateLocalUserAttributes(row, user, authRealm)
	return user, nil
}

func MigrateSsoUser(row RoleAssignmentRow, authRealm string) {
	var keycloakSsoUser *gocloak.User
	usersByEmail, _ := keycloak.GetUsersWithFilter(row.UserId, authRealm, gocloak.GetUsersParams{Email: &row.UserId})
	if len(usersByEmail) > 0 {
		keycloakSsoUser = usersByEmail[0]
	} else {
		usersByUsername, _ := keycloak.GetUsersWithFilter(row.UserId, authRealm, gocloak.GetUsersParams{Username: &row.UserId})
		if len(usersByUsername) > 0 {
			keycloakSsoUser = usersByUsername[0]
		}
	}
	if keycloakSsoUser == nil {
		logging.Infof("failed to find sso user with userId (email/username) %s in keycloak", row.UserId)
		return
	}
	updateSSOUserAttributes(keycloakSsoUser, authRealm)
}

func MigrateApplication(row RoleAssignmentRow, authRealm string) (*gocloak.Client, error) {
	if row.ClientId == nil {
		logging.Warnf("Application with user id %s has no client id in the backend db, cant find it in keycloak therefore skipping", row.UserId)
		return nil, fmt.Errorf("application with user id %s has no client id", row.UserId)
	}

	application, err := keycloak.GetAppByKeyCloakId(*row.ClientId, authRealm)
	if err != nil {
		logging.Errorf("failed getting application with client id %s from keycloak", *row.ClientId)
		return nil, err
	}
	if application == nil {
		logging.Infof("failed to find application with client id %s in keycloak", *row.ClientId)
		return nil, fmt.Errorf("failed to find application with client id %s in keycloak", *row.ClientId)
	}
	if isInternalClient(*application.ClientID) {
		updateAgentAttributes(row, application, authRealm)
		return application, nil
	}
	updateApplicationAttributes(row, application, authRealm)

	return application, nil
}

func updateLocalUserAttributes(row RoleAssignmentRow, user *gocloak.User, authRealm string) {
	shouldUpdate := false
	createdByUser := getCreatedByUser(row, authRealm)
	if user.Attributes == nil {
		shouldUpdate = true
		user.Attributes = &map[string][]string{}
	}

	if len((*user.Attributes)[IS_LOCAL_USER]) == 0 || (*user.Attributes)[IS_LOCAL_USER][0] != attrTrue {
		shouldUpdate = true
		(*user.Attributes)[IS_LOCAL_USER] = []string{attrTrue}
	}
	if shouldUpdateCreationTime(row, user) {
		shouldUpdate = true
		(*user.Attributes)[CREATION_TIME] = []string{fmt.Sprint(row.CreatedAt.Unix())}
	}
	if shouldUpdateUpdateTime(row, user) {
		shouldUpdate = true
		(*user.Attributes)[UPDATE_TIME] = []string{fmt.Sprint(row.UpdatedAt.Unix())}
	}
	if shouldUpdateCreatedBy(row, user, createdByUser) {
		shouldUpdate = true
		(*user.Attributes)[CREATED_BY] = []string{*createdByUser.Email}
	}
	if shouldUpdate {
		keycloak.UpdateUser(*user, authRealm)
	}
}

func updateSSOUserAttributes(ssoUser *gocloak.User, authRealm string) {
	shouldUpdate := false
	if ssoUser.Attributes == nil {
		shouldUpdate = true
		ssoUser.Attributes = &map[string][]string{}
	}

	if isUserMarkedLocal(ssoUser) {
		return
	}

	if shouldMarkUserAsSSO(ssoUser) {
		shouldUpdate = true
		(*ssoUser.Attributes)[IS_LOCAL_USER] = []string{attrFalse}
	}
	if shouldUpdate {
		keycloak.UpdateUser(*ssoUser, authRealm)
	}
}

func shouldMarkUserAsSSO(ssoUser *gocloak.User) bool {
	return len((*ssoUser.Attributes)[IS_LOCAL_USER]) == 0 || (*ssoUser.Attributes)[IS_LOCAL_USER][0] != attrFalse
}

func isUserMarkedLocal(ssoUser *gocloak.User) bool {
	return len((*ssoUser.Attributes)[IS_LOCAL_USER]) > 0 && (*ssoUser.Attributes)[IS_LOCAL_USER][0] == attrTrue
}

func updateAgentAttributes(row RoleAssignmentRow, agent *gocloak.Client, authRealm string) {
	updated := false

	if agent.Attributes == nil {
		updated = true
		agent.Attributes = &map[string]string{}
	}

	if (*agent.Attributes)[IS_EXTERNAL_APP] != attrFalse {
		updated = true
		(*agent.Attributes)[IS_EXTERNAL_APP] = attrFalse
	}

	if updated {
		keycloak.UpdateClient(*agent, authRealm)
	}
}

func updateApplicationAttributes(row RoleAssignmentRow, app *gocloak.Client, authRealm string) {
	updated := false
	createdByUser := getCreatedByUser(row, authRealm)
	if app.Attributes == nil {
		updated = true
		app.Attributes = &map[string]string{}
	}

	if (*app.Attributes)[IS_EXTERNAL_APP] != attrTrue {
		updated = true
		(*app.Attributes)[IS_EXTERNAL_APP] = attrTrue
	}
	if row.CreatedAt != nil && (*app.Attributes)[CREATION_TIME] != fmt.Sprint(row.CreatedAt.Unix()) {
		updated = true
		(*app.Attributes)[CREATION_TIME] = fmt.Sprint(row.CreatedAt.Unix())
	}
	if row.UpdatedAt != nil && (*app.Attributes)[UPDATE_TIME] != fmt.Sprint(row.UpdatedAt.Unix()) {
		updated = true
		(*app.Attributes)[UPDATE_TIME] = fmt.Sprint(row.UpdatedAt.Unix())
	}
	if createdByUser != nil && createdByUser.Email != nil && (*app.Attributes)[CREATED_BY] != *createdByUser.Email {
		updated = true
		(*app.Attributes)[CREATED_BY] = *createdByUser.Email
	}

	if updated {
		keycloak.UpdateClient(*app, authRealm)
	}
}

func getCreatedByUser(row RoleAssignmentRow, authRealm string) *gocloak.User {
	if row.CreatedBy == nil {
		logging.Infof("missing createdBy in db")
		return nil
	}
	createdByUser, err := keycloak.GetKeycloakUserById(*row.CreatedBy, authRealm)
	if err != nil {
		logging.Infof("can't find created by user in keycloak")
		return nil
	}
	return createdByUser
}

func isInternalClient(clientName string) bool {
	return strings.HasPrefix(clientName, RunaiAgentPrefix)
}

func shouldUpdateCreationTime(row RoleAssignmentRow, user *gocloak.User) bool {
	return row.CreatedAt != nil && (len((*user.Attributes)[CREATION_TIME]) == 0 || (*user.Attributes)[CREATION_TIME][0] != fmt.Sprint(row.CreatedAt.Unix()))
}

func shouldUpdateUpdateTime(row RoleAssignmentRow, user *gocloak.User) bool {
	return row.UpdatedAt != nil && (len((*user.Attributes)[UPDATE_TIME]) == 0 || (*user.Attributes)[UPDATE_TIME][0] != fmt.Sprint(row.UpdatedAt.Unix()))
}

func shouldUpdateCreatedBy(row RoleAssignmentRow, user *gocloak.User, createdByUser *gocloak.User) bool {
	return createdByUser != nil && createdByUser.Email != nil && (len((*user.Attributes)[CREATED_BY]) == 0 || (*user.Attributes)[CREATED_BY][0] != *createdByUser.Email)
}
