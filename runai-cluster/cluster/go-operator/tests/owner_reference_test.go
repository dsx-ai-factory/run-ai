/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"fmt"

	kc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"

	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/utils"
	. "github.com/onsi/ginkgo/v2"

	ccontroller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/controllers/run"
	. "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/tests/utils"
	. "github.com/onsi/gomega"
)

const (
	// NamespacedNameAnnotation is an annotation whose value encodes the name and namespace of a resource to
	// reconcile when a resource containing this annotation changes. Valid values are of the form
	// `<namespace>/<name>` for namespace-scoped owners and `<name>` for cluster-scoped owners.
	NamespacedNameAnnotation = "operator-sdk/primary-resource"
	// TypeAnnotation is an annotation whose value encodes the group and kind of a resource to reconcil when a
	// resource containing this annotation changes. Valid values are of the form `<Kind>` for resource in the
	// core group, and `<Kind>.<group>` for all other resources.
	TypeAnnotation = "operator-sdk/primary-resource-type"
)

var _ = Describe("Owner Reference", func() {
	var (
		hybridClient kc.HybridClient
	)
	BeforeEach(func() {
		hybridClient = kc.New(k8s, k8s, nil)
	})
	It("cluster-scoped resources should have owner annotations", func() {
		runaiConfig := GetTestingRunaiConfig(k8s)
		WaitForDeployedCondition(k8s, runaiConfig)
		managedOperands := ccontroller.NewManagedOperands(apiCtx, hybridClient, hybridClient, runaiConfig)
		resources, err := ccontroller.ResourcesForOperands(apiCtx,
			managedOperands.GetRemoteClusterEnabled())
		ExpectNoErr(err)

		enabledResources := resources.CreateOrUpdate

		for _, resource := range enabledResources {
			obj, err := resource.Object(apiCtx)
			ExpectNoErr(err)
			EventuallyWithTimeout(func() bool {
				return k8s.Get(apiCtx, utils.NamespacedNameFor(obj), obj) == nil
			}).Should(BeTrue())

			if obj.GetNamespace() == "" {
				By(fmt.Sprintf("checking owner annotations for cluster-scoped resource: %s",
					utils.ResourceUniqueID(obj)))

				Expect(obj.GetAnnotations()).NotTo(BeEmpty())
				Expect(obj.GetAnnotations()[runv1.NamespacedNameAnnotation]).NotTo(BeEmpty())
				Expect(obj.GetAnnotations()[runv1.TypeAnnotation]).NotTo(BeEmpty())
			}
		}
	})

	It("resources in RunaiConfig's namespace should have owner reference", func() {
		managedOperands := ccontroller.NewManagedOperands(apiCtx, hybridClient, hybridClient, runaiConfig)
		resources, err := ccontroller.ResourcesForOperands(apiCtx,
			managedOperands.GetRemoteClusterEnabled())
		ExpectNoErr(err)

		enabledResources := resources.CreateOrUpdate

		for _, resource := range enabledResources {
			obj, err := resource.Object(apiCtx)
			ExpectNoErr(err)
			ExpectGetResource(k8s, obj)

			if obj.GetNamespace() == runaiConfig.GetNamespace() {
				By(fmt.Sprintf("checking owner annotations for namespaced resource: %s",
					utils.ResourceUniqueID(obj)))

				Expect(obj.GetOwnerReferences()).NotTo(BeEmpty())
				Expect(obj.GetOwnerReferences()[0].UID).To(Equal(runaiConfig.GetUID()))
			}
		}
	})

	It("resources in a different namespace than RunaiConfig should have owner annotations", func() {
		managedOperands := ccontroller.NewManagedOperands(apiCtx, hybridClient, hybridClient, runaiConfig)
		resources, err := ccontroller.ResourcesForOperands(apiCtx,
			managedOperands.GetRemoteClusterEnabled())
		ExpectNoErr(err)

		enabledResources := resources.CreateOrUpdate

		for _, resource := range enabledResources {
			obj, err := resource.Object(apiCtx)
			ExpectNoErr(err)
			ExpectGetResource(k8s, obj)

			if obj.GetNamespace() != runaiConfig.GetNamespace() &&
				obj.GetNamespace() != "" {
				By(fmt.Sprintf("checking owner annotations for namespaced resource: %s",
					utils.ResourceUniqueID(obj)))

				Expect(obj.GetAnnotations()).NotTo(BeEmpty())
				Expect(obj.GetAnnotations()[NamespacedNameAnnotation]).NotTo(BeEmpty())
				Expect(obj.GetAnnotations()[TypeAnnotation]).NotTo(BeEmpty())
			}
		}
	})
})
