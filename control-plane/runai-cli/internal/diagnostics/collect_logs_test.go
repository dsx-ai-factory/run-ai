/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package diagnostics_test

import (
	"context"
	"fmt"

	mocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/diagnostics"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/diagnostics"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

var _ = Describe("CollectLogsCommand", func() {
	var mockCtrl *gomock.Controller
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
		mockCtrl = gomock.NewController(GinkgoT())
		config.GetInstance().SetLogger(zap.NewNop())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Context("Run", func() {
		filterRes := &diagnostics.FilterResult{
			Valid: []string{
				"namespace-a", "namespace-b", "namespace-c",
			},
		}

		It("happy flow", func() {
			clusterValidator := mocks.NewMockClusterValidatorer(mockCtrl)
			logsCollector := mocks.NewMockLogsCollectorer(mockCtrl)
			namespaceFilter := mocks.NewMockNamespaceFilterer(mockCtrl)

			clusterValidator.EXPECT().ValidateClusterIdentity(ctx).Return(true, nil)
			clusterValidator.EXPECT().ValidateConnectivity(ctx).Return(nil).Times(1)
			namespaceFilter.EXPECT().Filter(ctx).Return(filterRes, nil).Times(1)
			logsCollector.EXPECT().Collect(ctx, filterRes.Valid).Times(1)

			clc := diagnostics.NewCollectLogsCommand(clusterValidator, logsCollector, namespaceFilter, true)

			Expect(clc.Run(ctx)).To(BeNil())
		})

		It("should return no error, no valid namespaces after filtering, namespaces flag not changed", func() {
			clusterValidator := mocks.NewMockClusterValidatorer(mockCtrl)
			logsCollector := mocks.NewMockLogsCollectorer(mockCtrl)
			namespaceFilter := mocks.NewMockNamespaceFilterer(mockCtrl)

			clusterValidator.EXPECT().ValidateClusterIdentity(ctx).Return(true, nil)
			clusterValidator.EXPECT().ValidateConnectivity(ctx).Return(nil).Times(1)
			namespaceFilter.EXPECT().Filter(ctx).Return(&diagnostics.FilterResult{
				NonExisting: []string{
					"namespace-a", "namespace-b", "namespace-c",
				},
			}, nil).Times(1)

			clc := diagnostics.NewCollectLogsCommand(clusterValidator, logsCollector, namespaceFilter, false)
			Expect(clc.Run(ctx).Error()).To(Equal("no valid namespaces found to collect logs from. Verify your permissions and try again"))
		})

		It("should return error, no valid namespaces after filtering, namespaces flag changed", func() {
			clusterValidator := mocks.NewMockClusterValidatorer(mockCtrl)
			logsCollector := mocks.NewMockLogsCollectorer(mockCtrl)
			namespaceFilter := mocks.NewMockNamespaceFilterer(mockCtrl)

			clusterValidator.EXPECT().ValidateClusterIdentity(ctx).Return(true, nil)
			clusterValidator.EXPECT().ValidateConnectivity(ctx).Return(nil).Times(1)
			namespaceFilter.EXPECT().Filter(ctx).Return(&diagnostics.FilterResult{
				NonExisting: []string{
					"namespace-a", "namespace-b", "namespace-c",
				},
			}, nil).Times(1)

			clc := diagnostics.NewCollectLogsCommand(clusterValidator, logsCollector, namespaceFilter, true)
			Expect(clc.Run(ctx).Error()).To(Equal("no valid namespaces found to collect logs from. Verify your permissions and try again"))
		})

		It("should not return err if namespace not permitted or not exist, namespaces flag not changed", func() {
			clusterValidator := mocks.NewMockClusterValidatorer(mockCtrl)
			logsCollector := mocks.NewMockLogsCollectorer(mockCtrl)
			namespaceFilter := mocks.NewMockNamespaceFilterer(mockCtrl)

			clusterValidator.EXPECT().ValidateClusterIdentity(ctx).Return(true, nil)
			clusterValidator.EXPECT().ValidateConnectivity(ctx).Return(nil).Times(1)
			namespaceFilter.EXPECT().Filter(ctx).Return(&diagnostics.FilterResult{
				Valid:        filterRes.Valid,
				NonPermitted: []string{"other-namespaces"},
				NonExisting:  []string{"some-namespaces"},
			}, nil)
			logsCollector.EXPECT().Collect(ctx, filterRes.Valid).Times(1)

			clc := diagnostics.NewCollectLogsCommand(clusterValidator, logsCollector, namespaceFilter, false)

			Expect(clc.Run(ctx)).To(BeNil())
		})

		It("should return err if namespace not permitted, namespaces flag changed", func() {
			clusterValidator := mocks.NewMockClusterValidatorer(mockCtrl)
			namespaceFilter := mocks.NewMockNamespaceFilterer(mockCtrl)

			clusterValidator.EXPECT().ValidateClusterIdentity(ctx).Return(true, nil)
			clusterValidator.EXPECT().ValidateConnectivity(ctx).Return(nil)
			namespaceFilter.EXPECT().Filter(ctx).Return(&diagnostics.FilterResult{
				Valid:        []string{"valid-namespaces"},
				NonPermitted: []string{"other-namespaces"},
				NonExisting:  []string{"some-namespaces"},
			}, nil)

			clc := diagnostics.NewCollectLogsCommand(clusterValidator, nil, namespaceFilter, true)

			err := clc.Run(ctx)
			Expect(err.Error()).To(Equal(
				`namespaces do not exist in cluster [some-namespaces]
namespaces are not permitted in cluster [other-namespaces]`))
		})

		It("should return err if filter failed", func() {
			clusterValidator := mocks.NewMockClusterValidatorer(mockCtrl)
			namespaceFilter := mocks.NewMockNamespaceFilterer(mockCtrl)

			err := fmt.Errorf("failed to connect to k8s")
			clusterValidator.EXPECT().ValidateClusterIdentity(ctx).Return(true, nil)
			clusterValidator.EXPECT().ValidateConnectivity(ctx).Return(nil)
			namespaceFilter.EXPECT().Filter(ctx).Return(&diagnostics.FilterResult{}, err)

			clc := diagnostics.NewCollectLogsCommand(clusterValidator, nil, namespaceFilter, true)

			Expect(clc.Run(ctx)).To(Equal(err))
		})

		It("should return err if validate connectivity failed", func() {
			clusterValidator := mocks.NewMockClusterValidatorer(mockCtrl)

			err := fmt.Errorf("failed to connect to k8s")
			clusterValidator.EXPECT().ValidateClusterIdentity(ctx).Return(true, nil)
			clusterValidator.EXPECT().ValidateConnectivity(ctx).Return(err).Times(1)

			clc := diagnostics.NewCollectLogsCommand(clusterValidator, nil, nil, true)

			Expect(clc.Run(ctx).Error()).To(Equal(diagnostics.K8SCommunicationFailed))
		})

		It("should return err if validate cluster identity failed", func() {
			clusterValidator := mocks.NewMockClusterValidatorer(mockCtrl)

			err := fmt.Errorf("not same cluster in k8s configuration")
			clusterValidator.EXPECT().ValidateClusterIdentity(ctx).Return(true, err)

			clc := diagnostics.NewCollectLogsCommand(clusterValidator, nil, nil, true)

			Expect(clc.Run(ctx).Error()).To(Equal(diagnostics.ClusterValidationFailed))
		})
	})
})
