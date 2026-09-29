/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client_reconciler

import (
	"context"
	"errors"
	"testing"

	iamMock "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/mocks"
	iamModels "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	client_mocks "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/realm/client/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

func TestReconcileClient_ReturnsErrorWhenGetClientDataIsNil(t *testing.T) {
	const (
		realm    = "test-realm"
		clientID = "runai"
	)

	mockContainer := iamMock.NewMockServicesContainer()
	mockContainer.MockClientService.On(
		"GetClientByClientId", context.Background(), realm, clientID,
	).Return(nil, nil)

	clientMock := client_mocks.NewClient(t)
	clientMock.On("GetClientID").Return(clientID)
	clientMock.On("GetClientData", mock.Anything, mock.Anything).Return(nil)

	helper := NewClientReconcilerHelper(context.Background(), *mockContainer.ToServicesContainer(), nil)

	err := helper.ReconcileClient(context.Background(), realm, clientMock)

	assert.ErrorContains(t, err, "failed to prepare client data for client runai in realm test-realm")
}

func TestReconcileClient_GetClientByClientIdError(t *testing.T) {
	const (
		realm    = "test-realm"
		clientID = "runai"
	)

	mockContainer := iamMock.NewMockServicesContainer()
	mockContainer.MockClientService.On(
		"GetClientByClientId", context.Background(), realm, clientID,
	).Return(nil, errors.New("iam unavailable"))

	clientMock := client_mocks.NewClient(t)
	clientMock.On("GetClientID").Return(clientID)

	helper := NewClientReconcilerHelper(context.Background(), *mockContainer.ToServicesContainer(), nil)

	err := helper.ReconcileClient(context.Background(), realm, clientMock)

	assert.ErrorContains(t, err, "iam unavailable")
}

func TestReconcileClient_UpdatesExistingClient(t *testing.T) {
	const (
		realm    = "test-realm"
		clientID = "runai"
	)

	existingClient := &iamModels.Client{ClientID: ptr.To(clientID)}
	updatedClient := &iamModels.Client{ID: ptr.To("some-id"), ClientID: ptr.To(clientID)}

	mockContainer := iamMock.NewMockServicesContainer()
	mockContainer.MockClientService.On(
		"GetClientByClientId", context.Background(), realm, clientID,
	).Return(existingClient, nil)
	mockContainer.MockClientService.On(
		"UpdateClient", context.Background(), realm, updatedClient,
	).Return(nil)

	clientMock := client_mocks.NewClient(t)
	clientMock.On("GetClientID").Return(clientID)
	clientMock.On("GetClientData", mock.Anything, existingClient).Return(updatedClient)
	clientMock.On("NeedsEmailScopeDefault").Return(false)

	helper := NewClientReconcilerHelper(context.Background(), *mockContainer.ToServicesContainer(), nil)

	err := helper.ReconcileClient(context.Background(), realm, clientMock)

	assert.NoError(t, err)
	mockContainer.MockClientService.AssertCalled(t, "UpdateClient", context.Background(), realm, updatedClient)
}

func TestReconcileClient_UpdateClientError(t *testing.T) {
	const (
		realm    = "test-realm"
		clientID = "runai"
	)

	existingClient := &iamModels.Client{ClientID: ptr.To(clientID)}
	updatedClient := &iamModels.Client{ID: ptr.To("some-id"), ClientID: ptr.To(clientID)}

	mockContainer := iamMock.NewMockServicesContainer()
	mockContainer.MockClientService.On(
		"GetClientByClientId", context.Background(), realm, clientID,
	).Return(existingClient, nil)
	mockContainer.MockClientService.On(
		"UpdateClient", context.Background(), realm, updatedClient,
	).Return(errors.New("update failed"))

	clientMock := client_mocks.NewClient(t)
	clientMock.On("GetClientID").Return(clientID)
	clientMock.On("GetClientData", mock.Anything, existingClient).Return(updatedClient)

	helper := NewClientReconcilerHelper(context.Background(), *mockContainer.ToServicesContainer(), nil)

	err := helper.ReconcileClient(context.Background(), realm, clientMock)

	assert.ErrorContains(t, err, "update failed")
}

func TestReconcileClient_CreatesNewClient(t *testing.T) {
	const (
		realm    = "test-realm"
		clientID = "runai"
	)

	newClient := &iamModels.Client{ClientID: ptr.To(clientID)}

	mockContainer := iamMock.NewMockServicesContainer()
	mockContainer.MockClientService.On(
		"GetClientByClientId", context.Background(), realm, clientID,
	).Return(nil, nil)
	mockContainer.MockClientService.On(
		"CreateClient", context.Background(), realm, newClient,
	).Return("new-id", nil)

	clientMock := client_mocks.NewClient(t)
	clientMock.On("GetClientID").Return(clientID)
	clientMock.On("GetClientData", mock.Anything, (*iamModels.Client)(nil)).Return(newClient)
	clientMock.On("NeedsEmailScopeDefault").Return(false)

	helper := NewClientReconcilerHelper(context.Background(), *mockContainer.ToServicesContainer(), nil)

	err := helper.ReconcileClient(context.Background(), realm, clientMock)

	assert.NoError(t, err)
	mockContainer.MockClientService.AssertCalled(t, "CreateClient", context.Background(), realm, newClient)
}

func TestReconcileClient_SetsEmailScopeDefaultWhenProfileRequiresIt(t *testing.T) {
	const (
		realm      = "test-realm"
		clientID   = "runai-admin-ui"
		internalID = "some-internal-id"
	)

	existingClient := &iamModels.Client{ID: ptr.To(internalID), ClientID: ptr.To(clientID)}
	updatedClient := &iamModels.Client{ID: ptr.To(internalID), ClientID: ptr.To(clientID)}

	mockContainer := iamMock.NewMockServicesContainer()
	mockContainer.MockClientService.On(
		"GetClientByClientId", context.Background(), realm, clientID,
	).Return(existingClient, nil)
	mockContainer.MockClientService.On(
		"UpdateClient", context.Background(), realm, updatedClient,
	).Return(nil)
	mockContainer.MockClientService.On(
		"SetEmailScopeDefault", context.Background(), realm, internalID,
	).Return(nil)

	clientMock := client_mocks.NewClient(t)
	clientMock.On("GetClientID").Return(clientID)
	clientMock.On("GetClientData", mock.Anything, existingClient).Return(updatedClient)
	clientMock.On("NeedsEmailScopeDefault").Return(true)

	helper := NewClientReconcilerHelper(context.Background(), *mockContainer.ToServicesContainer(), nil)

	err := helper.ReconcileClient(context.Background(), realm, clientMock)

	assert.NoError(t, err)
	mockContainer.MockClientService.AssertCalled(t, "SetEmailScopeDefault", context.Background(), realm, internalID)
}

func TestReconcileClient_DoesNotSetEmailScopeDefaultWhenProfileDoesNotRequireIt(t *testing.T) {
	const (
		realm      = "test-realm"
		clientID   = "runai-token-exchange"
		internalID = "some-internal-id"
	)

	existingClient := &iamModels.Client{ID: ptr.To(internalID), ClientID: ptr.To(clientID)}
	updatedClient := &iamModels.Client{ID: ptr.To(internalID), ClientID: ptr.To(clientID)}

	mockContainer := iamMock.NewMockServicesContainer()
	mockContainer.MockClientService.On(
		"GetClientByClientId", context.Background(), realm, clientID,
	).Return(existingClient, nil)
	mockContainer.MockClientService.On(
		"UpdateClient", context.Background(), realm, updatedClient,
	).Return(nil)

	clientMock := client_mocks.NewClient(t)
	clientMock.On("GetClientID").Return(clientID)
	clientMock.On("GetClientData", mock.Anything, existingClient).Return(updatedClient)
	clientMock.On("NeedsEmailScopeDefault").Return(false)

	helper := NewClientReconcilerHelper(context.Background(), *mockContainer.ToServicesContainer(), nil)

	err := helper.ReconcileClient(context.Background(), realm, clientMock)

	assert.NoError(t, err)
	mockContainer.MockClientService.AssertNotCalled(t, "SetEmailScopeDefault")
}

func TestDeleteClientIfExists_DeletesExistingClient(t *testing.T) {
	const (
		realm      = "test-realm"
		clientID   = "runai"
		internalID = "some-internal-id"
	)

	existingClient := &iamModels.Client{ID: ptr.To(internalID), ClientID: ptr.To(clientID)}

	mockContainer := iamMock.NewMockServicesContainer()
	mockContainer.MockClientService.On(
		"GetClientByClientId", context.Background(), realm, clientID,
	).Return(existingClient, nil)
	mockContainer.MockClientService.On(
		"DeleteClient", context.Background(), realm, internalID,
	).Return(nil)

	helper := NewClientReconcilerHelper(context.Background(), *mockContainer.ToServicesContainer(), nil)

	err := helper.DeleteClientIfExists(context.Background(), realm, clientID)

	assert.NoError(t, err)
	mockContainer.MockClientService.AssertCalled(t, "DeleteClient", context.Background(), realm, internalID)
}

func TestDeleteClientIfExists_NoopWhenClientDoesNotExist(t *testing.T) {
	const (
		realm    = "test-realm"
		clientID = "runai"
	)

	mockContainer := iamMock.NewMockServicesContainer()
	mockContainer.MockClientService.On(
		"GetClientByClientId", context.Background(), realm, clientID,
	).Return(nil, nil)

	helper := NewClientReconcilerHelper(context.Background(), *mockContainer.ToServicesContainer(), nil)

	err := helper.DeleteClientIfExists(context.Background(), realm, clientID)

	assert.NoError(t, err)
	mockContainer.MockClientService.AssertNotCalled(t, "DeleteClient", mock.Anything, mock.Anything, mock.Anything)
}

func TestDeleteClientIfExists_GetClientByClientIdError(t *testing.T) {
	const (
		realm    = "test-realm"
		clientID = "runai"
	)

	mockContainer := iamMock.NewMockServicesContainer()
	mockContainer.MockClientService.On(
		"GetClientByClientId", context.Background(), realm, clientID,
	).Return(nil, errors.New("iam unavailable"))

	helper := NewClientReconcilerHelper(context.Background(), *mockContainer.ToServicesContainer(), nil)

	err := helper.DeleteClientIfExists(context.Background(), realm, clientID)

	assert.ErrorContains(t, err, "iam unavailable")
}

func TestDeleteClientIfExists_DeleteClientError(t *testing.T) {
	const (
		realm      = "test-realm"
		clientID   = "runai"
		internalID = "some-internal-id"
	)

	existingClient := &iamModels.Client{ID: ptr.To(internalID), ClientID: ptr.To(clientID)}

	mockContainer := iamMock.NewMockServicesContainer()
	mockContainer.MockClientService.On(
		"GetClientByClientId", context.Background(), realm, clientID,
	).Return(existingClient, nil)
	mockContainer.MockClientService.On(
		"DeleteClient", context.Background(), realm, internalID,
	).Return(errors.New("delete failed"))

	helper := NewClientReconcilerHelper(context.Background(), *mockContainer.ToServicesContainer(), nil)

	err := helper.DeleteClientIfExists(context.Background(), realm, clientID)

	assert.ErrorContains(t, err, "delete failed")
}

func TestReconcileClient_CreateClientError(t *testing.T) {
	const (
		realm    = "test-realm"
		clientID = "runai"
	)

	newClient := &iamModels.Client{ClientID: ptr.To(clientID)}

	mockContainer := iamMock.NewMockServicesContainer()
	mockContainer.MockClientService.On(
		"GetClientByClientId", context.Background(), realm, clientID,
	).Return(nil, nil)
	mockContainer.MockClientService.On(
		"CreateClient", context.Background(), realm, newClient,
	).Return("", errors.New("create failed"))

	clientMock := client_mocks.NewClient(t)
	clientMock.On("GetClientID").Return(clientID)
	clientMock.On("GetClientData", mock.Anything, (*iamModels.Client)(nil)).Return(newClient)

	helper := NewClientReconcilerHelper(context.Background(), *mockContainer.ToServicesContainer(), nil)

	err := helper.ReconcileClient(context.Background(), realm, clientMock)

	assert.ErrorContains(t, err, "create failed")
}
