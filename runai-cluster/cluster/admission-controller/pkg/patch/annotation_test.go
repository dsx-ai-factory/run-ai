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

func TestAddSingleAnnotation(t *testing.T) {
	objectMeta := &metav1.ObjectMeta{}
	AddAnnotationToObjectMeta(objectMeta, "key", "value")

	assert.Equal(t, len(objectMeta.Annotations), 1)
	annotation, found := objectMeta.Annotations["key"]
	assert.True(t, found)
	assert.Equal(t, annotation, "value")
}

func TestAddMultipleAnnotation(t *testing.T) {
	objectMeta := &metav1.ObjectMeta{}
	AddAnnotationToObjectMeta(objectMeta, "key1", "value1")
	AddAnnotationToObjectMeta(objectMeta, "key2", "value2")

	assert.Equal(t, len(objectMeta.Annotations), 2)
	annotation, found := objectMeta.Annotations["key1"]
	assert.True(t, found)
	assert.Equal(t, annotation, "value1")

	annotation, found = objectMeta.Annotations["key2"]
	assert.True(t, found)
	assert.Equal(t, annotation, "value2")
}

func TestOverrideAnnotation(t *testing.T) {
	objectMeta := &metav1.ObjectMeta{
		Annotations: map[string]string{"key": "value1"},
	}
	AddAnnotationToObjectMeta(objectMeta, "key", "value2")

	assert.Equal(t, len(objectMeta.Annotations), 1)
	annotation, found := objectMeta.Annotations["key"]
	assert.True(t, found)
	assert.Equal(t, annotation, "value2")
}
