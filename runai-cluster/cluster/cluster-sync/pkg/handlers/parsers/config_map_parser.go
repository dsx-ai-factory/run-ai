/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	k8sobjectstracker "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/util"
	corev1 "k8s.io/api/core/v1"
)

// ParsedConfigMap hold (run-ai) metadata on ConfigMap, such as scope and the scope id
type ParsedConfigMap struct {
	Scope           ScopeTypeAndId
	CreationRequest k8sobjectstracker.ConfigMap
}

// ParseConfigMapToServerScheme makes a parsed ConfigMap from a ConfigMap
func ParseConfigMapToServerScheme(configMap *corev1.ConfigMap) (*ParsedConfigMap, error) {
	result := ParsedConfigMap{
		CreationRequest: k8sobjectstracker.ConfigMap{
			Name:   configMap.Name,
			Issues: util.ExtractIssuesFromObject(configMap.GetAnnotations()),
		},
	}
	err := parseScope(&result.Scope, &configMap.ObjectMeta)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
