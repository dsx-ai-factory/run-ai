/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package patch

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestAddNodeAffinityToPodSpec(t *testing.T) {
	podSpec := &v1.PodSpec{}
	matchExpressionToAdd := v1.NodeSelectorRequirement{
		Key:      "run.ai/type",
		Operator: v1.NodeSelectorOpIn,
		Values:   []string{"val"},
	}
	AddNodeAffinityToPodSpec(podSpec, []v1.NodeSelectorRequirement{matchExpressionToAdd})

	terms := podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	assert.Equal(t, 1, len(terms))
	assert.Equal(t, 1, len(terms[0].MatchExpressions))
	assert.Equal(t, matchExpressionToAdd.Key, terms[0].MatchExpressions[0].Key)
	assert.Equal(t, matchExpressionToAdd.Values, terms[0].MatchExpressions[0].Values)
}

func TestAddMultipleNodeAffinitiesToPodSpec(t *testing.T) {
	podSpec := &v1.PodSpec{}
	matchExpressionToAdd1 := v1.NodeSelectorRequirement{
		Key:      "key1",
		Operator: v1.NodeSelectorOpIn,
		Values:   []string{"val1"},
	}
	matchExpressionToAdd2 := v1.NodeSelectorRequirement{
		Key:      "key2",
		Operator: v1.NodeSelectorOpIn,
		Values:   []string{"val2"},
	}
	AddNodeAffinityToPodSpec(podSpec, []v1.NodeSelectorRequirement{matchExpressionToAdd1})
	AddNodeAffinityToPodSpec(podSpec, []v1.NodeSelectorRequirement{matchExpressionToAdd2})

	terms := podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	assert.Equal(t, 1, len(terms))
	assert.Equal(t, 2, len(terms[0].MatchExpressions))
	key1Found := false
	key2Found := false
	for _, matchExpression := range terms[0].MatchExpressions {
		switch matchExpression.Key {
		case matchExpressionToAdd1.Key:
			assert.Equal(t, matchExpressionToAdd1.Values, matchExpression.Values)
			key1Found = true
		case matchExpressionToAdd2.Key:
			assert.Equal(t, matchExpressionToAdd2.Values, matchExpression.Values)
			key2Found = true
		default:
			assert.Fail(t, "unexpected matchExpression")
		}
	}
	assert.Equal(t, true, key1Found)
	assert.Equal(t, true, key2Found)
}

func TestAddNodeAffinityToExistingAffinityInPodSpec(t *testing.T) {
	podSpec := &v1.PodSpec{
		Affinity: &v1.Affinity{
			NodeAffinity: &v1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
					NodeSelectorTerms: []v1.NodeSelectorTerm{
						{
							MatchExpressions: []v1.NodeSelectorRequirement{
								{
									Key:    "key1",
									Values: []string{"val1, val2"},
								},
							},
						},
						{
							MatchExpressions: []v1.NodeSelectorRequirement{
								{
									Key:    "key2",
									Values: []string{"val3, val4"},
								},
							},
						},
					},
				},
			},
		},
	}
	matchExpressionToAdd := v1.NodeSelectorRequirement{
		Key:      "run.ai/type",
		Operator: v1.NodeSelectorOpIn,
		Values:   []string{"val"},
	}
	AddNodeAffinityToPodSpec(podSpec, []v1.NodeSelectorRequirement{matchExpressionToAdd})

	terms := podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	assert.Equal(t, 2, len(terms))
	assert.Equal(t, 2, len(terms[0].MatchExpressions))
	key1Found := false
	key2Found := false
	runaiKey1Found := false
	runaiKey2Found := false
	for _, term := range terms {
		for _, matchExpression := range term.MatchExpressions {
			switch matchExpression.Key {
			case "key1":
				assert.Equal(t, []string{"val1, val2"}, matchExpression.Values)
				key1Found = true
			case "key2":
				assert.Equal(t, []string{"val3, val4"}, matchExpression.Values)
				key2Found = true
			case "run.ai/type":
				assert.Equal(t, []string{"val"}, matchExpression.Values)
				if !runaiKey1Found {
					runaiKey1Found = true
				} else {
					runaiKey2Found = true
				}
			default:
				assert.Fail(t, "unexpected matchExpression")
			}
		}
	}
	assert.Equal(t, true, key1Found)
	assert.Equal(t, true, key2Found)
	assert.Equal(t, true, runaiKey1Found)
	assert.Equal(t, true, runaiKey2Found)
}

func TestOverrideNodeAffinity(t *testing.T) {
	podSpec := &v1.PodSpec{
		Affinity: &v1.Affinity{
			NodeAffinity: &v1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
					NodeSelectorTerms: []v1.NodeSelectorTerm{
						{
							MatchExpressions: []v1.NodeSelectorRequirement{
								{
									Key:    "key1",
									Values: []string{"val1, val2"},
								},
								{
									Key:    "key2",
									Values: []string{"val3, val4"},
								},
							},
						},
					},
				},
			},
		},
	}
	matchExpressionToAdd := v1.NodeSelectorRequirement{
		Key:      "key1",
		Operator: v1.NodeSelectorOpIn,
		Values:   []string{"val5, val6"},
	}
	AddNodeAffinityToPodSpec(podSpec, []v1.NodeSelectorRequirement{matchExpressionToAdd})

	terms := podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	assert.Equal(t, 1, len(terms))
	assert.Equal(t, 2, len(terms[0].MatchExpressions))
	key1Found := false
	key2Found := false
	for _, matchExpression := range terms[0].MatchExpressions {
		switch matchExpression.Key {
		case "key1":
			assert.Equal(t, []string{"val5, val6"}, matchExpression.Values)
			key1Found = true
		case "key2":
			assert.Equal(t, []string{"val3, val4"}, matchExpression.Values)
			key2Found = true
		default:
			assert.Fail(t, "unexpected matchExpression")
		}
	}
	assert.Equal(t, true, key1Found)
	assert.Equal(t, true, key2Found)
}

