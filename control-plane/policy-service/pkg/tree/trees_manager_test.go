/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tree

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TreesManager", func() {
	Context("add and get", func() {
		It("each tree added once", func() {
			ctx := context.Background()
			w1a, _ := HierarchyTrees.Get(ctx, 1, WorkloadTypeWorkspace)
			t2a, _ := HierarchyTrees.Get(ctx, 2, WorkloadTypeTraining)
			t2b, _ := HierarchyTrees.Get(ctx, 2, WorkloadTypeTraining)
			w2a, _ := HierarchyTrees.Get(ctx, 2, WorkloadTypeWorkspace)
			d1a, _ := HierarchyTrees.Get(ctx, 1, WorkloadTypeDistributed)
			d1b, _ := HierarchyTrees.Get(ctx, 1, WorkloadTypeDistributed)
			i1a, _ := HierarchyTrees.Get(ctx, 1, WorkloadTypeInference)
			i2a, _ := HierarchyTrees.Get(ctx, 2, WorkloadTypeInference)

			Expect(t2a.treeId.ToString()).To(Equal("2:Training"))
			Expect(t2b.treeId.ToString()).To(Equal("2:Training"))
			Expect(t2a).To(Equal(t2b))

			Expect(w1a.treeId.ToString()).To(Equal("1:Workspace"))
			Expect(w2a.treeId.ToString()).To(Equal("2:Workspace"))

			Expect(d1a.treeId.ToString()).To(Equal("1:Distributed"))
			Expect(d1b.treeId.ToString()).To(Equal("1:Distributed"))
			Expect(d1a).To(Equal(d1b))

			Expect(i1a.treeId.ToString()).To(Equal("1:Inference"))
			Expect(i2a.treeId.ToString()).To(Equal("2:Inference"))
		})
	})
})
