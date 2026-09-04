/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"fmt"

	workloadsManagerApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	clusterconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// applyNetworkingSubmissionProfile moves the server-owned profile from the
// typed control-plane update onto the WorkloadRunner. A same-name annotation on
// the attached workload is always removed so workload metadata cannot spoof the
// trusted marker.
//
// A missing profile remains valid during rolling upgrades: an older control
// plane can send endpoint networking to a newer agent. Once rollout capability
// gating is in place, newly enabled direct-v2 networking can require the marker
// without breaking existing generated-native and generated-NIM flows.
func applyNetworkingSubmissionProfile(
	manifest metav1.Object,
	runner metav1.Object,
	configuration *workloadsManagerApi.WorkloadV2UpdateConfiguration,
) error {
	manifestAnnotations := manifest.GetAnnotations()
	if manifestAnnotations != nil {
		delete(manifestAnnotations, clusterconstants.NetworkingSubmissionProfileAnnotation)
		manifest.SetAnnotations(manifestAnnotations)
	}

	if configuration == nil {
		return nil
	}

	profileValue, hasProfile := configuration.GetNetworkingSubmissionProfileOk()
	hasNetworkingEndpoints := hasNetworkingEndpointsConfiguration(configuration)
	if !hasProfile {
		return nil
	}
	if !hasNetworkingEndpoints {
		return fmt.Errorf("networking submission profile requires endpoint-array networking")
	}

	if !profileValue.IsValid() {
		return fmt.Errorf("unknown networking submission profile %q", *profileValue)
	}
	profile := string(*profileValue)

	runnerAnnotations := runner.GetAnnotations()
	if runnerAnnotations == nil {
		runnerAnnotations = make(map[string]string)
	}
	runnerAnnotations[clusterconstants.NetworkingSubmissionProfileAnnotation] = profile
	runner.SetAnnotations(runnerAnnotations)
	return nil
}
