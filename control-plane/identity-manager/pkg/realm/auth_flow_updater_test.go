/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package realm_updater

import (
	"context"
	"testing"

	iamMock "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/mocks"
	iamModels "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func initTestSubject() (*AuthFlowUpdater, *iamMock.MockServicesContainer) {
	mockIamServicesContainer := iamMock.NewMockServicesContainer()
	mockIam := mockIamServicesContainer.ToServicesContainer()
	return NewAuthFlowUpdater(context.Background(), *mockIam), mockIamServicesContainer
}

var authExecutionsLegacy = []*iamModels.ModifyAuthenticationExecutionRepresentation{
	{
		ProviderID: ptr.To("idp-review-profile"),
	},
}

var authExecutions = []*iamModels.ModifyAuthenticationExecutionRepresentation{
	{
		DisplayName: ptr.To("Review profile"),
		FlowID:      ptr.To("browser"),
		ID:          ptr.To("testID "),
	},
}

var cookieAuthExecutions = []*iamModels.ModifyAuthenticationExecutionRepresentation{
	{
		DisplayName: ptr.To("Cookie"),
		FlowID:      ptr.To("browser"),
		ID:          ptr.To("testID "),
	},
}

var authFlows = []*iamModels.AuthenticationFlowRepresentation{
	{
		Alias: ptr.To("browser"),
	},
}

func TestUpdateAuthFlows_HappyFlowWithLegacy(t *testing.T) {
	testSubject, iamMock := initTestSubject()

	iamMock.MockAuthFlowService.On("GetAuthExecutions", mock.Anything, mock.Anything, mock.Anything).Return(authExecutionsLegacy, nil)
	iamMock.MockAuthFlowService.On("GetAuthFlow", mock.Anything, mock.Anything, mock.Anything).Return(authFlows[0], nil)

	err := testSubject.UpdateAuthFlows(context.Background(), "test realm")

	assert.Nil(t, err)
	iamMock.MockAuthFlowService.AssertNumberOfCalls(t, "GetAuthExecutions", 1)
	iamMock.MockAuthFlowService.AssertNumberOfCalls(t, "GetAuthFlow", 1)
}

func TestUpdateAuthFlows_HappyFlowWithoutLegacy(t *testing.T) {
	testSubject, iamMock := initTestSubject()

	iamMock.MockAuthFlowService.On("GetAuthExecutions", mock.Anything, mock.Anything, mock.Anything).Return(authExecutions, nil)
	iamMock.MockAuthFlowService.On("GetAuthFlow", mock.Anything, mock.Anything, mock.Anything).Return(authFlows[0], nil)

	err := testSubject.UpdateAuthFlows(context.Background(), "test realm")

	assert.Nil(t, err)
	iamMock.MockAuthFlowService.AssertNumberOfCalls(t, "GetAuthExecutions", 1)
	iamMock.MockAuthFlowService.AssertNumberOfCalls(t, "GetAuthFlow", 1)
}

func TestUpdateBrowserFlows_HappyFlow(t *testing.T) {
	testSubject, iamMock := initTestSubject()

	iamMock.MockAuthFlowService.On("GetAuthExecutions", mock.Anything, mock.Anything, mock.Anything).Return(cookieAuthExecutions, nil)
	iamMock.MockAuthFlowService.On("UpdateAuthExecution", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := testSubject.alternativeAuthStepForAuthFlow(context.Background(), "test realm", utils.BROWSER_FLOW_NAME, "Cookie")

	assert.Nil(t, err)
	iamMock.MockAuthFlowService.AssertNumberOfCalls(t, "GetAuthExecutions", 1)
	iamMock.MockAuthFlowService.AssertNumberOfCalls(t, "UpdateAuthExecution", 1)
}
