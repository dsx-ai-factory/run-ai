/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"go.uber.org/zap"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cluster/apis"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Waiting for workload readiness", func() {
	BeforeEach(func() {
		originalClient := clusterApiClient
		originalLogger := instance.GetLogger()
		clusterApiClient = nil
		instance.SetLogger(zap.NewNop())
		DeferCleanup(func() {
			clusterApiClient = originalClient
			instance.SetLogger(originalLogger)
		})
	})

	DescribeTable("validates the phase returned by the cluster API route",
		func(workloadType api.WorkloadType, framework api.Framework, phase, route string, ready bool) {
			requests := make(chan string, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests <- r.URL.Path
				w.Header().Set("Content-Type", "application/json")
				_, err := fmt.Fprintf(w, `{"status":%q}`, phase)
				if err != nil {
					GinkgoWriter.Printf("write workload status: %v\n", err)
				}
			}))
			DeferCleanup(server.Close)
			workload := apis.NewWorkloadForClusterAPI("project", workloadType, "workload", framework)

			err := WaitForWorkloadInClusterRunning(context.Background(), workload, server.URL+"/cluster-api", "", 0)

			Expect(requests).To(Receive(Equal("/cluster-api/api/v1/project/workloads/" + route + "/status")))
			if ready {
				Expect(err).NotTo(HaveOccurred())
			} else {
				Expect(err).To(MatchError("workload's waiting time was exceeded"))
			}
		},
		Entry("resolved distributed inference is Running", api.DistributedInference, api.UndefinedFramework, "Running", "external/external/workload", true),
		Entry("resolved inference is Running", api.Inference, api.UndefinedFramework, "Running", "external/external/workload", true),
		Entry("resolved distributed inference is pending", api.DistributedInference, api.UndefinedFramework, "Pending", "external/external/workload", false),
		Entry("resolved inference is pending", api.Inference, api.UndefinedFramework, "Pending", "external/external/workload", false),
		Entry("native distributed inference is Ready", api.DistributedInference, api.LWS, "Ready", "distributedinference/runai/workload", true),
		Entry("native inference is Ready", api.Inference, api.Knative, "Ready", "inference/knative/workload", true),
		Entry("native distributed inference is not Ready", api.DistributedInference, api.LWS, "Running", "distributedinference/runai/workload", false),
		Entry("native inference is not Ready", api.Inference, api.Knative, "Running", "inference/knative/workload", false),
		Entry("NIM is Running", api.Inference, api.NimService, "Running", "external/external/ew-nimservice-workload", true),
		Entry("training is Running", api.Training, api.Runai, "Running", "training/runai/workload", true),
	)
})
