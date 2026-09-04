/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

type RunaiPhaseCalculator struct {
}

func (r *RunaiPhaseCalculator) Calculate(pod *v1.Pod) IRunaiPodPhaseExtended {
	// can't calculate state, keeping init values
	if pod == nil {
		klog.Error(fmt.Errorf("pod is nil"), "failed to calculate runai phase")
		return &Extended{runaiPhase: RunaiPhaseUnknown, runaiPhaseMessage: EmptyStr}
	}

	phaseExtended := NewRunaiPodPhaseExtended(pod)

	switch {
	case checkSuccess(phaseExtended):
		klog.V(4).Info("pod is in success phase", "ns", pod.Namespace, "name", pod.Name)
		phaseExtended.runaiPhase = RunaiPhaseSuccess
		phaseExtended.runaiPhaseMessage = pod.Status.Reason
	case checkFailed(phaseExtended):
		klog.V(4).Info("pod is in failed phase", "ns", pod.Namespace, "name", pod.Name)
		phaseExtended.runaiPhase = RunaiPhaseFailed
		phaseExtended.runaiPhaseMessage = pod.Status.Reason
	case checkPodCreated(phaseExtended):
		klog.V(4).Info("pod is in pending phase", "ns", pod.Namespace, "name", pod.Name)
		phaseExtended.runaiPhase = RunaiPhasePending
		phaseExtended.runaiPhaseMessage = pod.Status.Reason
	case checkPodScheduled(phaseExtended):
		klog.V(6).Info("pod scheduled", "ns", pod.Namespace, "name", pod.Name)
		klog.V(4).Info("pod is in initializing phase", "ns", pod.Namespace, "name", pod.Name)
		phaseExtended.runaiPhase = RunaiPhaseInitializing
		phaseExtended.runaiPhaseMessage = pod.Status.Reason
	case checkRunning(phaseExtended):
		updatePodRunningPhase(phaseExtended)
		klog.V(4).Info("pod running phase", "runaiPhase", GetPhaseStr(phaseExtended.GetPhase()), "ns", pod.Namespace, "name", pod.Name)
	default:
		klog.V(2).Info("unknown phase", "phase", pod.Status.Phase, "ns", pod.Namespace, "name", pod.Name)
	}
	return phaseExtended
}

func checkSuccess(pe *Extended) bool {
	return pe.pod.Status.Phase == v1.PodSucceeded
}

func checkFailed(pe *Extended) bool {
	return pe.pod.Status.Phase == v1.PodFailed
}

func checkPodCreated(pe *Extended) bool {
	pe.populateConditionMap()
	if pe.pod.Status.Phase != v1.PodPending {
		return false
	}

	if condition, found := pe.conditionMap[v1.PodScheduled]; found {
		if condition.Status == v1.ConditionFalse {
			pe.runaiPhaseMessage = condition.Reason
			return true
		}
	} else {
		return true
	}
	return false
}

// checkPodScheduled check if the pod status is pending but scheduled
func checkPodScheduled(pe *Extended) bool {
	if pe.pod.Status.Phase == v1.PodPending {
		if condition, found := pe.conditionMap[v1.PodScheduled]; found {
			if condition.Status == v1.ConditionTrue {
				pe.runaiPhaseMessage = condition.Reason
				return true
			}
		}
	}
	return false
}

func checkRunning(pe *Extended) bool {
	return pe.pod.Status.Phase == v1.PodRunning
}

func updatePodRunningPhase(pe *Extended) {
	var found bool
	var condition *v1.PodCondition

	if condition, found = pe.conditionMap[v1.PodReady]; found == true {
		if condition.Status == v1.ConditionTrue {
			pe.runaiPhase = RunaiPhaseReady
		} else {
			pe.runaiPhase = RunaiPhaseInitializing
			pe.runaiPhaseMessage = condition.Reason
		}
	} else if condition, found = pe.conditionMap[v1.PodScheduled]; found == true {
		pe.runaiPhase = RunaiPhaseInitializing
		pe.runaiPhaseMessage = condition.Reason
	} else if condition, found = pe.conditionMap[v1.ContainersReady]; found == true {
		pe.runaiPhase = RunaiPhaseInitializing
		pe.runaiPhaseMessage = condition.Reason
	} else {
		pe.runaiPhase = RunaiPhasePending
	}
}
