/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package topology_test

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/topology"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/stretchr/testify/require"
)

func TestActiveAnnotationValue(t *testing.T) {
	const annotationName = "example.com/placement"

	tests := []struct {
		name          string
		annotations   []workloadsApi.Annotation
		expectedValue string
		expectedFound bool
	}{
		{
			name: "last active value wins",
			annotations: []workloadsApi.Annotation{
				newAnnotation(annotationName, "first", false),
				newAnnotation("example.com/unrelated", "ignored", false),
				newAnnotation(annotationName, "last", false),
			},
			expectedValue: "last",
			expectedFound: true,
		},
		{
			name: "final exclusion removes value",
			annotations: []workloadsApi.Annotation{
				newAnnotation(annotationName, "value", false),
				newAnnotation(annotationName, "", true),
			},
			expectedValue: "",
			expectedFound: false,
		},
		{
			name: "active value after exclusion reactivates annotation",
			annotations: []workloadsApi.Annotation{
				newAnnotation(annotationName, "initial", false),
				newAnnotation(annotationName, "", true),
				newAnnotation(annotationName, "reactivated", false),
			},
			expectedValue: "reactivated",
			expectedFound: true,
		},
		{
			name: "active empty value remains present",
			annotations: []workloadsApi.Annotation{
				newAnnotation(annotationName, "", false),
			},
			expectedValue: "",
			expectedFound: true,
		},
		{
			name: "absent annotation is not found",
			annotations: []workloadsApi.Annotation{
				newAnnotation("example.com/unrelated", "value", false),
			},
			expectedValue: "",
			expectedFound: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, found := topology.ActiveAnnotationValue(test.annotations, annotationName)

			require.Equal(t, test.expectedValue, value)
			require.Equal(t, test.expectedFound, found)
		})
	}
}

func newAnnotation(name, value string, exclude bool) workloadsApi.Annotation {
	annotation := workloadsApi.NewAnnotation()
	annotation.SetName(name)
	annotation.SetValue(value)
	annotation.SetExclude(exclude)
	return *annotation
}
