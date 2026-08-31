/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod

import (
	v1 "k8s.io/api/core/v1"
)

type IRunaiPhaseCalculator interface {
	Calculate(pod *v1.Pod) IRunaiPodPhaseExtended
}

type IRunaiPodPhaseExtended interface {
	GetPod() *v1.Pod
	GetPhase() RunaiPodPhase
	GetPhaseMessage() string
	IsPodActive() bool
}

type RunaiPodPhaseHandler struct {
	runaiPhaseCalculator IRunaiPhaseCalculator
}

func NewRunaiPodPhaseHandler(runaiPhaseCalculator IRunaiPhaseCalculator) RunaiPodPhaseHandler {
	return RunaiPodPhaseHandler{runaiPhaseCalculator: runaiPhaseCalculator}
}

func (ph *RunaiPodPhaseHandler) Calculate(pod *v1.Pod) IRunaiPodPhaseExtended {
	return ph.runaiPhaseCalculator.Calculate(pod)
}
