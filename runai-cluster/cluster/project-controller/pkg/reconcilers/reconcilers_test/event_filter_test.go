/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package reconcilers_test

import (
	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/common"
	"github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/config"
	. "github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/reconcilers"
	. "github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/test"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const testNamespaceRunaiCool = "runai-cool"

var _ = Describe("Event Filter Tests", func() {

	Context("Project Filter", func() {

		var (
			namespace *corev1.Namespace
			project   *kaiv1alpha1.Project
			queue     *kaiv2.Queue
		)

		BeforeEach(func() {
			namespace = TestNamespace.DeepCopy()
			queue = TestQueue.DeepCopy()
			project = TestProject.DeepCopy()
		})

		It("Filters Namespace events", func() {

			By("Allowing namespaces with the 'runai/queue' label", func() {
				namespace.Name = testNamespaceRunaiCool
				Expect(FilterProjectEvent(namespace.DeepCopy())).To(BeTrue())
			})

			By("Allowing namespaces with the 'runai/queue' label - even without runai prefix", func() {
				namespace.Name = "non-runai-prefix"
				Expect(FilterProjectEvent(namespace.DeepCopy())).To(BeTrue())
			})

			By("Blocking namespaces without the 'runai/queue' label - even with runai prefix", func() {
				namespace.Name = testNamespaceRunaiCool
				delete(namespace.Labels, config.Get().NamespaceProjectLabelKey)
				Expect(FilterProjectEvent(namespace.DeepCopy())).To(BeFalse())
			})

			By("Blocking namespaces without the 'runai/queue' label - without runai prefix", func() {
				namespace.Name = "non-runai-prefix"
				delete(namespace.Labels, config.Get().NamespaceProjectLabelKey)
				Expect(FilterProjectEvent(namespace.DeepCopy())).To(BeFalse())
			})

			By("Blocking namespaces that have a blank 'runai/queue' label", func() {
				namespace.Name = testNamespaceRunaiCool
				namespace.Labels[config.Get().NamespaceProjectLabelKey] = ""
				Expect(FilterProjectEvent(namespace.DeepCopy())).To(BeFalse())
			})
		})

		It("Filters Queue events", func() {
			By("Allowing all Queue events", func() {
				Expect(FilterProjectEvent(queue.DeepCopy())).To(BeTrue())
			})
		})

		It("Filters Project events", func() {
			By("Allowing all Project events", func() {
				Expect(FilterProjectEvent(project.DeepCopy())).To(BeTrue())
			})
		})

		It("allows only the configured role-bindings ConfigMap when role-binding reconciliation is enabled", func() {
			projectConfig := RunaiConfigForTests()
			projectConfig.CreateRoleBindings = true
			projectConfig.RoleBindingsCm = "role-bindings"
			projectConfig.RoleBindingsCmNamespace = "runai"
			restore := config.SetForTest(projectConfig)
			defer restore()

			matchingConfigMap := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
				Name: projectConfig.RoleBindingsCm, Namespace: projectConfig.RoleBindingsCmNamespace,
			}}
			wrongName := matchingConfigMap.DeepCopy()
			wrongName.Name = "other"
			wrongNamespace := matchingConfigMap.DeepCopy()
			wrongNamespace.Namespace = "other"

			Expect(FilterProjectEvent(matchingConfigMap)).To(BeTrue())
			Expect(FilterProjectEvent(wrongName)).To(BeFalse())
			Expect(FilterProjectEvent(wrongNamespace)).To(BeFalse())

			projectConfig.CreateRoleBindings = false
			Expect(FilterProjectEvent(matchingConfigMap)).To(BeFalse())
		})

		It("Blocks manually overridden resources", func() {

			By("Blocking a manually overridden Namespace", func() {
				namespace.Labels = map[string]string{RunaiResourceManualOverrideLabel: "true"}
				Expect(FilterProjectEvent(namespace.DeepCopy())).To(BeFalse())
			})

			By("Blocking a manually overridden Project", func() {
				project.Labels = map[string]string{RunaiResourceManualOverrideLabel: "true"}
				Expect(FilterProjectEvent(namespace.DeepCopy())).To(BeFalse())
			})

			By("Blocking a manually overridden Queue", func() {
				queue.Labels = map[string]string{RunaiResourceManualOverrideLabel: "true"}
				Expect(FilterProjectEvent(queue.DeepCopy())).To(BeFalse())
			})
		})
	})

	Context("Configmap Filter", func() {

		var limitRangeConfigMap, unrelatedRunaiConfigMap, unrelatedConfigMap corev1.ConfigMap

		BeforeEach(func() {
			limitRangeConfigMap = *DefaultLimitRangeConfigMap.DeepCopy()

			unrelatedConfigMap = corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       ConfigMapKind,
					APIVersion: CoreV1ApiVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "unrelated",
					Namespace: "default",
				},
			}
			unrelatedRunaiConfigMap = corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       ConfigMapKind,
					APIVersion: CoreV1ApiVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "unrelated-namespaced",
					Namespace: config.Get().InstallNamespace,
				},
			}
		})

		It("Filters Configmap events", func() {

			By("Allowing the default LimitRange ConfigMap", func() {
				Expect(FilterConfigmapEvent(limitRangeConfigMap.DeepCopy())).To(BeTrue())
			})

			By("Blocking all other Secrets", func() {
				Expect(FilterConfigmapEvent(unrelatedConfigMap.DeepCopy())).To(BeFalse())
				Expect(FilterConfigmapEvent(unrelatedRunaiConfigMap.DeepCopy())).To(BeFalse())
			})
		})
	})
})
