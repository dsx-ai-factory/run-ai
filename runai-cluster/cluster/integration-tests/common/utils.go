/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"context"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetPodByPrefix fetches a pod that matches the given prefix in the specified namespace
func GetPodByPrefix(ctx context.Context, kubeClient client.Client, podPrefix, namespace string) (*corev1.Pod, error) {
	podList := &corev1.PodList{}
	err := kubeClient.List(ctx, podList, client.InNamespace(namespace))
	if err != nil {
		return nil, fmt.Errorf("failed to list pods in namespace '%s': %w", namespace, err)
	}

	for _, pod := range podList.Items {
		if strings.HasPrefix(pod.ObjectMeta.Name, podPrefix) {
			return &pod, nil
		}
	}
	return nil, nil
}

func ValidatePodStatus(ctx context.Context, kubeClient client.Client, podPrefix, namespace string, expectedStatus corev1.PodPhase) {
	err := WaitForCondition(ctx, 1*time.Second, 10*time.Second, func() bool {
		createdPod, _ := GetPodByPrefix(ctx, kubeClient, podPrefix, namespace)
		return createdPod != nil && createdPod.Status.Phase == expectedStatus
	})
	Expect(err).ToNot(HaveOccurred(),
		fmt.Sprintf("Expected one pod with a name starting with '%s' in namespace '%s' to have '%s' status.",
			podPrefix, namespace, expectedStatus))
}

func ValidatePodNotFound(ctx context.Context, kubeClient client.Client, podPrefix, namespace string) {
	err := WaitForCondition(ctx, 1*time.Second, 10*time.Second, func() bool {
		createdPod, _ := GetPodByPrefix(ctx, kubeClient, podPrefix, namespace)
		return createdPod == nil
	})
	Expect(err).ToNot(HaveOccurred(),
		fmt.Sprintf("Expected one pod with a name starting with '%s' in namespace '%s' to be deleted.",
			podPrefix, namespace))
}

// WaitForCondition repeatedly checks a condition function at a specified interval
// until either the condition is met or the timeout is reached.
// Parameters:
// - ctx: The context to control the execution and cancellation of the waiting process.
// - interval: The time between each condition check. The function waits this long before polling again.
// - timeout: The maximum time to wait before giving up and returning a timeout error.
// - conditionFunc: A function that returns a boolean indicating whether the condition has been met.
// The function will return nil if the condition is met within the timeout period,
// or an error if the timeout is reached.
func WaitForCondition(ctx context.Context, interval time.Duration, timeout time.Duration, conditionFunc func() bool) error {
	timeoutCh := time.After(timeout)

	for {
		if conditionFunc() {
			return nil // Condition met
		}

		select {
		case <-time.After(interval):
			// Wait for the next poll
		case <-timeoutCh:
			return fmt.Errorf("timeout reached")
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func GetPVC(ctx context.Context, kubeClient client.Client, namespace, pvcName string) (*corev1.PersistentVolumeClaim, error) {
	pvc := &corev1.PersistentVolumeClaim{}
	// Use kubeClient.Get to retrieve the PVC
	err := kubeClient.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      pvcName,
	}, pvc)

	if err != nil {
		return nil, fmt.Errorf("failed to get PVC %s/%s: %w", namespace, pvcName, err)
	}
	return pvc, nil
}

// GetRandomName creates a unique name using a prefix and a random 4-character string.
func GetRandomName(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, rand.String(4))
}

func CreatePodSpec(containerName string, priorityClassName string, cpuRequest, memRequest, gpuRequest, imageName string) corev1.PodSpec {
	podSpec := corev1.PodSpec{
		SchedulerName: "runai-scheduler",
		Containers: []corev1.Container{
			{
				Name:    containerName,
				Image:   imageName,
				Command: []string{"bash", "-c", "sleep infinity"},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{},
				},
			},
		},
	}

	if priorityClassName != "" {
		podSpec.PriorityClassName = priorityClassName
	}

	if cpuRequest != "" {
		podSpec.Containers[0].Resources.Requests[corev1.ResourceCPU] = resource.MustParse(cpuRequest)
	}

	if memRequest != "" {
		podSpec.Containers[0].Resources.Requests[corev1.ResourceMemory] = resource.MustParse(memRequest)
	}

	if gpuRequest != "" {
		podSpec.Containers[0].Resources.Requests["nvidia.com/gpu"] = resource.MustParse(gpuRequest)
	}

	return podSpec
}
