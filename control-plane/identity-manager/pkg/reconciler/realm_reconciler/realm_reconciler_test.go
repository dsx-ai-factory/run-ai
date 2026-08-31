/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package realm_reconciler

import (
	"context"
	"fmt"
	"testing"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"

	iamError "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/errors"
	iamMock "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/mocks"
	businessMocks "github.com/run-ai/runai/control-plane/identity-manager/pkg/business/mocks"
	pkgModels "github.com/run-ai/runai/control-plane/identity-manager/pkg/models"
	realmUpdaterMock "github.com/run-ai/runai/control-plane/identity-manager/pkg/realm/mocks"
	realmMock "github.com/run-ai/runai/control-plane/identity-manager/pkg/utils/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const testTenantId = "1001"

const (
	envDev            = "development"
	testSmtpHost      = "test-smtp-host"
	testSmtpPort      = 1234
	testSmtpFromEmail = "test@runai.com"
	testTheme         = "runai"
)

func initRealmReconcilerTest() (iamServicesContainerMock *iamMock.MockServicesContainer, testSubject *RealmReconciler) {
	mockIamServicesContainer := iamMock.NewMockServicesContainer()
	realmDalMock := &businessMocks.RealmDal{}
	realmDalMock.On("GetRealmByName", mock.Anything, mock.Anything).
		Return(&pkgModels.Realm{TenantId: testTenantId, Realm: realmMock.REALM_NAME}, nil)
	testSubject = &RealmReconciler{
		iamServices: *mockIamServicesContainer.ToServicesContainer(),
		config: RealmReconcilerConfig{
			env:           envDev,
			theme:         testTheme,
			smtpHost:      testSmtpHost,
			smtpPort:      testSmtpPort,
			smtpFromEmail: testSmtpFromEmail,
		},
		authFlowUpdater: &realmUpdaterMock.AuthFlowUpdaterInterface{},
		realmDal:        realmDalMock,
	}
	return mockIamServicesContainer, testSubject
}

func iamSetup(iamMock *iamMock.MockServicesContainer, realmName string, realmData *models.Realm, toggleRequiredActionErr error) {
	iamMock.MockUserService.On("ToggleUnmanagedUsersAttributes", context.Background(), realmName, true).Return(nil)
	iamMock.MockUserService.On("ToggleRequiredAction", context.Background(), realmName, mock.Anything, false).Return(toggleRequiredActionErr)

	iamMock.MockRealmService.On("GetRealm", mock.Anything, realmName).Return(realmData, nil)
	iamMock.MockRealmService.On("UpdateRealm", mock.Anything, mock.Anything).Return(nil)
	iamMock.MockRealmService.On("SetRealmEmailScopeOptional", mock.Anything, realmName).Return(nil)
}

func defaultAuthFlowUpdaterMockSetup(authFlowUpdaterMock *realmUpdaterMock.AuthFlowUpdaterInterface, realmName string) {
	authFlowUpdaterMock.On("UpdateAuthFlows", mock.Anything, realmName).Return(nil)
}

func TestRealmReconciler_Reconcile(t *testing.T) {
	iamMock, testSubject := initRealmReconcilerTest()
	realmName := realmMock.REALM_NAME
	realmData := realmMock.GetBasicRealmArray()[0]
	iamSetup(iamMock, realmName, realmData, nil)
	defaultAuthFlowUpdaterMockSetup(testSubject.authFlowUpdater.(*realmUpdaterMock.AuthFlowUpdaterInterface), realmName)
	testSubject.Reconcile(context.Background(), realmName, true)
	iamMock.MockRealmService.AssertNumberOfCalls(t, "GetRealm", 1)
	iamMock.MockRealmService.AssertNumberOfCalls(t, "UpdateRealm", 1)
}

func TestRealmReconciler_ReconcileRealmNotFound(t *testing.T) {
	iamMock, testSubject := initRealmReconcilerTest()
	realmName := realmMock.REALM_NAME
	iamMock.MockRealmService.On("GetRealm", mock.Anything, mock.Anything).Return(nil, iamError.ErrNotFound)
	iamMock.MockRealmService.On("CreateRealm", mock.Anything, mock.Anything).Return(realmName, nil)
	iamMock.MockRealmService.On("SetRealmEmailScopeOptional", mock.Anything, realmName).Return(nil)

	testSubject.Reconcile(context.Background(), realmName, true)
	iamMock.MockRealmService.AssertNumberOfCalls(t, "GetRealm", 1)
	iamMock.MockRealmService.AssertNumberOfCalls(t, "CreateRealm", 1)
	iamMock.MockRealmService.AssertNumberOfCalls(t, "SetRealmEmailScopeOptional", 1)
}

func TestRealmReconciler_ToggleRequiredActionNotFound(t *testing.T) {
	iamMock, testSubject := initRealmReconcilerTest()
	realmName := realmMock.REALM_NAME
	realmData := realmMock.GetBasicRealmArray()[0]
	iamSetup(iamMock, realmName, realmData, iamError.ErrNotFound)

	defaultAuthFlowUpdaterMockSetup(testSubject.authFlowUpdater.(*realmUpdaterMock.AuthFlowUpdaterInterface), realmName)
	testSubject.Reconcile(context.Background(), realmName, true)

	iamMock.MockUserService.AssertNumberOfCalls(t, "ToggleUnmanagedUsersAttributes", 1)
}

func TestRealmReconciler_ToggleRequiredActionErr(t *testing.T) {
	iamMock, testSubject := initRealmReconcilerTest()
	realmName := realmMock.REALM_NAME
	realmData := realmMock.GetBasicRealmArray()[0]
	e := fmt.Errorf("error")
	iamSetup(iamMock, realmName, realmData, e)

	defaultAuthFlowUpdaterMockSetup(testSubject.authFlowUpdater.(*realmUpdaterMock.AuthFlowUpdaterInterface), realmName)
	testSubject.Reconcile(context.Background(), realmName, true)

	iamMock.MockUserService.AssertNumberOfCalls(t, "ToggleUnmanagedUsersAttributes", 0)
}

func TestRealmReconciler_ConfigureSmtpSettings(t *testing.T) {
	_, testSubject := initRealmReconcilerTest()
	realmName := realmMock.REALM_NAME
	realmData := &models.Realm{
		Name: &realmName,
	}

	defaults := &models.RealmDefaults{
		Theme:         &testSubject.config.theme,
		SmtpHost:      testSubject.config.smtpHost,
		SmtpPort:      testSubject.config.smtpPort,
		SmtpFromEmail: testSubject.config.smtpFromEmail,
	}
	realmData.SetRealmDefaults(defaults)

	assert.NotNil(t, realmData.SMTPServer, "SMTPServer should not be nil")

	smtpSettings := *realmData.SMTPServer
	assert.Equal(t, testSmtpHost, smtpSettings["host"], "SMTP host should match")
	assert.Equal(t, "1234", smtpSettings["port"], "SMTP port should match")
	assert.Equal(t, testSmtpFromEmail, smtpSettings["from"], "SMTP from email should match")
	assert.Equal(t, "false", smtpSettings["starttls"], "SMTP starttls should be false")
	assert.Equal(t, "false", smtpSettings["auth"], "SMTP auth should be false")
	assert.Equal(t, "false", smtpSettings["ssl"], "SMTP ssl should be false")
}

func TestRealmReconciler_ReconcileExistingRealm_IncludesSmtpConfig(t *testing.T) {
	iamMock, testSubject := initRealmReconcilerTest()
	realmName := realmMock.REALM_NAME
	realmData := realmMock.GetBasicRealmArray()[0]

	iamMock.MockRealmService.On("GetRealm", mock.Anything, realmName).Return(realmData, nil)
	iamMock.MockUserService.On("ToggleRequiredAction", context.Background(), realmName, mock.Anything, false).Return(nil)
	iamMock.MockUserService.On("ToggleUnmanagedUsersAttributes", context.Background(), realmName, true).Return(nil)

	var updatedRealm *models.Realm
	iamMock.MockRealmService.On("UpdateRealm", mock.Anything, mock.MatchedBy(func(realm *models.Realm) bool {
		updatedRealm = realm
		return true
	})).Return(nil)

	defaultAuthFlowUpdaterMockSetup(testSubject.authFlowUpdater.(*realmUpdaterMock.AuthFlowUpdaterInterface), realmName)
	iamMock.MockRealmService.On("SetRealmEmailScopeOptional", mock.Anything, realmName).Return(nil)

	err := testSubject.Reconcile(context.Background(), realmName, true)
	assert.NoError(t, err, "Reconcile should succeed")

	assert.NotNil(t, updatedRealm, "Updated realm should be captured")
	assert.NotNil(t, updatedRealm.TenantId, "TenantId should be stamped for the tenant-id realm attribute")
	assert.Equal(t, testTenantId, *updatedRealm.TenantId, "TenantId should come from the realm DB")
	assert.NotNil(t, updatedRealm.SMTPServer, "SMTPServer should be configured")

	smtpSettings := *updatedRealm.SMTPServer
	assert.Equal(t, testSmtpHost, smtpSettings["host"], "SMTP host should be configured")
	assert.Equal(t, testSmtpFromEmail, smtpSettings["from"], "SMTP from email should be configured")

	iamMock.MockRealmService.AssertNumberOfCalls(t, "GetRealm", 1)
	iamMock.MockRealmService.AssertNumberOfCalls(t, "UpdateRealm", 1)
}

func TestRealmReconciler_CreateNewRealm_IncludesSmtpConfig(t *testing.T) {
	iamMock, testSubject := initRealmReconcilerTest()
	realmName := realmMock.REALM_NAME

	iamMock.MockRealmService.On("GetRealm", mock.Anything, realmName).Return(nil, iamError.ErrNotFound)

	var createdRealm *models.Realm
	iamMock.MockRealmService.On("CreateRealm", mock.Anything, mock.MatchedBy(func(realm *models.Realm) bool {
		createdRealm = realm
		return true
	})).Return(realmName, nil)
	iamMock.MockRealmService.On("SetRealmEmailScopeOptional", mock.Anything, realmName).Return(nil)

	err := testSubject.Reconcile(context.Background(), realmName, true)
	assert.NoError(t, err, "Reconcile should succeed")

	assert.NotNil(t, createdRealm, "Created realm should be captured")
	assert.NotNil(t, createdRealm.TenantId, "TenantId should be stamped for the tenant-id realm attribute")
	assert.Equal(t, testTenantId, *createdRealm.TenantId, "TenantId should come from the realm DB")
	assert.NotNil(t, createdRealm.SMTPServer, "SMTPServer should be configured on new realm")

	smtpSettings := *createdRealm.SMTPServer
	assert.Equal(t, testSmtpHost, smtpSettings["host"], "SMTP host should be configured")
	assert.Equal(t, "1234", smtpSettings["port"], "SMTP port should be configured")
	assert.Equal(t, testSmtpFromEmail, smtpSettings["from"], "SMTP from email should be configured")

	iamMock.MockRealmService.AssertNumberOfCalls(t, "GetRealm", 1)
	iamMock.MockRealmService.AssertNumberOfCalls(t, "CreateRealm", 1)
}

func TestRealmReconciler_SmtpConfigValues(t *testing.T) {
	_, testSubject := initRealmReconcilerTest()

	assert.Equal(t, testSmtpHost, testSubject.config.smtpHost)
	assert.Equal(t, testSmtpPort, testSubject.config.smtpPort)
	assert.Equal(t, testSmtpFromEmail, testSubject.config.smtpFromEmail)
}
