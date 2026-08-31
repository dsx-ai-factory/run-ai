/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package realms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Nerzal/gocloak/v13"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
)

const (
	passwordPolicy = "digits(1) and lowerCase(1) and upperCase(1) and length(8) and specialChars(1) " +
		"and forceExpiredPasswordChange(365) and notUsername(undefined) and passwordHistory(10) and notEmail(undefined)"

	SsoSessionMaxLifespanSeconds = 86400 // 1 day
)

func ToRealmModel(realm *gocloak.RealmRepresentation) *models.Realm {
	kcRealm := &models.Realm{
		Name:                 realm.Realm,
		Enabled:              realm.Enabled,
		RememberMe:           realm.RememberMe,
		ResetPasswordAllowed: realm.ResetPasswordAllowed,
		EnabledEventTypes:    realm.EnabledEventTypes,
		EventsListeners:      realm.EventsListeners,
		EventsEnabled:        realm.EventsEnabled,
		EventsExpiration:     realm.EventsExpiration,
		LoginTheme:           realm.LoginTheme,
		EmailTheme:           realm.EmailTheme,
		SMTPServer:           realm.SMTPServer,
	}
	if realm.ID != nil {
		kcRealm.ID = realm.ID
	}
	return kcRealm
}

func FromRealmModel(realm *models.Realm, kcRealm *gocloak.RealmRepresentation) *gocloak.RealmRepresentation {
	setDefaultKcRealmSettings(kcRealm)
	kcRealm.Enabled = realm.Enabled
	kcRealm.Realm = realm.Name
	kcRealm.RememberMe = realm.RememberMe
	kcRealm.ResetPasswordAllowed = realm.ResetPasswordAllowed
	kcRealm.EnabledEventTypes = realm.EnabledEventTypes
	kcRealm.EventsListeners = realm.EventsListeners
	kcRealm.EventsEnabled = realm.EventsEnabled
	kcRealm.DisplayName = realm.Name
	kcRealm.EventsExpiration = realm.EventsExpiration
	kcRealm.LoginTheme = realm.LoginTheme
	kcRealm.EmailTheme = realm.EmailTheme
	kcRealm.SMTPServer = realm.SMTPServer
	if realm.TenantId != nil {
		attributes := map[string]string{}
		if kcRealm.Attributes != nil {
			attributes = *kcRealm.Attributes
		}
		attributes[models.TenantIdRealmAttribute] = *realm.TenantId
		kcRealm.Attributes = &attributes
	}
	if realm.ID != nil {
		kcRealm.ID = realm.ID
	}
	return kcRealm
}

func ToModifyAutheExecutionRepModel(kcModifyAuthRep *gocloak.ModifyAuthenticationExecutionRepresentation) *models.ModifyAuthenticationExecutionRepresentation {
	modifyAuthRep := &models.ModifyAuthenticationExecutionRepresentation{
		ID:                   kcModifyAuthRep.ID,
		ProviderID:           kcModifyAuthRep.ProviderID,
		AuthenticationConfig: kcModifyAuthRep.AuthenticationConfig,
		AuthenticationFlow:   kcModifyAuthRep.AuthenticationFlow,
		Requirement:          kcModifyAuthRep.Requirement,
		FlowID:               kcModifyAuthRep.FlowID,
		DisplayName:          kcModifyAuthRep.DisplayName,
		Alias:                kcModifyAuthRep.Alias,
		RequirementChoices:   kcModifyAuthRep.RequirementChoices,
		Configurable:         kcModifyAuthRep.Configurable,
		Level:                kcModifyAuthRep.Level,
		Index:                kcModifyAuthRep.Index,
		Description:          kcModifyAuthRep.Description,
	}
	return modifyAuthRep
}

func FromModifyAutheExecutionRepModel(modifyAuthRep *models.ModifyAuthenticationExecutionRepresentation) *gocloak.ModifyAuthenticationExecutionRepresentation {
	kcModifyAuthRep := &gocloak.ModifyAuthenticationExecutionRepresentation{
		ID:                   modifyAuthRep.ID,
		ProviderID:           modifyAuthRep.ProviderID,
		AuthenticationConfig: modifyAuthRep.AuthenticationConfig,
		AuthenticationFlow:   modifyAuthRep.AuthenticationFlow,
		Requirement:          modifyAuthRep.Requirement,
		FlowID:               modifyAuthRep.FlowID,
		DisplayName:          modifyAuthRep.DisplayName,
		Alias:                modifyAuthRep.Alias,
		RequirementChoices:   modifyAuthRep.RequirementChoices,
		Configurable:         modifyAuthRep.Configurable,
		Level:                modifyAuthRep.Level,
		Index:                modifyAuthRep.Index,
		Description:          modifyAuthRep.Description,
	}
	return kcModifyAuthRep
}

