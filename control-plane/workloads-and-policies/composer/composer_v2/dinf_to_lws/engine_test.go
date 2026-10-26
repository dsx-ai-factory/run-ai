/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dinf_to_lws

import (
	"context"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/stretchr/testify/require"
)

func TestIsNIMLeader(t *testing.T) {
	tests := []struct {
		name     string
		image    string
		labels   []workloadsApi.Label
		expected bool
	}{
		{
			name:     "canonical label identifies custom image",
			image:    "registry.example.com/nim:latest",
			labels:   []workloadsApi.Label{newNIMLabel(constants.LabelNIMWorkload, "true", false)},
			expected: true,
		},
		{
			name:     "legacy label identifies custom image",
			image:    "registry.example.com/nim:latest",
			labels:   []workloadsApi.Label{newNIMLabel(legacyNIMWorkloadLabel, "true", false)},
			expected: true,
		},
		{
			name:     "canonical image is detected",
			image:    "nvcr.io/nim/meta/llama3:latest",
			expected: true,
		},
		{
			name:     "regular inference NIM repository image is detected",
			image:    "nvcr.io/nvidia/nim:latest",
			expected: true,
		},
		{
			name:     "explicit false suppresses image detection",
			image:    "nvcr.io/nim/meta/llama3:latest",
			labels:   []workloadsApi.Label{newNIMLabel(constants.LabelNIMWorkload, "false", false)},
			expected: false,
		},
		{
			name:     "excluded label is ignored",
			image:    "registry.example.com/nim:latest",
			labels:   []workloadsApi.Label{newNIMLabel(constants.LabelNIMWorkload, "true", true)},
			expected: false,
		},
		{
			name:  "custom workload is not detected",
			image: "registry.example.com/server:latest",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			leader := workloadsApi.NewDistributedInferenceLeaderWorkerSpec()
			leader.SetImage(test.image)
			leader.SetLabels(test.labels)

			require.Equal(t, test.expected, isNIMLeader(leader))
		})
	}
}

func newNIMLabel(name, value string, exclude bool) workloadsApi.Label {
	label := workloadsApi.NewLabel()
	label.SetName(name)
	label.SetValue(value)
	label.SetExclude(exclude)
	return *label
}

func TestComposePropagatesNIMIdentityForDistributedInference(t *testing.T) {
	spec := newDistributedInferenceSpec("registry.example.com/nim-cpu:latest")
	leader := spec.Leader.Get()
	leader.SetLabels([]workloadsApi.Label{newNIMLabel(constants.LabelNIMWorkload, "true", false)})

	target := composeDistributedInference(t, spec)

	require.Equal(t, constants.BoolTrueAsString, target.LeaderWorkerSet.Labels[constants.LabelNIMWorkload])
	require.Equal(
		t,
		constants.BoolTrueAsString,
		target.LeaderWorkerSet.Spec.LeaderWorkerTemplate.LeaderTemplate.Labels[constants.LabelNIMWorkload],
	)
	require.Len(t, target.Config.Networking, 1)
	serviceSpec := target.Config.Networking[0].Services[0].GetSpec()
	serviceLabels := serviceSpec.GetLabels()
	require.Equal(t, constants.BoolTrueAsString, serviceLabels[constants.LabelNIMWorkload])
	require.Equal(t, constants.KindDistributedInferenceWorkload, serviceLabels[constants.LabelWorkloadKind])
	require.Equal(t, "dist-inf", serviceLabels[constants.LabelWorkloadName])
}

func TestComposeInfersNIMIdentityFromCanonicalImage(t *testing.T) {
	target := composeDistributedInference(t, newDistributedInferenceSpec("nvcr.io/nim/meta/llama3:latest"))

	require.Equal(t, constants.BoolTrueAsString, target.LeaderWorkerSet.Labels[constants.LabelNIMWorkload])
	serviceSpec := target.Config.Networking[0].Services[0].GetSpec()
	serviceLabels := serviceSpec.GetLabels()
	require.Equal(t, constants.BoolTrueAsString, serviceLabels[constants.LabelNIMWorkload])
}

func newDistributedInferenceSpec(image string) *workloadsApi.DistributedInferenceSpecSpec {
	spec := workloadsApi.NewDistributedInferenceSpecSpec()
	leader := workloadsApi.NewDistributedInferenceLeaderWorkerSpec()
	leader.SetImage(image)
	spec.SetLeader(*leader)
	servingPort := workloadsApi.NewDistributedInferenceServingPort()
	servingPort.SetPort(8000)
	servingPort.SetExposeExternally(false)
	spec.SetServingPort(*servingPort)
	return spec
}

func composeDistributedInference(
	t *testing.T,
	spec *workloadsApi.DistributedInferenceSpecSpec,
) *LeaderWorkerSetTarget {
	dinfComposer := NewDistributedInferenceToLWSComposer(
		"dist-inf", "project-a", "namespace-a", "user-a",
		nil, nil, composer.FieldsComposerOptions{}, nil,
	)
	target := NewMinimalLeaderWorkerSetTarget()

	require.Nil(t, dinfComposer.Compose(context.Background(), spec, target)())
	return target
}
