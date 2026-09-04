/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/utils"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/utils/ptr"
)

var _ = Describe("Deployment", func() {

	var runaiConfig *runaiv1.RunaiConfig

	BeforeEach(func() {
		runaiConfig = utils.GetDefaultRunaiConfig()
	})

	Context("Change Default Image", func() {

		deployment := &v1.Deployment{}

		It("update image and tag", func() {
			deploymentForRunaiConfig(runaiConfig, deployment)
			tag := runaiConfig.Spec.PodGrouper.Image.Tag
			name := runaiConfig.Spec.PodGrouper.Image.Name
			repository := runaiConfig.Spec.PodGrouper.Image.Registry
			Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(BeEquivalentTo(fmt.Sprintf("%s/%s:%s",
				*repository, *name, *tag)))

			runaiConfig.Spec.PodGrouper.Image.Tag = ptr.To[string]("tag")
			runaiConfig.Spec.PodGrouper.Image.Name = ptr.To[string]("name")

			deploymentForRunaiConfig(runaiConfig, deployment)
			Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(BeEquivalentTo("gcr.io/run-ai-env/name:tag"))
		})
	})
})
