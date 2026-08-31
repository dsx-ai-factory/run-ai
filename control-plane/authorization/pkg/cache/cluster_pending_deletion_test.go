/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cache

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-common-packages/cache/v2"
)

var _ = Describe("Cluster pending deletion cache", func() {
	const clusterUUID = "aabbccdd-1111-2222-3333-aabbccddeeff"

	ctx := context.Background()

	BeforeEach(func() {
		cache.ClearCache()
	})

	AfterEach(func() {
		cache.ClearCache()
	})

	Describe("pendingDeletionKey", func() {
		It("formats the key with cluster UUID", func() {
			Expect(pendingDeletionKey("abc-123")).To(Equal("cluster-pending-deletion:abc-123"))
		})
	})

	Describe("IsClusterPendingDeletion", func() {
		It("returns false when no marker has been set", func() {
			Expect(IsClusterPendingDeletion(ctx, clusterUUID)).To(BeFalse())
		})

		It("returns true after MarkClusterPendingDeletion is called", func() {
			MarkClusterPendingDeletion(ctx, clusterUUID)
			Expect(IsClusterPendingDeletion(ctx, clusterUUID)).To(BeTrue())
		})

		It("returns false for a different cluster UUID even after another is marked", func() {
			MarkClusterPendingDeletion(ctx, clusterUUID)
			Expect(IsClusterPendingDeletion(ctx, "other-uuid")).To(BeFalse())
		})
	})

	Describe("MarkClusterPendingDeletion", func() {
		It("sets a marker that IsClusterPendingDeletion detects", func() {
			MarkClusterPendingDeletion(ctx, clusterUUID)
			Expect(IsClusterPendingDeletion(ctx, clusterUUID)).To(BeTrue())
		})

		It("is idempotent - marking twice still yields a detectable marker", func() {
			MarkClusterPendingDeletion(ctx, clusterUUID)
			MarkClusterPendingDeletion(ctx, clusterUUID)
			Expect(IsClusterPendingDeletion(ctx, clusterUUID)).To(BeTrue())
		})
	})

	Describe("RemoveClusterPendingDeletion", func() {
		It("clears the marker so IsClusterPendingDeletion returns false", func() {
			MarkClusterPendingDeletion(ctx, clusterUUID)
			Expect(IsClusterPendingDeletion(ctx, clusterUUID)).To(BeTrue())

			RemoveClusterPendingDeletion(ctx, clusterUUID)
			Expect(IsClusterPendingDeletion(ctx, clusterUUID)).To(BeFalse())
		})

		It("is a no-op when no marker exists and does not panic", func() {
			Expect(func() { RemoveClusterPendingDeletion(ctx, clusterUUID) }).NotTo(Panic())
		})

		It("only removes the marker for the specified cluster UUID", func() {
			const otherUUID = "other-uuid-5678"
			MarkClusterPendingDeletion(ctx, clusterUUID)
			MarkClusterPendingDeletion(ctx, otherUUID)

			RemoveClusterPendingDeletion(ctx, clusterUUID)

			Expect(IsClusterPendingDeletion(ctx, clusterUUID)).To(BeFalse())
			Expect(IsClusterPendingDeletion(ctx, otherUUID)).To(BeTrue())
		})
	})

	Describe("full lifecycle", func() {
		It("mark, detect, remove, undetected", func() {
			Expect(IsClusterPendingDeletion(ctx, clusterUUID)).To(BeFalse())

			MarkClusterPendingDeletion(ctx, clusterUUID)
			Expect(IsClusterPendingDeletion(ctx, clusterUUID)).To(BeTrue())

			RemoveClusterPendingDeletion(ctx, clusterUUID)
			Expect(IsClusterPendingDeletion(ctx, clusterUUID)).To(BeFalse())
		})
	})

	Describe("initPendingDeletionCache", func() {
		It("marks all cluster UUIDs returned by the fetch function", func() {
			const secondUUID = "11223344-aaaa-bbbb-cccc-ddeeff001122"
			fetchFn := func(_ context.Context) ([]string, error) {
				return []string{clusterUUID, secondUUID}, nil
			}

			Expect(initPendingDeletionCache(ctx, fetchFn)).To(Succeed())

			Expect(IsClusterPendingDeletion(ctx, clusterUUID)).To(BeTrue())
			Expect(IsClusterPendingDeletion(ctx, secondUUID)).To(BeTrue())
		})

		It("is a no-op when the fetch function returns an empty list", func() {
			fetchFn := func(_ context.Context) ([]string, error) {
				return []string{}, nil
			}

			Expect(initPendingDeletionCache(ctx, fetchFn)).To(Succeed())
			Expect(IsClusterPendingDeletion(ctx, clusterUUID)).To(BeFalse())
		})

		It("returns an error when the fetch function fails", func() {
			fetchFn := func(_ context.Context) ([]string, error) {
				return nil, errors.New("db unavailable")
			}

			Expect(initPendingDeletionCache(ctx, fetchFn)).To(MatchError(ContainSubstring("db unavailable")))
		})

		It("does not clear markers that were already set before Init is called", func() {
			MarkClusterPendingDeletion(ctx, clusterUUID)

			const newUUID = "99887766-5544-3322-1100-aabbccddeeff"
			fetchFn := func(_ context.Context) ([]string, error) {
				return []string{newUUID}, nil
			}

			Expect(initPendingDeletionCache(ctx, fetchFn)).To(Succeed())

			Expect(IsClusterPendingDeletion(ctx, clusterUUID)).To(BeTrue())
			Expect(IsClusterPendingDeletion(ctx, newUUID)).To(BeTrue())
		})
	})
})
