/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pkg

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"sigs.k8s.io/controller-runtime/pkg/cache"
)

// TestCacheOptionsFiltersPodsByRunaiScheduler ensures non-Run:ai pods are excluded
// server-side before reaching the controller-runtime informer cache.
func TestCacheOptionsFiltersPodsByRunaiScheduler(t *testing.T) {
	options := cacheOptions()
	require.NotNil(t, options.ByObject)

	var podByObject *cache.ByObject
	for obj, byObject := range options.ByObject {
		if _, ok := obj.(*v1.Pod); ok {
			b := byObject
			podByObject = &b
			break
		}
	}

	require.NotNil(t, podByObject, "expected *v1.Pod to be configured in cache.Options.ByObject")
	require.NotNil(t, podByObject.Field, "expected field selector to be set for *v1.Pod cache")

	assert.True(t, podByObject.Field.Matches(fields.Set{
		podSchedulerNameField: constants.RunaiScheduler,
	}))
	assert.False(t, podByObject.Field.Matches(fields.Set{
		podSchedulerNameField: "default-scheduler",
	}))
}

// TestCacheOptionsPreservesOtherObjectsUnfiltered ensures CRDs like PodGroup and Queue
// (which do not support spec.schedulerName field selectors) remain unfiltered.
func TestCacheOptionsPreservesOtherObjectsUnfiltered(t *testing.T) {
	options := cacheOptions()
	assert.Len(t, options.ByObject, 1)
	for obj := range options.ByObject {
		_, isPod := obj.(*v1.Pod)
		assert.True(t, isPod, "expected only *v1.Pod in ByObject, got %T", obj)
	}
}
