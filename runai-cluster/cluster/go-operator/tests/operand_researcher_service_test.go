/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	researcher_service "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/researcher-service"
	researcher_service_resources "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/researcher-service/resources"
	kc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	. "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/tests/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

// researcher-service was retired in 2.26.0: its source and image build were
// removed, but the operator kept reconciling its Deployment. These specs pin the
// retirement — the operand must stay disabled regardless of the RunaiConfig
// field, and must clean up the objects it used to own.
var _ = Describe("Researcher Service Operand (retired)", func() {
	BeforeEach(func() {
		// The value that reproduces the bug: the product default through 2.25.
		runaiConfig.Spec.ResearcherService.Enabled = ptr.To[bool](true)
		EventuallyUpdateResource(kc.New(k8s, k8s, nil), runaiConfig)
		WaitForDeployedCondition(kc.New(k8s, k8s, nil), runaiConfig)
	})

	Context("Operand", func() {
		It("stays disabled even when researcher-service.enabled is true", func() {
			op := researcher_service.Operand(kc.New(k8s, k8s, nil), runaiConfig)
			Expect(op.Enabled()).To(BeFalse())
		})

		It("reports every resource disabled", func() {
			op := researcher_service.Operand(kc.New(k8s, k8s, nil), runaiConfig)
			Expect(op.Resources()).NotTo(BeEmpty())

			for _, res := range op.Resources() {
				Expect(res.Enabled()).To(BeFalse())
			}
		})

		// Every resource is routed to deletion, so Object is called on the
		// delete path. It must yield a bare object identifying what to delete
		// rather than erroring — the Deployment in particular used to fail here
		// when the control-plane oauth config was absent.
		It("builds bare, non-erroring objects for the delete path", func() {
			op := researcher_service.Operand(kc.New(k8s, k8s, nil), runaiConfig)

			for _, res := range op.Resources() {
				obj, err := res.Object(apiCtx)
				ExpectNoErr(err)
				Expect(obj).NotTo(BeNil())
				Expect(obj.GetName()).NotTo(BeEmpty())
				Expect(obj.GetNamespace()).To(Equal(runaiConfig.Namespace))
			}
		})

		It("does not create a researcher-service Deployment", func() {
			ExpectDoesNotExist(k8s, &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      researcher_service_resources.DeploymentName,
					Namespace: runaiConfig.Namespace,
				},
			})
		})
	})

	Context("Upgrade cleanup", func() {
		// The upgrade case from the bug report: a cluster arrives at 2.26 with a
		// researcher-service Deployment left over from 2.25. The operator must
		// remove it instead of reconciling it onto a tag that has no image.
		It("deletes a leftover researcher-service Deployment", func() {
			leftover := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      researcher_service_resources.DeploymentName,
					Namespace: runaiConfig.Namespace,
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": researcher_service_resources.DeploymentName},
					},
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"app": researcher_service_resources.DeploymentName},
						},
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  researcher_service_resources.DeploymentName,
									Image: "researcher-service:2.25.0",
								},
							},
						},
					},
				},
			}

			ExpectCreateResource(k8s, leftover)

			TriggerController(kc.New(k8s, k8s, nil), runaiConfig)

			EventuallyWithTimeout(func() bool {
				return ResourceExists(k8s, &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      researcher_service_resources.DeploymentName,
						Namespace: runaiConfig.Namespace,
					},
				})
			}).Should(BeFalse())
		})
	})

	Context("Ingress", func() {
		// IngressStatus backs the cluster-wide ingress-controller readiness
		// check. With the component retired it must report ready rather than
		// hold the cluster unhealthy.
		It("reports the ingress status ready", func() {
			ingressRes := researcher_service_resources.NewIngressResource(
				"researcher-service", kc.New(k8s, k8s, nil), runaiConfig)
			Expect(ingressRes.Enabled()).To(BeFalse())

			runaiConfig.Spec.Internal.Global.ClusterURL = "https://test.com"
			runaiConfig.Spec.Internal.ResearcherService.Ingress.Enabled = ptr.To[bool](true)
			EventuallyUpdateResource(kc.New(k8s, k8s, nil), runaiConfig)
			WaitForDeployedCondition(kc.New(k8s, k8s, nil), runaiConfig)

			status, err := researcher_service_resources.IngressStatus(apiCtx, kc.New(k8s, k8s, nil), runaiConfig)
			ExpectNoErr(err)
			Expect(status.Ready).To(BeTrue())
		})
	})
})
