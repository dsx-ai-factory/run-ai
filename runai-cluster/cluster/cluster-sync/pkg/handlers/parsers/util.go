/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	clustercommonconsts "github.com/run-ai/runai/runai-cluster/cluster/common/constants"

	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/constants"
	runaiJobV1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetOwner(ownerReference []metav1.OwnerReference) *metav1.OwnerReference {
	if len(ownerReference) == 0 {
		return nil
	}
	owner := ownerReference[0]
	for _, potentialOwner := range ownerReference {
		if potentialOwner.Kind == clustercommonconsts.KindRunaiJob {
			owner = potentialOwner
		}
	}
	return &owner
}

func GetImageName(pod *corev1.Pod) string {
	if len(pod.Spec.Containers) == 0 {
		return ""
	}
	return pod.Spec.Containers[0].Image
}

//nolint:unused
func _convertTimeToMilli(time *metav1.Time) int64 {
	var startTime int64
	if time != nil {
		startTime = time.UnixMilli()
	}
	return startTime
}

func parseBoolLabel(labels map[string]string, label string) (bool, error) {
	if value, err := strconv.ParseBool(labels[label]); err == nil {
		return value, nil
	} else {
		return false, err
	}
}

//nolint:unused
func _parseStringAnnotationWithDefault(annotations map[string]string, annotation string, defaultVal string) string {
	if value, ok := annotations[annotation]; ok {
		return value
	}
	return defaultVal
}

//nolint:unused
func _getImageNameFromPods(pods []*corev1.Pod) string {
	if len(pods) == 0 {
		return ""
	} else {
		return GetImageName(pods[0])
	}
}

func GetProjectNameFromLabelsOrNamespace(cachedReader client.Reader, labels map[string]string, namespaceName string) (string, error) {
	projectNameFromLabel, found := labels[constants.RunaiProjectNameLabel]
	if found && projectNameFromLabel != "" {
		return projectNameFromLabel, nil
	}

	return GetProjectNameFromNamespace(cachedReader, namespaceName)
}

func GetProjectNameFromNamespace(cachedReader client.Reader, namespaceName string) (string, error) {
	namespace := &corev1.Namespace{}
	err := cachedReader.Get(context.TODO(), types.NamespacedName{Name: namespaceName}, namespace)
	if err != nil {
		return "", err
	}
	errNotFound := fmt.Errorf("project name was not found in namespace %s", namespaceName)

	if value, ok := namespace.Labels[constants.RunaiQueueLabel]; ok {
		return value, nil
	} else {
		return "", errNotFound
	}
}

func GetUser(podGroup *kaiv2alpha2.PodGroup) string {
	user := ""
	if userAnnotation, found := podGroup.Annotations["user"]; found && userAnnotation != "" {
		user = userAnnotation
	} else if userLabel, found := podGroup.Labels["user"]; found {
		user = userLabel
	}
	return user
}

func parseStringAnnotation(annotations map[string]string, backupAnnotations map[string]string, annotation string, objectName string) string {
	if value, ok := annotations[annotation]; ok {
		return value
	} else {
		zap.S().Debugf("Failed to parse annotation for %s annotation: %s", objectName, annotation)

		if value, ok := backupAnnotations[annotation]; ok {
			zap.S().Infof("Found Annotation in backup, fallback on it")
			return value
		}
		return ""
	}
}

func parseFloatAnnotation(annotations map[string]string, backupAnnotations map[string]string, annotation string, objectName string) float64 {
	if value, err := strconv.ParseFloat(annotations[annotation], 64); err == nil {
		return value
	} else {
		zap.S().Debugf("Failed to parse annotation for %s annotation: %s", objectName, annotation)
		if value, err := strconv.ParseFloat(backupAnnotations[annotation], 64); err == nil {
			zap.S().Infof("Found Annotation in backup, fallback on it")
			return value
		}
		return 0.0
	}
}

func GetParallelism(runaiJob *runaiJobV1.RunaiJob) int32 {
	var parallelism int32 = constants.ParallelismDefault
	if runaiJob != nil && runaiJob.Spec.Parallelism != nil {
		parallelism = *runaiJob.Spec.Parallelism
	}
	return parallelism
}

func GetCompletions(runaiJob *runaiJobV1.RunaiJob) int32 {
	var completions int32 = constants.CompletionsDefault
	if runaiJob != nil && runaiJob.Spec.Completions != nil {
		completions = *runaiJob.Spec.Completions
	}
	return completions
}

func GetTotalLimitCpuMemory(pods []*corev1.Pod) float64 {
	var totalLimitCpuMemory float64
	for _, pod := range pods {
		for _, container := range pod.Spec.Containers {
			totalLimitCpuMemory += container.Resources.Limits.Memory().AsApproximateFloat64()
		}
	}
	return totalLimitCpuMemory
}

func GetParentWorkloadName(labels map[string]string) string {
	var workloadName = ""
	if labels != nil {
		workloadName = labels["workloadName"]
	}
	return workloadName
}

func GetTotalRequestedCpuMemory(pods []*corev1.Pod) float64 {
	var totalRequestedCpuMemory float64
	for _, pod := range pods {
		for _, container := range pod.Spec.Containers {
			totalRequestedCpuMemory += container.Resources.Requests.Memory().AsApproximateFloat64()
		}
	}
	return totalRequestedCpuMemory
}

func GetTotalLimitCPU(pods []*corev1.Pod) float64 {
	var totalLimitCPU float64
	for _, pod := range pods {
		for _, container := range pod.Spec.Containers {
			totalLimitCPU += container.Resources.Limits.Cpu().AsApproximateFloat64()
		}
	}
	return totalLimitCPU
}

func GetTotalRequestedCPU(pods []*corev1.Pod) float64 {
	var totalRequestedCPU float64
	for _, pod := range pods {
		for _, container := range pod.Spec.Containers {
			totalRequestedCPU += container.Resources.Requests.Cpu().AsApproximateFloat64()
		}
	}
	return totalRequestedCPU
}

func GetTotalRequestedMigDevices(pods []*corev1.Pod) string {
	var totalRequestedMigArray []string
	pgRequestedMigDevices := map[string]int{}
	for _, pod := range pods {
		for resourceFullName, resourceCount := range pod.Annotations {
			migDeviceName := ConvertMigResourceFullNameToMigInstance(resourceFullName)
			// Ignore annotations that are not relevant to mig
			if migDeviceName == "" || migDeviceName == resourceFullName {
				continue
			}
			if _, ok := pgRequestedMigDevices[migDeviceName]; !ok {
				pgRequestedMigDevices[migDeviceName] = 0
			}
			num, _ := strconv.Atoi(resourceCount)
			pgRequestedMigDevices[migDeviceName] += num
		}
	}
	for migDeviceName, count := range pgRequestedMigDevices {
		for i := 0; i < count; i++ {
			totalRequestedMigArray = append(totalRequestedMigArray, migDeviceName)
		}
	}
	sort.Strings(totalRequestedMigArray)
	return strings.Join(totalRequestedMigArray, ",")
}

func ConvertMigResourceFullNameToMigInstance(migResourceFullName string) string {
	return strings.TrimPrefix(migResourceFullName, constants.MigResourcePrefix)
}
