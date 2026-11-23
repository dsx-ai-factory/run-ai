/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"strings"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/stretchr/testify/require"
)

func TestQualifiedAffinityKeysInWorkloadAndPodSync(t *testing.T) {
	domain := strings.Repeat("a", 63) + "." + strings.Repeat("b", 63) + "." + strings.Repeat("c", 63) + "." + strings.Repeat("d", 61)
	for _, tc := range []struct {
		name  string
		key   string
		valid bool
	}{
		{"bare name", strings.Repeat("a", 63), true},
		{"managed NVIDIA placement", "run.ai.node-restriction.kubernetes.io/accelerator-nvidia.com_gpu", true},
		{"maximum qualified label", domain + "/" + strings.Repeat("X", 63), true},
		{"name too long", strings.Repeat("n", 64), false},
		{"qualified name too long", "example.org/" + strings.Repeat("n", 64), false},
		{"prefix too long", domain + "x/n", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			affinity := api.PodNodeAffinity{Required: &api.PodNodeAffinityRequired{
				NodeSelectorTerms: []api.NodeSelectorTerm{{MatchExpressions: []api.MatchExpression{
					{Key: tc.key, Operator: api.MATCHEXPRESSIONOPERATOR_IN, Values: []string{"true"}},
				}}},
			}}
			// These nested validators run on both cluster-to-control-plane status endpoints.
			podErr := api.AssertPodNodeAffinityConstraints(affinity)
			workloadErr := api.AssertWorkloadNodeAffinityConstraints(api.WorkloadNodeAffinity{SpecName: "worker", NodeAffinity: affinity})
			for _, err := range []error{podErr, workloadErr} {
				if tc.valid {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
				}
			}
		})
	}
}
