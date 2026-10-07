/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"

	schemeinit "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/scheme-init"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Deployment", func() {
	It("does not inject the cluster version", func() {
		runaiConfig := utils.GetDefaultRunaiConfig()
		k8sClient := fake.NewClientBuilder().WithScheme(schemeinit.Scheme()).Build()

		object, err := NewDeploymentResource("runtime-rules", k8sClient, runaiConfig).Object(context.Background())
		Expect(err).NotTo(HaveOccurred())

		deployment := object.(*appsv1.Deployment)
		Expect(deployment.Spec.Template.Spec.Containers).To(HaveLen(1))
		Expect(deployment.Spec.Template.Spec.Containers[0].Env).NotTo(ContainElement(HaveField("Name", "CLUSTER_VERSION")))
	})
})
