/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	smocks "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/mocks/pkg/services"
	errtypes "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/common/errors"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

var _ = Describe("ApiConfigService", func() {
	type Mocks struct {
		clusterConfigProvider *smocks.ClusterConfigProvider
	}

	var (
		mocks *Mocks

		ctx context.Context
	)

	const (
		clusterUUID = "cluster-uuid"
		tenantID    = int32(1001)
	)

	BeforeEach(func() {
		mocks = &Mocks{
			clusterConfigProvider: &smocks.ClusterConfigProvider{},
		}
		ctx = context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenantID))
	})

	Context("GetClusterControlPlaneConfig", func() {
		Context("Happy path", func() {
			It("Should return 200", func() {
				// Setup
				expectedClusterConfig := &api.ClusterControlPlaneConfig{}
				mocks.clusterConfigProvider.On("GetClusterControlPlaneConfig", ctx, int32(1001), clusterUUID).Return(expectedClusterConfig, nil)

				// Act
				response, err := services.NewControlPlaneConfigApiService(mocks.clusterConfigProvider).GetClusterControlPlaneConfig(ctx, clusterUUID)

				// Assert
				Expect(err).Should(BeNil())
				Expect(response.Code).Should(Equal(http.StatusOK))
				Expect(response.Body).Should(Equal(expectedClusterConfig))
			})
		})

		Context("Error path", func() {
			When("Failed to parse tenant ID", func() {
				It("Should return 400", func() {
					// Setup
					ctx := context.Background()

					// Act
					response, err := services.NewControlPlaneConfigApiService(mocks.clusterConfigProvider).GetClusterControlPlaneConfig(ctx, clusterUUID)

					// Assert
					Expect(err).Should(BeNil())
					Expect(response.Code).Should(Equal(http.StatusBadRequest))
				})
			})

			When("Failed to get cluster control plane config", func() {
				It("Should return 500", func() {
					// Setup
					mocks.clusterConfigProvider.On("GetClusterControlPlaneConfig", ctx, int32(1001), clusterUUID).Return(nil, fmt.Errorf("error"))

					// Act
					response, err := services.NewControlPlaneConfigApiService(mocks.clusterConfigProvider).GetClusterControlPlaneConfig(ctx, clusterUUID)

					// Assert
					Expect(err).Should(BeNil())
					Expect(response.Code).Should(Equal(http.StatusInternalServerError))
				})
			})

			When("Cluster control plane config not found", func() {
				It("Should return 404", func() {
					// Setup
					mocks.clusterConfigProvider.On("GetClusterControlPlaneConfig", ctx, int32(1001), clusterUUID).Return(nil, errtypes.ErrNotFound)

					// Act
					response, err := services.NewControlPlaneConfigApiService(mocks.clusterConfigProvider).GetClusterControlPlaneConfig(ctx, clusterUUID)

					// Assert
					Expect(err).Should(BeNil())
					Expect(response.Code).Should(Equal(http.StatusNotFound))
				})
			})
		})
	})
})
