/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package source

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Kubeflow training Karta replica-type pod selectors", func() {
	DescribeTable("matches the lowercase replica-type label while spec paths keep the CRD key",
		func(build func() models.Karta, replicaSpecsField string, expected map[string][2]string) {
			children := build().Spec.StructureDefinition.ChildComponents
			Expect(children).To(HaveLen(len(expected)))

			for _, child := range children {
				want, ok := expected[child.Name]
				Expect(ok).To(BeTrue(), "unexpected component %q", child.Name)
				specKey, labelValue := want[0], want[1]

				Expect(child.PodSelector).NotTo(BeNil())
				Expect(child.PodSelector.ComponentTypeSelector).NotTo(BeNil())
				Expect(child.PodSelector.ComponentTypeSelector.KeyPath).To(Equal(".metadata.labels[\"training.kubeflow.org/replica-type\"]"))
				Expect(child.PodSelector.ComponentTypeSelector.Value).NotTo(BeNil())
				Expect(*child.PodSelector.ComponentTypeSelector.Value).To(Equal(labelValue))

				Expect(child.SpecDefinition).NotTo(BeNil())
				Expect(child.SpecDefinition.PodTemplateSpecPath).NotTo(BeNil())
				Expect(*child.SpecDefinition.PodTemplateSpecPath).To(Equal(".spec." + replicaSpecsField + "." + specKey + ".template"))
			}
		},
		// component name -> {replica spec key, replica-type label value}
		Entry("TF v1", BuildKubeflowTFJobV1, "tfReplicaSpecs", map[string][2]string{
			"Chief":     {"Chief", "chief"},
			"Worker":    {"Worker", "worker"},
			"PS":        {"PS", "ps"},
			"Evaluator": {"Evaluator", "evaluator"},
		}),
		Entry("XGBoost v1", BuildKubeflowXGBoostJobV1, "xgbReplicaSpecs", map[string][2]string{
			"Master": {"Master", "master"},
			"Worker": {"Worker", "worker"},
		}),
		Entry("JAX v1", BuildKubeflowJAXJobV1, "jaxReplicaSpecs", map[string][2]string{
			"Worker": {"Worker", "worker"},
		}),
		Entry("PyTorch v1", BuildKubeflowPyTorchJobV1, "pytorchReplicaSpecs", map[string][2]string{
			"master": {"Master", "master"},
			"worker": {"Worker", "worker"},
		}),
	)
})
