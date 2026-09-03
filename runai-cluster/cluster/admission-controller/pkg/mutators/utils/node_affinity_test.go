/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
)

// The node types live in an annotation on the kai.resources Project CRD.

func projectWithAnnotationOnly(key string, nodeTypes []string) *unstructured.Unstructured {
	encoded, err := json.Marshal(nodeTypes)
	if err != nil {
		panic(err)
	}

	project := &unstructured.Unstructured{Object: map[string]interface{}{}}
	project.SetName("test-project")
	project.SetAnnotations(map[string]string{key: string(encoded)})

	return project
}

func TestNodeAffinityTypesOfProjectFromAnnotation(t *testing.T) {
	nodeTypes := []string{"aaa", "bbb"}

	for _, tc := range []struct {
		name        string
		interactive bool
		key         string
	}{
		{"train", false, commonconstants.NodeAffinityTrainAnnotation},
		{"interactive", true, commonconstants.NodeAffinityInteractiveAnnotation},
	} {
		t.Run(tc.name, func(t *testing.T) {
			project := projectWithAnnotationOnly(tc.key, nodeTypes)

			assert.Equal(t, nodeTypes, NodeAffinityTypesFromAnnotation(project, tc.interactive))
		})
	}
}

func TestNodeAffinityTypesOfProjectMissingAndMalformed(t *testing.T) {
	empty := &unstructured.Unstructured{Object: map[string]interface{}{}}
	assert.Nil(t, NodeAffinityTypesFromAnnotation(empty, false))

	// A malformed annotation must not panic or force affinity onto the pod.
	malformed := &unstructured.Unstructured{Object: map[string]interface{}{}}
	malformed.SetName("test-project")
	malformed.SetAnnotations(map[string]string{commonconstants.NodeAffinityTrainAnnotation: "not-json"})
	assert.Nil(t, NodeAffinityTypesFromAnnotation(malformed, false))
}