func TestAddSingleToleration(t *testing.T) {
	podSpec := &v1.PodSpec{}
	AddTolerationToSpec(podSpec, "key", "value", v1.TolerationOpExists, v1.TaintEffectNoSchedule)

	assert.Equal(t, len(podSpec.Tolerations), 1)
	assert.Equal(t, podSpec.Tolerations[0].Key, "key")
	assert.Equal(t, podSpec.Tolerations[0].Value, "value")
	assert.Equal(t, podSpec.Tolerations[0].Operator, v1.TolerationOpExists)
	assert.Equal(t, podSpec.Tolerations[0].Effect, v1.TaintEffectNoSchedule)
}

func TestAddMultipleTolerations(t *testing.T) {
	podSpec := &v1.PodSpec{}
	AddTolerationToSpec(podSpec, "key1", "value1", v1.TolerationOpExists, v1.TaintEffectNoSchedule)
	AddTolerationToSpec(podSpec, "key2", "value2", v1.TolerationOpEqual, v1.TaintEffectNoSchedule)

	assert.Equal(t, len(podSpec.Tolerations), 2)
	for _, toleration := range podSpec.Tolerations {
		switch toleration.Key {
		case "key1":
			assert.Equal(t, toleration.Value, "value1")
			assert.Equal(t, toleration.Operator, v1.TolerationOpExists)
			assert.Equal(t, toleration.Effect, v1.TaintEffectNoSchedule)
		case "key2":
			assert.Equal(t, toleration.Value, "value2")
			assert.Equal(t, toleration.Operator, v1.TolerationOpEqual)
			assert.Equal(t, toleration.Effect, v1.TaintEffectNoSchedule)
		default:
			assert.Fail(t, "unexpected toleration")
		}
	}
}

func TestGPUSelectionIntersectsPlacementOnEveryAdmission(t *testing.T) {
	model := v1.NodeSelectorRequirement{Key: "nvidia.com/gpu.product", Operator: v1.NodeSelectorOpIn, Values: []string{"A100"}}
	conflict := v1.NodeSelectorRequirement{Key: model.Key, Operator: v1.NodeSelectorOpIn, Values: []string{"H100"}}
	zone := v1.NodeSelectorRequirement{Key: "zone", Operator: v1.NodeSelectorOpIn, Values: []string{"west"}}
	name := v1.NodeSelectorRequirement{Key: "metadata.name", Operator: v1.NodeSelectorOpIn, Values: []string{"worker-1"}}
	selector := &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{{MatchExpressions: []v1.NodeSelectorRequirement{model}, MatchFields: []v1.NodeSelectorRequirement{name}}, {}}}
	spec := &v1.PodSpec{Affinity: &v1.Affinity{NodeAffinity: &v1.NodeAffinity{RequiredDuringSchedulingIgnoredDuringExecution: selector}}}
	selection := map[string]string{constants.AnnotationGPUTypeLabel: model.Key + "=A100"}
	alternatives := []v1.NodeSelectorRequirement{conflict, zone}

	AddPlacementRequirements(spec, selection, alternatives)
	assert.Equal(t, []v1.NodeSelectorTerm{
		{MatchExpressions: []v1.NodeSelectorRequirement{model, conflict}, MatchFields: []v1.NodeSelectorRequirement{name}},
		{MatchExpressions: []v1.NodeSelectorRequirement{model, zone}, MatchFields: []v1.NodeSelectorRequirement{name}},
		{},
	}, selector.NodeSelectorTerms)
	once := spec.DeepCopy()
	AddPlacementRequirements(spec, selection, alternatives)
	assert.Equal(t, once, spec)

	selector.NodeSelectorTerms = []v1.NodeSelectorTerm{}
	AddPlacementRequirements(spec, selection, alternatives)
	assert.Empty(t, selector.NodeSelectorTerms)
}

func TestKnativeResourceOnlyAffinityUsesStrictPlacement(t *testing.T) {
	key := constants.AcceleratorResourceNodeLabel("amd.com/gpu")
	requested := v1.NodeSelectorRequirement{Key: key, Operator: v1.NodeSelectorOpIn, Values: []string{"true"}}
	conflict := v1.NodeSelectorRequirement{Key: key, Operator: v1.NodeSelectorOpDoesNotExist}
	required := &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{{MatchExpressions: []v1.NodeSelectorRequirement{requested}}}}
	spec := &v1.PodSpec{Affinity: &v1.Affinity{NodeAffinity: &v1.NodeAffinity{RequiredDuringSchedulingIgnoredDuringExecution: required}}}
	AddPlacementRequirements(spec, nil, []v1.NodeSelectorRequirement{conflict})
	assert.Equal(t, []v1.NodeSelectorRequirement{requested, conflict}, required.NodeSelectorTerms[0].MatchExpressions)
}
