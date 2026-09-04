/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TestExternalWorkloadIntegrator", Ordered, func() {
	var externalWorkload *v1alpha1.ExternalWorkload
	var childToRemove v1alpha1.ChildId
	BeforeAll(func() {
		childToRemove := v1alpha1.ChildId{
			UID:  "ID-1",
			Type: "Type-1",
		}

		externalWorkload = &v1alpha1.ExternalWorkload{
			Status: v1alpha1.ExternalWorkloadStatus{
				ObservedStatus: v1alpha1.ExternalWorkloadObservedStatus{
					ChildIds: []v1alpha1.ChildId{childToRemove,
						{
							UID:  "ID-2",
							Type: "Type-1",
						},
						{
							UID:  "ID-3",
							Type: "Type-1",
						},
					},
				},
			},
		}
	})
	It("Test objects in namespace with enforce runai scheduler", func() {
		exists, index := IsChildIdInStatus(externalWorkload, childToRemove)
		Expect(exists).To(BeTrue())
		Expect(index).To(Equal(-1))

		notExistingChildId := v1alpha1.ChildId{
			UID: "ID-notexisting",
		}
		exists, index = IsChildIdInStatus(externalWorkload, notExistingChildId)
		Expect(exists).To(BeFalse())
		Expect(index).To(Equal(0))
	})
})
