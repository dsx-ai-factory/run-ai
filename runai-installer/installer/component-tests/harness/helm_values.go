/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package harness

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
)

// Canned upgrade.HelmValuesReader functions for the common "what would helm
// report for the existing release" shapes the upgrade tests need.
//
// Each returns the minimum subset of values required by the upgrade flow:
// global.domain on CP, cluster.uid on the cluster release, plus whatever the
// specific scenario adds (customCA, distribution, etc.).

func HelmValues(release, _ string) (map[string]interface{}, error) {
	switch release {
	case step.CPReleaseName:
		return map[string]interface{}{
			"global": map[string]interface{}{
				"domain": "test.example.com",
			},
		}, nil
	case step.ClusterReleaseName:
		return map[string]interface{}{
			"cluster": map[string]interface{}{
				"uid": "uuid-abc",
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown release %s", release)
	}
}

func HelmValuesOpenShift(release, _ string) (map[string]interface{}, error) {
	switch release {
	case step.CPReleaseName:
		return map[string]interface{}{
			"global": map[string]interface{}{
				"domain": "test.example.com",
				"config": map[string]interface{}{
					"kubernetesDistribution": "openshift",
				},
			},
		}, nil
	case step.ClusterReleaseName:
		return map[string]interface{}{
			"cluster": map[string]interface{}{
				"uid": "uuid-abc",
			},
			"global": map[string]interface{}{
				"config": map[string]interface{}{
					"kubernetesDistribution": "openshift",
				},
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown release %s", release)
	}
}

func HelmValuesWithCustomCA(release, _ string) (map[string]interface{}, error) {
	switch release {
	case step.CPReleaseName:
		return map[string]interface{}{
			"global": map[string]interface{}{
				"domain": "test.example.com",
				"customCA": map[string]interface{}{
					"enabled": true,
					"secret": map[string]interface{}{
						"name": "runai-ca-cert",
						"key":  "runai-ca.pem",
					},
				},
			},
		}, nil
	case step.ClusterReleaseName:
		return map[string]interface{}{
			"cluster": map[string]interface{}{
				"uid": "uuid-abc",
			},
			"global": map[string]interface{}{
				"customCA": map[string]interface{}{
					"enabled": true,
					"secret": map[string]interface{}{
						"name": "runai-ca-cert",
						"key":  "runai-ca.pem",
					},
				},
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown release %s", release)
	}
}
