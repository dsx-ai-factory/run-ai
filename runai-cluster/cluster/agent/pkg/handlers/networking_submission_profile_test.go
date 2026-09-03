/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	workloadsManagerApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	clusterconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = Describe("networking submission profile conversion", func() {
	configuration := func(profile workloadsManagerApi.NetworkingSubmissionProfile) *workloadsManagerApi.WorkloadV2UpdateConfiguration {
		return &workloadsManagerApi.WorkloadV2UpdateConfiguration{
			NetworkingEndpoints: mapToNetworkingConfig(map[string]interface{}{
				"services": []interface{}{map[string]interface{}{
					"ref": map[string]interface{}{
						"name": "demo-service",
						"port": float64(8080),
					},
				}},
			}),
			NetworkingSubmissionProfile: &profile,
		}
	}

	manifest := func() *unstructured.Unstructured {
		object := &unstructured.Unstructured{Object: map[string]interface{}{
			"apiVersion": "batch/v1",
			"kind":       "Job",
			"metadata": map[string]interface{}{
				"name": "demo",
			},
		}}
		object.SetAnnotations(map[string]string{
			clusterconstants.NetworkingSubmissionProfileAnnotation: "caller-selected",
			"caller.example/keep": "value",
		})
		return object
	}

	DescribeTable("writes each trusted profile to the WorkloadRunner",
		func(profile workloadsManagerApi.NetworkingSubmissionProfile) {
			attachedWorkload := manifest()
			runner, err := wrapWithWorkloadRunner(attachedWorkload, "workload-id", configuration(profile))

			Expect(err).NotTo(HaveOccurred())
			Expect(runner.GetAnnotations()).To(HaveKeyWithValue(
				clusterconstants.NetworkingSubmissionProfileAnnotation,
				string(profile),
			))
			Expect(attachedWorkload.GetAnnotations()).NotTo(HaveKey(clusterconstants.NetworkingSubmissionProfileAnnotation))
			Expect(attachedWorkload.GetAnnotations()).To(HaveKeyWithValue("caller.example/keep", "value"))
		},
		Entry("direct v2", workloadsManagerApi.NETWORKINGSUBMISSIONPROFILE_DIRECT_V2),
		Entry("generated native", workloadsManagerApi.NETWORKINGSUBMISSIONPROFILE_GENERATED_NATIVE),
		Entry("generated NIM", workloadsManagerApi.NETWORKINGSUBMISSIONPROFILE_GENERATED_NIM),
	)

	It("rejects an unknown profile", func() {
		_, err := wrapWithWorkloadRunner(manifest(), "workload-id", configuration(
			workloadsManagerApi.NetworkingSubmissionProfile("caller-selected"),
		))

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unknown networking submission profile"))
	})

	It("rejects a profile without endpoint-array networking", func() {
		profile := workloadsManagerApi.NETWORKINGSUBMISSIONPROFILE_DIRECT_V2
		config := &workloadsManagerApi.WorkloadV2UpdateConfiguration{
			NetworkingSubmissionProfile: &profile,
		}

		_, err := wrapWithWorkloadRunner(manifest(), "workload-id", config)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("requires endpoint-array networking"))
	})

	It("keeps markerless updates compatible while stripping a spoofed workload annotation", func() {
		config := configuration(workloadsManagerApi.NETWORKINGSUBMISSIONPROFILE_DIRECT_V2)
		config.NetworkingSubmissionProfile = nil
		attachedWorkload := manifest()

		runner, err := wrapWithWorkloadRunner(attachedWorkload, "workload-id", config)

		Expect(err).NotTo(HaveOccurred())
		Expect(runner.GetAnnotations()).NotTo(HaveKey(clusterconstants.NetworkingSubmissionProfileAnnotation))
		Expect(attachedWorkload.GetAnnotations()).NotTo(HaveKey(clusterconstants.NetworkingSubmissionProfileAnnotation))
	})
})
