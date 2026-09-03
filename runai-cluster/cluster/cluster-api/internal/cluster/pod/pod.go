/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/api"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

const (
	ErrPodNotReady = "pod is not ready"
)

func ValidatePodRunning(pod *v1.Pod) error {
	if pod.Status.Phase != v1.PodRunning {
		log.Warn().Str("pod", pod.Name).Str("phase", string(pod.Status.Phase)).Msg("Pod is not in running phase")
		return api.NewNotReadyError(ErrPodNotReady)
	}

	return nil
}

func ValidatePodNotPending(pod *v1.Pod) error {
	log.Debug().Str("podName", pod.Name).Msg("Starting pod validation")

	if pod.Status.Phase == v1.PodPending {
		log.Warn().Str("podName", pod.Name).Msg(ErrPodNotReady)
		return api.NewNotReadyError(ErrPodNotReady)
	}

	log.Debug().Str("podName", pod.Name).Msg("Pod validation successful")

	return nil
}
