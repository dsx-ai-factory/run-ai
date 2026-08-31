/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"fmt"
	"sort"
	"strings"

	runaiPodPhase "github.com/run-ai/runai/runai-cluster/cluster/common/utils/pod"
	"github.com/run-ai/runai/runai-cluster/cluster/runaijob-controller/pkg/common/runaiJobUtils"
	runajob "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/apis/core/validation"
)

var phase2ConditionMap = map[runaiPodPhase.RunaiPodPhase]runajob.JobConditionType{
	runaiPodPhase.RunaiPhasePending:      runajob.JobPending,
	runaiPodPhase.RunaiPhaseInitializing: runajob.JobInitializing,
	runaiPodPhase.RunaiPhaseReady:        runajob.JobRunning,
}

func GenerateExtendedPods(pods []*v1.Pod) *runaiPodPhase.ExtendedList {
	var result runaiPodPhase.ExtendedList
	runaiPhaseHandler := runaiPodPhase.NewRunaiPodPhaseHandler(&runaiPodPhase.RunaiPhaseCalculator{})
	for _, p := range pods {
		runaiPhase := runaiPhaseHandler.Calculate(p).(*runaiPodPhase.Extended)
		if runaiPhase.GetPhase() != runaiPodPhase.RunaiPhaseUnknown {
			result = append(result, *runaiPhase)
		} else {
			klog.V(4).Infof("Ignoring inactive pod %v/%v in state %v, deletion time %v",
				p.Namespace, p.Name, p.Status.Phase, p.DeletionTimestamp)
		}
	}
	if result != nil {
		sort.Sort(result)
	}
	return &result
}

func ShouldUpdateRunaiJobStatus(runaiJob *runajob.RunaiJob, runaiPods *runaiPodPhase.ExtendedList) bool {
	var shouldUpdate bool
	conditionMap := generateConditionMap(runaiPods)
	shouldUpdate = jobCheckAndUpdateCondition(&runaiJob.Status.Conditions, conditionMap)
	shouldUpdate = jobCheckAndUpdateStatus(&runaiJob.Status, runaiPods) || shouldUpdate

	return shouldUpdate
}

func ClearPodsStatus(runaiJob *runajob.RunaiJob) {
	runaiJob.Status.Pending = 0
	runaiJob.Status.Initialized = 0
	runaiJob.Status.Running = 0
}

func generateConditionMap(runaiPods *runaiPodPhase.ExtendedList) *runaiJobUtils.JobConditionMap {
	var (
		conditionMap     = make(runaiJobUtils.JobConditionMap)
		found            bool
		jobConditionType runajob.JobConditionType
		condition        runajob.JobCondition
	)

	for _, pod := range *runaiPods {
		jobConditionType, found = phase2ConditionMap[pod.GetPhase()]
		if !found {
			continue
		}

		condition, found = conditionMap[jobConditionType]
		if !found {
			conditionMap[jobConditionType] = runaiJobUtils.NewCondition(jobConditionType, v1.ConditionTrue, runaiPodPhase.EmptyStr, runaiPodPhase.EmptyStr)
		}

		condition, found = conditionMap[jobConditionType]
		if len(pod.GetPhaseMessage()) != 0 {
			if condition.Message == runaiPodPhase.EmptyStr {
				condition.Message = fmt.Sprintf("%s:%s", pod.GetPod().Name, pod.GetPhaseMessage())
				conditionMap[jobConditionType] = condition
			} else {
				condition.Message += fmt.Sprintf(",%s:%s", pod.GetPod().Name, pod.GetPhaseMessage())
				conditionMap[jobConditionType] = condition
			}
		}
	}
	return &conditionMap
}

