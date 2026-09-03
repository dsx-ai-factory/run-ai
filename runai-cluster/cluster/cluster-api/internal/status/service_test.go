/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package status

import (
	"context"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/generated/pkg/api"
	workloadDomain "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/workload"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	projectName       = "project"
	workloadType      = "training"
	workloadFramework = "runai"
	workloadName      = "jp"
	workloadPhase     = "Creating"
)

var _ = Describe("ApiCliService", func() {
	var (
		statusApiService api.ClusterAPIAPIServicer
		ctx              context.Context
		workload         *workloadDomain.Workload
	)

	BeforeEach(func() {
		statusApiService = NewStatusApiService()

		workload = &workloadDomain.Workload{
			Name:        workloadName,
			Namespace:   "",
			Project:     projectName,
			BackendUUID: "",
			ClusterUUID: "",
			JobType:     "runai-job",
			Type:        workloadType,
			Phase:       workloadPhase,
		}

		ctx = workloadDomain.SetIntoContext(context.Background(), workload)
	})

	Context("Get Workload Status", func() {
		It("success", func() {
			res, err := statusApiService.GetWorkloadStatus(ctx, projectName, workloadType, workloadName, workloadFramework)
			Expect(err).NotTo(HaveOccurred())

			Expect(res.Code).To(Equal(http.StatusOK))
			Expect(res.Body).To(Equal(api.WorkloadStatus{Status: workloadPhase}))
		})

		It("workload not it context", func() {
			ctx = context.Background()
			res, err := statusApiService.GetWorkloadStatus(ctx, projectName, workloadType, workloadName, workloadFramework)
			Expect(err).To(HaveOccurred())

			Expect(res.Code).To(Equal(http.StatusInternalServerError))
			Expect(res.Body).To(BeNil())
		})
	})
})
