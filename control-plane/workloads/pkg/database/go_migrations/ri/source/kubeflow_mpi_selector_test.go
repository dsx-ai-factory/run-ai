/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package source

import (
	"encoding/json"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/itchyny/gojq"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("MPI Karta pod selectors", func() {
	const (
		launcherRole  = "launcher"
		workerRole    = "worker"
		workerSpecKey = "Worker"
	)

	DescribeTable("identifies launcher and worker pods using labels emitted by each operator",
		func(build func() models.Karta, roleLabel, unrelatedRoleLabel string) {
			children := build().Spec.StructureDefinition.ChildComponents
			Expect(children).To(HaveLen(2))

			for _, child := range children {
				Expect(child.Name).To(BeElementOf(launcherRole, workerRole))
				Expect(child.PodSelector).NotTo(BeNil())
				Expect(child.PodSelector.ComponentTypeSelector).NotTo(BeNil())
				Expect(child.PodSelector.ComponentTypeSelector.Value).NotTo(BeNil())
				Expect(child.SpecDefinition).NotTo(BeNil())
				Expect(child.SpecDefinition.PodTemplateSpecPath).NotTo(BeNil())

				specKey := workerSpecKey
				if child.Name == launcherRole {
					specKey = "Launcher"
				}
				Expect(*child.SpecDefinition.PodTemplateSpecPath).To(Equal(".spec.mpiReplicaSpecs." + specKey + ".template"))
				selector := child.PodSelector.ComponentTypeSelector
				query, err := gojq.Parse(fmt.Sprintf("%s == %q", selector.KeyPath, *selector.Value))
				Expect(err).NotTo(HaveOccurred())

				for _, pod := range []struct {
					name   string
					labels map[string]string
					role   string
				}{
					{name: launcherRole, labels: map[string]string{roleLabel: launcherRole}, role: launcherRole},
					{name: "worker-0", labels: map[string]string{roleLabel: workerRole}, role: workerRole},
					{name: "worker-1", labels: map[string]string{roleLabel: workerRole}, role: workerRole},
					{name: "missing-role"},
					{name: "unknown-role", labels: map[string]string{roleLabel: "unknown"}},
					{name: "unrelated-role", labels: map[string]string{unrelatedRoleLabel: workerRole}},
					{name: "launcher-with-master-role", labels: map[string]string{roleLabel: launcherRole, unrelatedRoleLabel: "master"}, role: launcherRole},
				} {
					data, err := json.Marshal(&corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{Name: pod.name, Labels: pod.labels},
					})
					Expect(err).NotTo(HaveOccurred())
					var object map[string]any
					Expect(json.Unmarshal(data, &object)).To(Succeed())
					matches, ok := query.Run(object).Next()
					Expect(ok).To(BeTrue())
					Expect(matches).To(Equal(pod.role == child.Name), "pod %q against component %q", pod.name, child.Name)
				}
			}
		},
		Entry("Training Operator MPI v1", BuildKubeflowMPIJobV1, "training.kubeflow.org/replica-type", "training.kubeflow.org/job-role"),
		Entry("standalone MPI Operator v2beta1", BuildKubeflowMPIJobV2beta1, "training.kubeflow.org/job-role", "training.kubeflow.org/replica-type"),
	)
})