func GetUserProfileMetadata(ctx context.Context, realmName, accessToken, keycloakURL string) (*models.UserProfileMetadata, error) {
	url := fmt.Sprintf("%s/admin/realms/%s/users/profile", keycloakURL, realmName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user profile metadata: %s", resp.Status)
	}
	var metadata models.UserProfileMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, err
	}
	return &metadata, nil
}

func UpdateUserProfileMetadata(ctx context.Context, realmName, accessToken, keycloakURL string, userProfileMetadata models.UserProfileMetadata) error {
	url := fmt.Sprintf("%s/admin/realms/%s/users/profile", keycloakURL, realmName)

	jsonData, err := json.Marshal(userProfileMetadata)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update user profile metadata: %s, %s", resp.Status, body)
	}
	return nil
}

func setDefaultKcRealmSettings(realm *gocloak.RealmRepresentation) {
	// General
	realm.DisplayName = realm.Realm
	realm.LoginTheme = gocloak.StringP("runai-theme")
	realm.EmailTheme = gocloak.StringP("runai-theme")
	realm.SslRequired = gocloak.StringP("external")
	// Login
	realm.RegistrationAllowed = gocloak.BoolP(false)
	realm.RegistrationEmailAsUsername = gocloak.BoolP(false)
	realm.EditUsernameAllowed = gocloak.BoolP(false)
	realm.LoginWithEmailAllowed = gocloak.BoolP(true)
	realm.DuplicateEmailsAllowed = gocloak.BoolP(false)
	realm.VerifyEmail = gocloak.BoolP(false)
	// Events
	realm.AdminEventsEnabled = gocloak.BoolP(false)
	// Security Defenses
	realm.BruteForceProtected = gocloak.BoolP(true)
	realm.MaxFailureWaitSeconds = gocloak.IntP(900)          // 15 minutes
	realm.MinimumQuickLoginWaitSeconds = gocloak.IntP(60)    // 1 minute
	realm.WaitIncrementSeconds = gocloak.IntP(60)            // 1 minute
	realm.QuickLoginCheckMilliSeconds = gocloak.Int64P(1000) // 1 second
	realm.MaxDeltaTimeSeconds = gocloak.IntP(43200)          // 12 hours
	realm.FailureFactor = gocloak.IntP(10)
	realm.PasswordPolicy = gocloak.StringP(passwordPolicy)
	// Token
	realm.AccessTokenLifespan = gocloak.IntP(900)                   // 15 minutes
	realm.AccessTokenLifespanForImplicitFlow = gocloak.IntP(900)    // 15 minutes
	realm.AccessCodeLifespan = gocloak.IntP(60)                     // 1 minute
	realm.AccessCodeLifespanLogin = gocloak.IntP(2592000)           // 30 days
	realm.AccessCodeLifespanUserAction = gocloak.IntP(300)          // 5 minutes
	realm.ActionTokenGeneratedByUserLifespan = gocloak.IntP(300)    // 5 minutes
	realm.ActionTokenGeneratedByAdminLifespan = gocloak.IntP(43200) // 12 hours
	realm.EventsExpiration = gocloak.Int64P(90 * 24 * 60 * 60)      // 90 days
	realm.RevokeRefreshToken = gocloak.BoolP(true)
	realm.RefreshTokenMaxReuse = gocloak.IntP(1)
	// Sessions
	realm.SsoSessionIdleTimeout = gocloak.IntP(SsoSessionMaxLifespanSeconds)
	realm.SsoSessionMaxLifespan = gocloak.IntP(SsoSessionMaxLifespanSeconds)
	realm.OfflineSessionMaxLifespan = gocloak.IntP(2592000) // 30 days
	realm.OfflineSessionMaxLifespanEnabled = gocloak.BoolP(true)
}
