/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	runutils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Project role-binding templates", func() {
	buildConfigMap := func(workloadEnabled, inferenceEnabled, jobEnabled bool) *corev1.ConfigMap {
		runaiConfig := runutils.GetDefaultRunaiConfig()
		runaiConfig.Namespace = runv1.DefaultNamespace
		runaiConfig.Spec.WorkloadController.Enabled = ptr.To(workloadEnabled)
		runaiConfig.Spec.InferenceWorkloadController.Enabled = ptr.To(inferenceEnabled)
		runaiConfig.Spec.JobController.Enabled = ptr.To(jobEnabled)
		client := fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).Build()

		object, err := ConfigMapForRunaiConfig(context.Background(), client, runaiConfig)
		Expect(err).NotTo(HaveOccurred())
		configMap, ok := object.(*corev1.ConfigMap)
		Expect(ok).To(BeTrue())
		return configMap
	}

	It("includes legacy controller templates while their controllers are enabled", func() {
		data := buildConfigMap(true, true, true).Data

		Expect(data).To(HaveKey("workload-controller-rw.yaml"))
		Expect(data).To(HaveKey("inference-workload-controller.yaml"))
		Expect(data).To(HaveKey("job-controller.yaml"))
		Expect(data).To(HaveKey("runai-job-controller-project.yaml"))
	})

	It("omits only templates owned by disabled legacy controllers", func() {
		data := buildConfigMap(false, false, false).Data

		Expect(data).NotTo(HaveKey("workload-controller-rw.yaml"))
		Expect(data).NotTo(HaveKey("inference-workload-controller.yaml"))
		Expect(data).NotTo(HaveKey("job-controller.yaml"))
		Expect(data).NotTo(HaveKey("runai-job-controller-project.yaml"))
		Expect(data).To(HaveKey("researcher-service-rw.yaml"))
	})
})
