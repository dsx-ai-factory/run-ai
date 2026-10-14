/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nim_workloads

import (
	"context"
	"encoding/json"
	"net/http"

	nimv1alpha1 "github.com/NVIDIA/k8s-nim-operator/api/apps/v1alpha1"
	clustersApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	api "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NIM compute mode", func() {
	DescribeTable("validates the cluster version before composing the mode", func(version, errorMessage string) {
		cluster := &clustersApi.DisplayedCluster{}
		Expect(json.Unmarshal([]byte(`{"status":{"dependencies":{"required":{"gpu-stack":{"available":true,"components":{"nvidia":{"available":true}}}},"optional":{}}}}`), cluster)).To(Succeed())
		cluster.SetVersion(version)
		c := NewNimServiceComposer("model-service", "user", cluster, "p1", composer.OptionsForNimService())
		var src api.NimServiceSpecAllOfCompute
		Expect(json.Unmarshal([]byte(`{"gpuMemoryPortionRequest":0.25,"gpuComputeMode":"sm-sharing"}`), &src)).To(Succeed())
		var target nimv1alpha1.NIMServiceSpec
		var sizing composer.GpuSizingInfo
		err := c.processComputeFields(context.Background(), &src, &target, &sizing)()
		if errorMessage != "" {
			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(Equal(int32(http.StatusBadRequest)))
			Expect(err.Message).To(Equal(errorMessage))
			return
		}
		Expect(err).To(BeNil())
		Expect(target.Annotations).To(HaveKeyWithValue("nvidia.com/container.model-service-ctr.gpu-compute.mode", "sm-sharing"))
		Expect(target.Annotations).To(HaveKeyWithValue(composer.GpuFractionAnnotation, "0.25"))
	},
		Entry("uses the NIM operator container name on a supported cluster", "3.0.14", ""),
		Entry("reports an unknown cluster version clearly", "", "compute.gpuComputeMode: cluster version is unknown"),
	)
})