func jobCheckAndUpdateCondition(conditions *[]runajob.JobCondition, conditionMap *runaiJobUtils.JobConditionMap) bool {
	// if condition is empty either update with condition, or keep empty.
	if conditions == nil {
		// no news is good news.
		if len(*conditionMap) == 0 {
			return false
		}

		*conditions = make([]runajob.JobCondition, 0)
		for _, val := range *conditionMap {
			*conditions = append(*conditions, val)
		}
		return true
	}

	var shouldUpdate = false
	jobConditionsMap := runaiJobUtils.GenerateConditionMap(conditions)

	// check if generated condition match the current job condition
	for _, podCondition := range runaiJobUtils.PodConditions {
		// fetch job current condition
		jobCurrentCondition, jobCurrentFound := jobConditionsMap[podCondition]
		// fetch generated condition
		generatedCondition, generatedFound := (*conditionMap)[podCondition]

		if jobCurrentFound && generatedFound { // if we have a match, check message value
			if strings.Compare(jobCurrentCondition.Message, generatedCondition.Message) != 0 ||
				strings.Compare(jobCurrentCondition.Reason, generatedCondition.Reason) != 0 ||
				jobCurrentCondition.Status != generatedCondition.Status {
				runaiJobUtils.EnsureJobConditionStatus(*conditions, podCondition, generatedCondition.Status, generatedCondition.Reason, generatedCondition.Message)
				shouldUpdate = true
			}
		} else if jobCurrentFound && !generatedFound { // if job contain but generated doesn't, set status to false and clear the message
			runaiJobUtils.EnsureJobConditionStatus(*conditions, podCondition, v1.ConditionFalse, runaiPodPhase.EmptyStr, runaiPodPhase.EmptyStr)
			shouldUpdate = true
		} else if !jobCurrentFound && generatedFound { // if job doesn't contain condition (add it)
			*conditions = append(*conditions, runaiJobUtils.NewCondition(generatedCondition.Type, generatedCondition.Status, generatedCondition.Reason, generatedCondition.Message))
			shouldUpdate = true
		} else { // if not found in job and generated condition, no need for a change.
			klog.V(4).Infof("job condition %s is unchanged", podCondition)
		}
	}
	return shouldUpdate
}

func jobCheckAndUpdateStatus(jobStatus *runajob.JobStatus, pods *runaiPodPhase.ExtendedList) bool {
	var (
		pending, initialized, running, ready int32
	)

	for _, podExtended := range *pods {
		switch podExtended.GetPhase() {
		case runaiPodPhase.RunaiPhasePending:
			pending++
		case runaiPodPhase.RunaiPhaseInitializing:
			initialized++
		case runaiPodPhase.RunaiPhaseReady:
			ready++
		}
		if podExtended.IsPodActive() {
			running++
		}

	}

	if pending != jobStatus.Pending || initialized != jobStatus.Initialized ||
		running != jobStatus.Running || ready != jobStatus.Ready {
		jobStatus.Pending = pending
		jobStatus.Initialized = initialized
		jobStatus.Running = running
		jobStatus.Ready = ready
		return true
	}
	return false
}

func GetPodsLabelSet(template *v1.PodTemplateSpec) labels.Set {
	desiredLabels := make(labels.Set)
	for k, v := range template.Labels {
		desiredLabels[k] = v
	}
	return desiredLabels
}

func GetPodsFinalizers(template *v1.PodTemplateSpec) []string {
	desiredFinalizers := make([]string, len(template.Finalizers))
	copy(desiredFinalizers, template.Finalizers)
	return desiredFinalizers
}

func GetPodsAnnotationSet(template *v1.PodTemplateSpec) labels.Set {
	desiredAnnotations := make(labels.Set)
	for k, v := range template.Annotations {
		desiredAnnotations[k] = v
	}
	return desiredAnnotations
}

func GetPodsPrefix(prefix string) string {
	// use the dash (if the name isn't too long) to make the pod name a bit prettier
	prefixWDash := fmt.Sprintf("%s-", prefix)
	if len(validation.ValidatePodName(prefixWDash, true)) != 0 {
		prefixWDash = prefix
	}
	return prefixWDash
}
