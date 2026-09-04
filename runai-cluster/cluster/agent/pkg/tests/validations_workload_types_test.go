/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"
	"encoding/json"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog/log"
	"reflect"
	"strings"
	"time"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func validateKartaExistsEventually(name string, k8sClient client.Client) {
	var karta *kartav1alpha1.Karta = nil
	EventuallyWithOffset(1, func() *kartav1alpha1.Karta {
		karta = getKartaFromClient(name, k8sClient)
		return karta
	}, DefaultTimeout).ShouldNot(BeNil())
}

func validateKartaDeletedEventually(name string, k8sClient client.Client) {
	EventuallyWithOffset(1, func() *kartav1alpha1.Karta {
		return getKartaFromClient(name, k8sClient)
	}, DefaultTimeout).Should(BeNil())
}

func getKartaFromClient(name string, k8sClient client.Client) *kartav1alpha1.Karta {
	key := types.NamespacedName{Name: name}
	obj := &kartav1alpha1.Karta{}
	if err := k8sClient.Get(context.Background(), key, obj); err != nil {
		return nil
	}
	return obj
}

// buildResourceDefinitionForGvk constructs a minimal Karta resourceDefinition map
// similar to the control-plane helper, for use in agent tests.
func buildResourceDefinitionForGvk(group, version, kind string) map[string]interface{} {
	return map[string]interface{}{
		"spec": map[string]interface{}{
			"structureDefinition": map[string]interface{}{
				"rootComponent": map[string]interface{}{
					"kind": map[string]interface{}{
						"group":   group,
						"version": version,
						"kind":    kind,
					},
					"name": "",
				},
			},
			"optimizationInstructions": map[string]interface{}{},
		},
		"metadata": map[string]interface{}{
			"creationTimestamp": nil,
		},
		"status": map[string]interface{}{},
	}
}

func validateKartaIsUpToDate(expected, actual *kartav1alpha1.Karta) bool {
	expectedSpecJson, err := json.Marshal(expected.Spec)
	if err != nil {
		log.Error().Msgf("Failed to json marshal Karta spec, error: %s", err.Error())
		return false
	}
	actualSpecJson, err := json.Marshal(actual.Spec)
	if err != nil {
		log.Error().Msgf("Failed to json marshal Karta spec, error: %s", err.Error())
		return false
	}

	if !reflect.DeepEqual(expectedSpecJson, actualSpecJson) {
		return false
	}

	if !reflect.DeepEqual(expected.Labels, actual.Labels) {
		return false
	}

	return true
}

func buildExpectedKartaWithLabels(karta *kartav1alpha1.Karta, workloadTypeId *string, updatedAt *time.Time) *kartav1alpha1.Karta {
	expected := karta.DeepCopy()
	kartaGvk := expected.Spec.StructureDefinition.RootComponent.Kind
	Expect(kartaGvk).ToNot(BeNil())
	Expect(workloadTypeId).ToNot(BeNil())

	updatedAtStr := ""
	if updatedAt != nil {
		updatedAtStr = strings.ReplaceAll(updatedAt.Format(time.RFC3339), ":", "-")
	}

	expected.Labels = v1alpha1.GetKartaLabels(*workloadTypeId, kartaGvk.Group, kartaGvk.Version, kartaGvk.Kind, updatedAtStr)
	return expected
}
