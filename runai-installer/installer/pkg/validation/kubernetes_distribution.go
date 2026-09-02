/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validation

const (
	KubernetesDistributionOpenShift = "openshift"
)

var ValidKubernetesDistributions = []string{KubernetesDistributionOpenShift}

func IsValidKubernetesDistribution(distribution string) bool {
	if distribution == "" {
		return true
	}
	for _, valid := range ValidKubernetesDistributions {
		if distribution == valid {
			return true
		}
	}
	return false
}
