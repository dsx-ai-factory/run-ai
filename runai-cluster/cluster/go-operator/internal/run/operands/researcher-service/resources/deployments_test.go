/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/utils"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	runv1_internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/utils/ptr"
)

var _ = Describe("Deployment", func() {

	var runaiConfig *runaiv1.RunaiConfig

	BeforeEach(func() {
		runaiConfig = utils.GetDefaultRunaiConfig()
		runaiConfig.Spec.Internal.Global.ControlPlane.Config.ClusterName = ptr.To[string]("cluster_name")
	})

	Context("Change Default Image", func() {

		deployment := &v1.Deployment{}

		It("update image and tag", func() {
			initDeployment(runaiConfig, deployment)
			tag := runaiConfig.Spec.ResearcherService.Image.Tag
			name := runaiConfig.Spec.ResearcherService.Image.Name
			repository := runaiConfig.Spec.ResearcherService.Image.Registry
			Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(BeEquivalentTo(fmt.Sprintf("%s/%s:%s",
				*repository, *name, *tag)))

			runaiConfig.Spec.ResearcherService.Image.Tag = ptr.To[string]("tag")
			runaiConfig.Spec.ResearcherService.Image.Name = ptr.To[string]("name")

			initDeployment(runaiConfig, deployment)
			Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(BeEquivalentTo("gcr.io/run-ai-prod/name:tag"))
		})
	})

	Context("Openshift flags", func() {
		It("should have openshift=false cli flag", func() {
			deployment := &v1.Deployment{}
			initDeployment(runaiConfig, deployment)
			Expect(deployment.Spec.Template.Spec.Containers[0].Args).To(ContainElement("--openshift=false"))
		})

		It("should have openshift=true cli flag", func() {
			runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor = runv1_internal.KubernetesFlavorOpenShift
			deployment := &v1.Deployment{}
			initDeployment(runaiConfig, deployment)
			Expect(deployment.Spec.Template.Spec.Containers[0].Args).To(ContainElement("--openshift=true"))
		})
	})
})
