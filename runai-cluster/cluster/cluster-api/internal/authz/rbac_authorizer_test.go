/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package authz

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cacheTesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/testing"

	authv1 "k8s.io/api/authorization/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

var _ = Describe("RBAC Authorizer", func() {
	var ctrl *gomock.Controller

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("Cached Responses", func() {
		var (
			k8s            client.Client
			rbacAuthorizer *RBACAuthorizer
			ctx            context.Context
			cacheMock      *cacheTesting.MockCacher
		)

		BeforeEach(func() {
			k8s = fake.NewClientBuilder().WithInterceptorFuncs(interceptor.Funcs{
				Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
					switch v := obj.(type) {
					default:
						panic("type must be SubjectAccessReview")
					case *authv1.SubjectAccessReview:
						if v.Spec.User == "authorized" {
							v.Status.Allowed = true
							v.Status.Denied = false
						} else {
							v.Status.Allowed = false
							v.Status.Denied = true
						}

						return nil
					}
				},
			}).Build()

			cacheMock = cacheTesting.SetMock(GinkgoT())
			rbacAuthorizer = NewRBACAuthorizer(k8s)
			ctx = context.Background()
		})

		It("should store a response in the cache", func() {
			req := authorizedAccessRequest()

			cacheKey := cacheKeyForRequest(req)

			cacheMock.On("Get", ctx, cacheTesting.KeyFromType[RBACAuthorizeResponse](), cacheKey).Return(nil, fmt.Errorf("failed"))
			cacheMock.On("Set", ctx, cacheTesting.KeyFromType[RBACAuthorizeResponse](), cacheKey, RBACAuthorizeResponse{AccessAllowed: true}).Return(nil)

			err := rbacAuthorizer.Authorize(ctx, req)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should used a cached response from the cache", func() {
			req := authorizedAccessRequest()

			cacheKey := cacheKeyForRequest(req)

			cacheMock.On("Get", ctx, cacheTesting.KeyFromType[RBACAuthorizeResponse](), cacheKey).Return(&RBACAuthorizeResponse{
				AccessAllowed: true,
			}, nil)

			err := rbacAuthorizer.Authorize(ctx, req)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should refresh an expired cache response", func() {
			req := authorizedAccessRequest()

			cacheKey := cacheKeyForRequest(req)

			cacheMock.On("Get", ctx, cacheTesting.KeyFromType[RBACAuthorizeResponse](), cacheKey).Return(nil, nil)
			cacheMock.On("Set", ctx, cacheTesting.KeyFromType[RBACAuthorizeResponse](), cacheKey, RBACAuthorizeResponse{AccessAllowed: true}).Return(nil)

			err := rbacAuthorizer.Authorize(ctx, req)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

func authorizedAccessRequest() *AccessRequest {
	return &AccessRequest{
		Resource: &Resource{
			Namespace: "testnamespace",
			Name:      "testnamespacename",
			Group:     "testgroup/v1",
		},
		Subject: "authorized",
		Groups:  []string{"A", "B"},
		Action:  "get",
	}
}
