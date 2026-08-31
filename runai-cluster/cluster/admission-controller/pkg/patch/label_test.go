/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package patch

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAddSingleLabel(t *testing.T) {
	objectMeta := &metav1.ObjectMeta{}
	AddLabelToObjectMeta(objectMeta, "key", "value")

	assert.Equal(t, len(objectMeta.Labels), 1)
	annotation, found := objectMeta.Labels["key"]
	assert.True(t, found)
	assert.Equal(t, annotation, "value")
}

func TestAddMultipleLabel(t *testing.T) {
	objectMeta := &metav1.ObjectMeta{}
	AddLabelToObjectMeta(objectMeta, "key1", "value1")
	AddLabelToObjectMeta(objectMeta, "key2", "value2")

	assert.Equal(t, len(objectMeta.Labels), 2)
	annotation, found := objectMeta.Labels["key1"]
	assert.True(t, found)
	assert.Equal(t, annotation, "value1")

	annotation, found = objectMeta.Labels["key2"]
	assert.True(t, found)
	assert.Equal(t, annotation, "value2")
}

func TestOverrideLabel(t *testing.T) {
	objectMeta := &metav1.ObjectMeta{
		Labels: map[string]string{"key": "value1"},
	}
	AddLabelToObjectMeta(objectMeta, "key", "value2")

	assert.Equal(t, len(objectMeta.Labels), 1)
	annotation, found := objectMeta.Labels["key"]
	assert.True(t, found)
	assert.Equal(t, annotation, "value2")
}
