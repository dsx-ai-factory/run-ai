/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	"reflect"
	"testing"

	clusterapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	api "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/component-helpers/scheduling/corev1/nodeaffinity"
)

func TestGPUSelectionChoosesResourceAndModelInMixedPool(t *testing.T) {
	const amd = "amd.com/gpu"
	const amdLabel = "hardware.example.com/model"
	amdNode := corev1.Node{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{amdLabel: "MI300X"}}}
	h100 := corev1.Node{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{NvidiaProductLabel: "H100"}}}
	a100 := corev1.Node{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{NvidiaProductLabel: "A100"}}}
	acc := &api.AcceleratorSelection{}
	acc.SetResourceName(amd)
	amdModel := *acc
	amdModel.SetDeviceLabelKey(amdLabel)
	for _, tc := range []struct {
		name     string
		model    *string
		acc      *api.AcceleratorSelection
		resource string
		matches  []bool
	}{
		{"nvidia-default", convutils.Ref("H100"), nil, NvidiaResourceName, []bool{false, true, false}},
		{"amd-resource-only-has-no-model-affinity", nil, acc, amd, []bool{true, true, true}},
		{"amd-model", convutils.Ref("MI300X"), &amdModel, amd, []bool{true, false, false}},
		{"amd-absent", convutils.Ref("MI325X"), &amdModel, amd, []bool{false, false, false}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fc := FieldsComposer{Options: FieldsComposerOptions{NodeAffinitySupported: tc.model != nil}}
			var pod corev1.Pod
			if err := fc.ComposeGPUPlacement(tc.model, tc.acc, nil, &pod.Spec, &pod.ObjectMeta)(); err != nil {
				t.Fatal(err.Message)
			}
			snapshot := pod.DeepCopy()
			if err := fc.ComposeGPUPlacement(tc.model, tc.acc, nil, &pod.Spec, &pod.ObjectMeta)(); err != nil || !reflect.DeepEqual(snapshot, &pod) {
				t.Fatal("placement changed on repeat", err)
			}
			var resources corev1.ResourceRequirements
			if err := WholeGpusToK8sResources(&GpuSizingInfo{NumDevices: convutils.Ref(int32(2))}, &resources, tc.acc); err != nil {
				t.Fatal(err.Message)
			}
			for _, list := range []corev1.ResourceList{resources.Requests, resources.Limits} {
				q := list[corev1.ResourceName(tc.resource)]
				if len(list) != 1 || q.Value() != 2 {
					t.Fatalf("unexpected resources %v", list)
				}
			}
			if len(pod.Spec.NodeSelector) != 0 {
				t.Fatalf("GPU placement added a nodeSelector: %v", pod.Spec.NodeSelector)
			}
			for i, node := range []corev1.Node{amdNode, h100, a100} {
				node.Name = "gpu-node"
				match, err := nodeaffinity.GetRequiredNodeAffinity(&pod).Match(&node)
				if err != nil || match != tc.matches[i] {
					t.Errorf("labels %v match=%v err=%v", node.Labels, match, err)
				}
			}
		})
	}
}

func TestGPUSelectionIntersectsPoolAndEmptySelectors(t *testing.T) {
	require := func(key, value string) corev1.NodeSelectorRequirement {
		return corev1.NodeSelectorRequirement{Key: key, Operator: corev1.NodeSelectorOpIn, Values: []string{value}}
	}
	fc := FieldsComposer{Options: FieldsComposerOptions{NodeAffinitySupported: true}}
	for _, terms := range [][]corev1.NodeSelectorTerm{nil, {{}}, {{MatchExpressions: []corev1.NodeSelectorRequirement{require("rack", "north")}}, {MatchExpressions: []corev1.NodeSelectorRequirement{require("rack", "south")}}}} {
		pod := corev1.Pod{Spec: corev1.PodSpec{Affinity: &corev1.Affinity{NodeAffinity: &corev1.NodeAffinity{RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{NodeSelectorTerms: terms}}}}}
		if err := fc.ComposeGPUPlacement(convutils.Ref("A100"), nil, []string{"h100-pool"}, &pod.Spec, &pod.ObjectMeta)(); err != nil {
			t.Fatal(err.Message)
		}
		intersectNodePools(pod.Spec.Affinity.NodeAffinity, []corev1.NodeSelectorRequirement{require(NvidiaProductLabel, "H100")})
		for _, model := range []string{"A100", "H100"} {
			for _, rack := range []string{"north", "south", "other"} {
				node := corev1.Node{ObjectMeta: metav1.ObjectMeta{Labels: labels.Set{"rack": rack, NvidiaProductLabel: model}}}
				match, err := nodeaffinity.GetRequiredNodeAffinity(&pod).Match(&node)
				if err != nil || match {
					t.Fatalf("conflict or empty selector matched %v, err=%v", node.Labels, err)
				}
			}
		}
	}
}

func TestGPUSelectionRejectsMalformedNamesAndUnknownKeys(t *testing.T) {
	for _, tc := range []struct{ resource, key, model string }{
		{"cpu", "", ""}, {"kubernetes.io/fake", "", ""}, {"requests.amd.com/gpu", "", ""}, {"amd.com/gpu", "", "MI300X"}, {"amd.com/gpu", "invalid key", "MI300X"}, {NvidiaResourceName, "", "bad.model/"},
	} {
		acc := &api.AcceleratorSelection{}
		acc.SetResourceName(tc.resource)
		if tc.key != "" {
			acc.SetDeviceLabelKey(tc.key)
		}
		var model *string
		if tc.model != "" {
			model = &tc.model
		}
		if _, _, err := ResolveGPUSelection(model, acc); err == nil {
			t.Errorf("accepted invalid selection %+v", tc)
		}
	}
}

func TestGPUSelectionRejectsOnlyDeletingPools(t *testing.T) {
	pool := clusterapi.Nodepool{}
	pool.SetName("retiring")
	pool.SetPhase(clusterapi.NODEPOOLPHASE_DELETING)
	input := &NodePoolsInfo{RequestedNodePools: []string{"retiring"}, ActualNodePools: []clusterapi.Nodepool{pool}, PreserveRequirements: true}
	var affinity *corev1.NodeAffinity
	if err := NodePoolsToK8s(input, &affinity, nil); err == nil || err.Code != 400 {
		t.Fatalf("expected a bad request when every explicitly selected pool is being deleted, got %v", err)
	}
}
