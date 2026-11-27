/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	orgUnitApi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	ourConfig "github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/config"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	assetsApi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type credentialAssetsClientFunc func(context.Context, bool) (*assetsApi.CredentialsListResponse, *assetsApi.Error)

func (f credentialAssetsClientFunc) ListCredentialsAssets(ctx context.Context, privateOnly bool) (*assetsApi.CredentialsListResponse, *assetsApi.Error) {
	return f(ctx, privateOnly)
}

var _ = Describe("User credential validation", func() {
	owned := assetsApi.CredentialsListResponseEntry{Meta: assetsApi.AssetMeta{
		Name: "test-credential", Id: "11111111-1111-4111-8111-111111111111", Kind: assetsApi.ASSETKIND_GENERIC_SECRET, Scope: assetsApi.SCOPE_TENANT,
	}}
	secondOwned := owned
	secondOwned.Meta.Name = "second-test-credential"
	secondOwned.Meta.Id = "33333333-3333-4333-8333-333333333333"
	registry := owned
	registry.Meta.Kind = assetsApi.ASSETKIND_DOCKER_REGISTRY
	ngc := owned
	ngc.Meta.Kind = assetsApi.ASSETKIND_NGC_API_KEY
	password := owned
	password.Meta.Name = "password-credential"
	password.Meta.Kind = assetsApi.ASSETKIND_PASSWORD
	accessKey := owned
	accessKey.Meta.Name = "access-key-credential"
	accessKey.Meta.Kind = assetsApi.ASSETKIND_ACCESS_KEY

	type testCase struct {
		refs      []string
		entries   []assetsApi.CredentialsListResponseEntry
		lookupErr *assetsApi.Error
		wantCode  int32
	}

	It("uses the configured assets service after startup", func(ctx SpecContext) {
		response, err := json.Marshal(assetsApi.CredentialsListResponse{Entries: []assetsApi.CredentialsListResponseEntry{owned}})
		Expect(err).NotTo(HaveOccurred())
		requests := make(chan *http.Request, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests <- r
			w.Header().Set("Content-Type", "application/json")
			_, writeErr := w.Write(response)
			if writeErr != nil {
				GinkgoWriter.Printf("write credentials response: %v\n", writeErr)
			}
		}))
		DeferCleanup(server.Close)
		originalConfig := *config.Get[*ourConfig.Services]()
		servicesConfig := originalConfig
		servicesConfig.AssetsServiceUrl = server.URL
		Expect(config.Set(&servicesConfig)).To(Succeed())
		DeferCleanup(func() { Expect(config.Set(&originalConfig)).To(Succeed()) })
		service := impl
		callerCtx := context.WithValue(ctx, runaictx.CfSubjectID, "test-user")
		callerCtx = context.WithValue(callerCtx, runaictx.CfTenantId, "123")

		validationErr := service.validateUserCredentials(callerCtx, []api.WorkloadV2UpdateConfigurationUserCredentialSource{{Ref: "test-credential-11111111-1"}})

		Expect(validationErr).To(BeNil())
		var request *http.Request
		Expect(requests).To(Receive(&request))
		Expect(request.Method).To(Equal(http.MethodGet))
		Expect(request.URL.Path).To(Equal("/api/v1/asset/credentials"))
		Expect(request.URL.Query().Get("privateOnly")).To(Equal("true"))
		Expect(request.Header.Get("X-Runai-Subject-Id")).To(Equal("test-user"))
		Expect(request.Header.Get("X-Runai-Tenant-Id")).To(Equal("123"))
	})

	DescribeTable("building a workload configuration", func(ctx SpecContext, tc testCase) {
		service := &WorkloadsV2Impl{credentialsClient: credentialAssetsClientFunc(func(requestCtx context.Context, privateOnly bool) (*assetsApi.CredentialsListResponse, *assetsApi.Error) {
			Expect(tc.refs).NotTo(BeEmpty())
			Expect(requestCtx).To(Equal(ctx))
			Expect(privateOnly).To(BeTrue())
			return &assetsApi.CredentialsListResponse{Entries: tc.entries}, tc.lookupErr
		})}
		params := api.WorkloadV2Configuration{}
		if len(tc.refs) > 0 {
			params.UserCredentials = &api.WorkloadV2UpdateConfigurationUserCredentials{}
			for _, ref := range tc.refs {
				params.UserCredentials.Sources = append(params.UserCredentials.Sources, api.WorkloadV2UpdateConfigurationUserCredentialSource{Ref: ref})
			}
		}

		configuration, err := service.buildConfiguration(ctx, params, &orgUnitApi.Project{}, &workloadsApi.WorkloadTypeConfig{}, &ResourceInfo{}, "test-workload", "")
		if tc.wantCode != 0 {
			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(Equal(tc.wantCode))
			Expect(configuration).To(BeNil())
			if tc.wantCode == http.StatusNotFound {
				Expect(err.Message).To(Equal("user credential not found"))
			}
			return
		}
		Expect(err).To(BeNil())
		Expect(configuration).NotTo(BeNil())
		if len(tc.refs) > 0 {
			Expect(configuration.UserCredentials).NotTo(BeNil())
			Expect(configuration.UserCredentials.Sources).To(Equal(tc.refs))
		} else {
			Expect(configuration.UserCredentials).To(BeNil())
		}
	},
		Entry("rejects another user's credential even if its name matches", testCase{
			refs: []string{"test-credential-22222222-2"}, entries: []assetsApi.CredentialsListResponseEntry{owned}, wantCode: http.StatusNotFound,
		}),
		Entry("accepts the user's own generic credential", testCase{
			refs: []string{"test-credential-11111111-1"}, entries: []assetsApi.CredentialsListResponseEntry{owned},
		}),
		Entry("accepts the user's own Docker registry credential", testCase{
			refs: []string{"test-credential-11111111-1"}, entries: []assetsApi.CredentialsListResponseEntry{registry},
		}),
		Entry("rejects the workload if any credential belongs to another user", testCase{
			refs: []string{"test-credential-11111111-1", "test-credential-22222222-2"}, entries: []assetsApi.CredentialsListResponseEntry{owned}, wantCode: http.StatusNotFound,
		}),
		Entry("accepts all requested credentials owned by the user", testCase{
			refs: []string{"test-credential-11111111-1", "second-test-credential-33333333-3"}, entries: []assetsApi.CredentialsListResponseEntry{owned, secondOwned},
		}),
		Entry("rejects NGC Secret names for a generic credential", testCase{
			refs: []string{"ngcgs-test-credential-11111111-1"}, entries: []assetsApi.CredentialsListResponseEntry{owned}, wantCode: http.StatusNotFound,
		}),
		Entry("rejects a credential that does not exist", testCase{
			refs: []string{"test-credential-22222222-2"}, wantCode: http.StatusNotFound,
		}),
		Entry("fails if the user's credentials cannot be loaded", testCase{
			refs: []string{"test-credential-11111111-1"}, lookupErr: assetsApi.NewError(http.StatusServiceUnavailable, "unavailable"), wantCode: http.StatusInternalServerError,
		}),
		Entry("accepts both secrets from the user's own NGC API key", testCase{
			refs: []string{"ngcgs-test-credential-11111111-1", "ngcdr-test-credential-11111111-1"}, entries: []assetsApi.CredentialsListResponseEntry{ngc},
		}),
		Entry("ignores unsupported credentials for valid references", testCase{
			refs: []string{"test-credential-11111111-1"}, entries: []assetsApi.CredentialsListResponseEntry{password, owned, accessKey},
		}),
		Entry("rejects password credential references", testCase{
			refs: []string{"password-credential-11111111-1"}, entries: []assetsApi.CredentialsListResponseEntry{password, owned}, wantCode: http.StatusNotFound,
		}),
		Entry("rejects access-key credential references", testCase{
			refs: []string{"access-key-credential-11111111-1"}, entries: []assetsApi.CredentialsListResponseEntry{accessKey, owned}, wantCode: http.StatusNotFound,
		}),
		Entry("skips credential lookup if no credentials are requested", testCase{}),
	)
})
