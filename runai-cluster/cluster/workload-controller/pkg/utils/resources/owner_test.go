/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Owner", func() {
	var (
		ownerUid types.UID
		object   client.Object
	)

	BeforeEach(func() {
		ownerUid = "owner-uid"
		object = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				OwnerReferences: []metav1.OwnerReference{
					{
						UID: ownerUid,
					},
				},
			},
		}
	})

	Context("IsObjectOwnedBy", func() {
		It("should return true if object is owned by owner", func() {
			Expect(IsObjectOwnedBy(object, ownerUid)).To(BeTrue())
		})

		It("should return false if object is not owned by owner", func() {
			Expect(IsObjectOwnedBy(object, "other-owner-uid")).To(BeFalse())
		})
	})
	Context("BackwardCompatibilityIsObjectCreatedBy", func() {
		It("should return false, if object has owner reference", func() {
			Expect(BackwardCompatibilityIsObjectCreatedBy(object, "owner-name")).To(BeFalse())
		})
		It("should return false, if object has no labels", func() {
			object.SetOwnerReferences(nil)
			Expect(BackwardCompatibilityIsObjectCreatedBy(object, "owner-name")).To(BeFalse())
		})
		It("should return false, if object has no label release", func() {
			object.SetOwnerReferences(nil)
			object.SetLabels(map[string]string{"other-label": "other-value"})
			Expect(BackwardCompatibilityIsObjectCreatedBy(object, "owner-name")).To(BeFalse())
		})
		It("should return false, if object has label release but not equal to owner name", func() {
			object.SetOwnerReferences(nil)
			object.SetLabels(map[string]string{constants.LabelRelease: "other-owner-name"})
			Expect(BackwardCompatibilityIsObjectCreatedBy(object, "owner-name")).To(BeFalse())
		})
		It("should return true, if object has label release and equal to owner name", func() {
			object.SetOwnerReferences(nil)
			object.SetLabels(map[string]string{constants.LabelRelease: "owner-name"})
			Expect(BackwardCompatibilityIsObjectCreatedBy(object, "owner-name")).To(BeTrue())
		})

	})

})
