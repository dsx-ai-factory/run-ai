/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package status

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cluster/apis"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/cluster-api-client/generated"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

const (
	TestFilePath  = "internal/tests/config_samples/"
	EmptyJsonFile = "empty_json.json"
)

var _ = Describe("Status", func() {
	const (
		RunningStatus = "Running"
		PendingStatus = "Pending"
	)

	var (
		ctrl         *gomock.Controller
		statusClient *mocks.MockClusterStatusClient
		service      *ClusterService
		workload     *apis.Workload
		ctx          context.Context
	)

	BeforeEach(func() {
		homePath := getHomePath()
		err := config.GetInstance().Initializing(config.Options{Debug: true, ConfigPath: homePath, ConfigFile: EmptyJsonFile})
		if err != nil {
			Fail("failed to create config instance")
		}
		ctx = context.Background()
		ctrl = gomock.NewController(GinkgoT())
		statusClient = mocks.NewMockClusterStatusClient(ctrl)
		service = NewClusterService(statusClient)

		workload = &apis.Workload{
			Type:    api.Workspace,
			Project: "proj",
			Name:    "w",
		}
	})

	Describe("ClusterService", func() {
		Context("get cluster status", func() {
			It("response body is nil - should create default Metadata (all false)", func() {
				statusClient.EXPECT().GetClusterStatus(gomock.Any()).Return(nil, nil)
				metadata, err := service.GetCluster(ctx)
				Expect(err).ToNot(HaveOccurred())
				Expect(metadata).ToNot(BeNil())
				Expect(metadata.Features.MessageEnabled).To(BeFalse())
			})

			It("response body is 200 OK - older version cluster response test - should create default Metadata (all false)", func() {
				statusClient.EXPECT().GetClusterStatus(gomock.Any()).Return([]byte("\"OK\""), nil)
				metadata, err := service.GetCluster(ctx)
				Expect(err).ToNot(HaveOccurred())
				Expect(metadata).ToNot(BeNil())
				Expect(metadata.Features.MessageEnabled).To(BeFalse())
			})

			It("response body is empty json - if json is valid but has not attribute - should create default Metadata (all false)", func() {
				statusClient.EXPECT().GetClusterStatus(gomock.Any()).Return([]byte("{}"), nil)
				metadata, err := service.GetCluster(ctx)
				Expect(err).ToNot(HaveOccurred())
				Expect(metadata).ToNot(BeNil())
				Expect(metadata.Features.MessageEnabled).To(BeFalse())
			})

			It("response body is metadata structure", func() {
				metaByte, _ := json.Marshal(Metadata{
					Version:  "1.1",
					Features: Features{MessageEnabled: true, TerminalResizeEnabled: true},
				})
				statusClient.EXPECT().GetClusterStatus(gomock.Any()).Return(metaByte, nil)
				metadata, err := service.GetCluster(ctx)
				Expect(err).ToNot(HaveOccurred())
				Expect(metadata).ToNot(BeNil())
				Expect(metadata.Version).To(Equal("1.1"))
				Expect(metadata.Features.MessageEnabled).To(BeTrue())
				Expect(metadata.Features.TerminalResizeEnabled).To(BeTrue())
			})

			It("return any error should fail the health check", func() {
				statusClient.EXPECT().GetClusterStatus(gomock.Any()).Return(nil, fmt.Errorf("bla"))
				metadata, err := service.GetCluster(ctx)
				Expect(err).To(HaveOccurred())
				Expect(metadata).To(BeNil())
			})
		})

		Context("validate workload status", func() {
			It("nil workload", func() {
				err := service.ValidateWorkloadStatus(ctx, nil, RunningStatus)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(workloadIsNilMsg))
			})

			It("status match", func() {
				statusClient.EXPECT().GetWorkloadStatus(gomock.Any(), gomock.Any()).Return(&openapi.WorkloadStatus{Status: ptr.To(RunningStatus)}, nil)
				err := service.ValidateWorkloadStatus(ctx, workload, RunningStatus)
				Expect(err).NotTo(HaveOccurred())
			})

			It("status mismatch", func() {
				statusClient.EXPECT().GetWorkloadStatus(gomock.Any(), gomock.Any()).Return(&openapi.WorkloadStatus{Status: ptr.To(RunningStatus)}, nil)
				err := service.ValidateWorkloadStatus(ctx, workload, PendingStatus)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(fmt.Sprintf(statusMismatchTemplate, workload.Name, PendingStatus)))
			})

			It("client error", func() {
				expectedErr := errors.New("fail fail fail")
				statusClient.EXPECT().GetWorkloadStatus(gomock.Any(), gomock.Any()).Return(nil, expectedErr)
				err := service.ValidateWorkloadStatus(ctx, workload, RunningStatus)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(expectedErr))
			})
		})
	})
})

func getHomePath() string {
	homePath, err := os.Getwd()
	if err != nil {
		Fail("failed to get pwd")
	}
	homePath = strings.SplitAfter(homePath, "runai-cli/")[0]
	homePath += TestFilePath
	return homePath
}
