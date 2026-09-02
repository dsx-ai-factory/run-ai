/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package control_plane

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	authzapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/mocks"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/api"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/auth"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/authz/authorizers"
)

var _ = Describe("Control Plane Authorizer", func() {
	Context("Authorization logic", func() {
		var (
			mockAuthzClient              *mocks.IAuthorizationExternalClient
			mockPermissionsAPI           *mocks.PermissionsAPI
			authzReq                     *authorizers.Request
			ctx                          context.Context
			apiEvaluatePermissionRequest authzapi.ApiEvaluatePermissionRequest
			cpAuthorizer                 *ControlPlaneAuthorizer
		)

		BeforeEach(func() {
			mockAuthzClient, mockPermissionsAPI = mockedObjects(GinkgoTB())
			ctx, authzReq = testAuthzRequest()
			apiEvaluatePermissionRequest = testAPIEvaluatePermissionsRequest(authzReq, mockPermissionsAPI)
			cpAuthorizer = NewControlPlaneAuthorizerWithClient(mockAuthzClient)
		})

		It("should pass authorization", func() {
			mockAuthzClient.On("EvaluatePermission", ctx).
				Return(apiEvaluatePermissionRequest)

			mockPermissionsAPI.On("EvaluatePermissionExecute", apiEvaluatePermissionRequest).
				Return(nil, &http.Response{
					StatusCode: 204,
				}, nil)

			allowed, reason, err := cpAuthorizer.Authorize(ctx, authzReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(allowed).To(BeTrue())
			Expect(reason).To(Equal(authorizers.ReasonAuthorized))
		})

		It("should fail authorization due to a lack of permissions", func() {
			mockPermissionsAPI.On("EvaluatePermissionExecute", apiEvaluatePermissionRequest).
				Return(nil, &http.Response{
					StatusCode: 403,
				}, nil)

			mockAuthzClient.On("EvaluatePermission", ctx).
				Return(apiEvaluatePermissionRequest)

			allowed, reason, err := cpAuthorizer.Authorize(ctx, authzReq)
			Expect(err).To(HaveOccurred())
			Expect(allowed).To(BeFalse())
			Expect(reason).To(Equal(authorizers.ReasonUnauthorized))
			Expect(err.Error()).To(ContainSubstring("forbidden"))
		})

		It("should fail authorization due to an unexpected error without response body", func() {
			mockPermissionsAPI.On("EvaluatePermissionExecute", apiEvaluatePermissionRequest).
				Return(nil, &http.Response{
					StatusCode: 500,
				}, nil)

			mockAuthzClient.On("EvaluatePermission", ctx).
				Return(apiEvaluatePermissionRequest)

			allowed, reason, err := cpAuthorizer.Authorize(ctx, authzReq)
			Expect(err).To(HaveOccurred())
			Expect(allowed).To(BeFalse())
			Expect(reason).To(Equal(authorizers.ReasonError))
			Expect(err.Error()).To(ContainSubstring("body"))
		})

		It("should fail authorization due to unexpected error", func() {
			mockPermissionsAPI.On("EvaluatePermissionExecute", apiEvaluatePermissionRequest).
				Return(nil, &http.Response{
					StatusCode: 500,
					Body:       io.NopCloser(bytes.NewReader([]byte(""))),
				}, nil)

			mockAuthzClient.On("EvaluatePermission", ctx).
				Return(apiEvaluatePermissionRequest)

			allowed, reason, err := cpAuthorizer.Authorize(ctx, authzReq)
			Expect(err).To(HaveOccurred())
			Expect(allowed).To(BeFalse())
			Expect(reason).To(Equal(authorizers.ReasonError))
			Expect(err.Error()).To(ContainSubstring("unexpected"))
		})
	})

	Context("WorkloadAction and WorkloadKind conversion", func() {
		It("should authorize read operations of all kinds against the generic authzapi.RESOURCETYPE_WORKLOADS type", func() {
			workloadAction := api.Logs
			workloadKind := api.Training
			authzAction, authzResourceType := workloadActionAndTypeToAuthorizationActionAndResourceType(workloadAction, workloadKind)
			Expect(authzAction).To(Equal(authzapi.ACTION_READ))
			Expect(authzResourceType).To(Equal(authzapi.RESOURCETYPE_WORKLOADS))
		})

		It("should authorize update operations against a workload's specific type", func() {
			Skip("unskip this once resolved")
			workloadAction := api.Exec
			workloadKind := api.Training
			authzAction, authzResourceType := workloadActionAndTypeToAuthorizationActionAndResourceType(workloadAction, workloadKind)
			Expect(authzAction).To(Equal(authzapi.ACTION_UPDATE))
			Expect(authzResourceType).To(Equal(authzapi.RESOURCETYPE_TRAININGS))
		})
	})
})

func mockedObjects(t testing.TB) (*mocks.IAuthorizationExternalClient, *mocks.PermissionsAPI) {
	mockAuthzClient := mocks.NewIAuthorizationExternalClient(t)
	mockPermissionsAPI := mocks.NewPermissionsAPI(t)

	return mockAuthzClient, mockPermissionsAPI
}

func testAuthzRequest() (context.Context, *authorizers.Request) {
	workloadName := "test-workload"
	workloadUID := "1234"

	authzRequest := &authorizers.Request{
		AuthInfo: &auth.Info{
			Username:          "test-user",
			Subject:           "test-subject",
			OriginalSubject:   "test-original-subject",
			Groups:            nil,
			PreferredUsername: "test-preferred-username",
			RawToken:          "test-token",
		},
		Workload:    workloadName,
		Action:      api.Exec,
		WorkloadUID: workloadUID,
		Kind:        api.Training,
	}

	ctx := context.WithValue(context.Background(), authzapi.ContextAccessToken, authzRequest.AuthInfo.RawToken)

	return ctx, authzRequest
}

func testAPIEvaluatePermissionsRequest(authzReq *authorizers.Request, mockPermissionsAPI *mocks.PermissionsAPI) authzapi.ApiEvaluatePermissionRequest {
	requestAction := workloadActionToAuthorizationAction(authzReq.Action)
	resourceType := authzapi.RESOURCETYPE_WORKLOADS // TODO use this once resolved: authzapi.RESOURCETYPE_TRAININGS
	apiEvaluatePermissionsRequest := authzapi.ApiEvaluatePermissionRequest{}.BackendEvaluationRequest(
		authzapi.BackendEvaluationRequest{
			ResourceId:    *authzapi.NewNullableString(&authzReq.WorkloadUID),
			RequestAction: *authzapi.NewNullableAction(&requestAction),
			ResourceType:  *authzapi.NewNullableResourceType(&resourceType),
		})
	apiEvaluatePermissionsRequest.ApiService = mockPermissionsAPI

	return apiEvaluatePermissionsRequest
}
