/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package webhook

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/config"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestNodePoolWebhook(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "NodePool Webhook Suite")
}

func webhookTestScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
	return scheme
}

func nodePool(name, labelKey, labelValue string) *v1alpha1.NodePool {
	return &v1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: v1alpha1.NodePoolSpec{
			LabelKey:   labelKey,
			LabelValue: labelValue,
		},
	}
}

var _ = Describe("NodePool duplicate-label validation", func() {
	var (
		ctx     context.Context
		scheme  *runtime.Scheme
		newPool *v1alpha1.NodePool
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = webhookTestScheme()
		DeferCleanup(config.SetForTest(&config.NodePoolControllerConfig{DefaultNodepoolName: "default"}))
	})

	validatorWith := func(existing ...client.Object) *nodePoolValidator {
		c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(existing...).Build()
		return &nodePoolValidator{client: c}
	}

	Describe("ValidateCreate", func() {
		It("allows a nodepool with a unique labelKey/labelValue pair", func() {
			v := validatorWith(nodePool("pool-a", "gpu", "a100"))
			newPool = nodePool("pool-b", "gpu2", "a1002")

			_, err := v.ValidateCreate(ctx, newPool)
			Expect(err).ToNot(HaveOccurred())
		})

		It("rejects a nodepool duplicating an existing labelKey/labelValue pair", func() {
			v := validatorWith(nodePool("pool-a", "gpu", "a100"))
			newPool = nodePool("pool-b", "gpu", "a100")

			_, err := v.ValidateCreate(ctx, newPool)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("pool-a"))
			Expect(err.Error()).To(ContainSubstring("pool-b"))
		})

		It("allows a nodepool that shares only the labelKey but not the labelValue", func() {
			v := validatorWith(nodePool("pool-a", "gpu", "a100"))
			newPool = nodePool("pool-b", "gpu", "h100")

			_, err := v.ValidateCreate(ctx, newPool)
			Expect(err).ToNot(HaveOccurred())
		})

		It("allows the default nodepool with an empty labelKey/labelValue", func() {
			v := validatorWith()
			newPool = nodePool(config.Get().DefaultNodepoolName, "", "")

			_, err := v.ValidateCreate(ctx, newPool)
			Expect(err).ToNot(HaveOccurred())
		})

		It("rejects the default nodepool when it sets a labelKey/labelValue", func() {
			v := validatorWith()
			newPool = nodePool(config.Get().DefaultNodepoolName, "gpu", "a100")

			_, err := v.ValidateCreate(ctx, newPool)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("must not set"))
		})

		It("rejects a non-default nodepool with an empty labelKey", func() {
			v := validatorWith()
			newPool = nodePool("pool-a", "", "a100")

			_, err := v.ValidateCreate(ctx, newPool)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("non-empty"))
		})

		It("rejects a non-default nodepool with an empty labelValue", func() {
			v := validatorWith()
			newPool = nodePool("pool-a", "gpu", "")

			_, err := v.ValidateCreate(ctx, newPool)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("non-empty"))
		})

		It("rejects a non-default nodepool with an empty labelKey and labelValue", func() {
			v := validatorWith()
			newPool = nodePool("pool-a", "", "")

			_, err := v.ValidateCreate(ctx, newPool)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("non-empty"))
		})

		It("rejects when it collides with any of several existing nodepools", func() {
			v := validatorWith(
				nodePool("pool-a", "gpu", "a100"),
				nodePool("pool-b", "gpu", "h100"),
				nodePool("pool-c", "region", "us-east"),
			)
			newPool = nodePool("pool-d", "gpu", "h100")

			_, err := v.ValidateCreate(ctx, newPool)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("pool-b"))
		})

		It("returns an error for a non-NodePool object", func() {
			v := validatorWith()

			_, err := v.ValidateCreate(ctx, &v1alpha1.NodePoolList{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("expected a NodePool"))
		})
	})

	Describe("ValidateUpdate", func() {
		// ValidateUpdate is a safety net that only rejects a changed labelKey/labelValue
		// pair; the pair is meant to be immutable via the CRD. A no-op update passes.
		It("allows a no-op update that keeps a non-default pool's pair", func() {
			existing := nodePool("pool-a", "gpu", "a100")
			v := validatorWith(existing)
			updated := nodePool("pool-a", "gpu", "a100")

			_, err := v.ValidateUpdate(ctx, existing, updated)
			Expect(err).ToNot(HaveOccurred())
		})

		It("allows a no-op update of the default nodepool's empty pair", func() {
			existing := nodePool(config.Get().DefaultNodepoolName, "", "")
			v := validatorWith(existing)
			updated := nodePool(config.Get().DefaultNodepoolName, "", "")

			_, err := v.ValidateUpdate(ctx, existing, updated)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("ValidateDelete", func() {
		// This test is to verify the deletion is not blocked even though before the deletion, the nodepool still
		// exists and the labelKey/labelValue pair is not unique.
		It("does not block deletion of a non-default nodepool", func() {
			v := validatorWith(nodePool("pool-a", "gpu", "a100"))

			_, err := v.ValidateDelete(ctx, nodePool("pool-a", "gpu", "a100"))
			Expect(err).ToNot(HaveOccurred())
		})

		It("refuses to delete the default nodepool", func() {
			defaultName := config.Get().DefaultNodepoolName
			v := validatorWith(nodePool(defaultName, "", ""))

			_, err := v.ValidateDelete(ctx, nodePool(defaultName, "", ""))
			Expect(err).To(MatchError(ContainSubstring(defaultName)))
		})

		It("follows the configured default nodepool name", func() {
			DeferCleanup(config.SetForTest(&config.NodePoolControllerConfig{DefaultNodepoolName: "all-nodes"}))
			v := validatorWith()

			_, err := v.ValidateDelete(ctx, nodePool("all-nodes", "", ""))
			Expect(err).To(HaveOccurred())

			_, err = v.ValidateDelete(ctx, nodePool("default", "gpu", "a100"))
			Expect(err).ToNot(HaveOccurred(), "the name is configurable, not hardcoded")
		})

		It("rejects an object that is not a NodePool", func() {
			v := validatorWith()

			_, err := v.ValidateDelete(ctx, &v1alpha1.Project{})
			Expect(err).To(MatchError(ContainSubstring("expected a NodePool")))
		})
	})
})
