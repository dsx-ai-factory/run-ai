/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sampler_actions

import (
	"context"
	"fmt"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	dwv2alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	twv2alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ApplyUnstructured(ctx context.Context, dynamicClient dynamic.Interface, gvr schema.GroupVersionResource, obj *unstructured.Unstructured) error {
	_, err := dynamicClient.Resource(gvr).Namespace(obj.GetNamespace()).Apply(ctx, obj.GetName(), obj, metav1.ApplyOptions{
		FieldManager: "scale-test",
		Force:        true,
	})
	return err
}

func GetK8sTrainingWorkloadFromCluster(ctx context.Context, namespace string, kubeClient client.Client) (map[string]*twv2alpha1.TrainingWorkload, error) {
	data := make(map[string]*twv2alpha1.TrainingWorkload)
	workloadList := &twv2alpha1.TrainingWorkloadList{}
	listOptions := &client.ListOptions{
		Limit:     200,
		Namespace: namespace,
	}
	for {
		err := kubeClient.List(ctx, workloadList, listOptions)
		if err != nil {
			return nil, err
		}
		for _, workload := range workloadList.Items {
			workloadName := workload.Name
			data[workloadName] = &workload
		}
		if workloadList.Continue == "" {
			break
		}
		listOptions.Continue = workloadList.Continue
	}
	return data, nil
}

func GetK8sDistributedWorkloadFromCluster(ctx context.Context, namespace string, kubeClient client.Client) (map[string]*dwv2alpha1.DistributedWorkload, error) {
	data := make(map[string]*dwv2alpha1.DistributedWorkload)
	workloadList := &dwv2alpha1.DistributedWorkloadList{}
	listOptions := &client.ListOptions{
		Limit:     200,
		Namespace: namespace,
	}
	for {
		err := kubeClient.List(ctx, workloadList, listOptions)
		if err != nil {
			return nil, err
		}
		for _, workload := range workloadList.Items {
			workloadName := workload.Name
			data[workloadName] = &workload
		}
		if workloadList.Continue == "" {
			break
		}
		listOptions.Continue = workloadList.Continue
	}
	return data, nil
}

func GetK8sJobsFromCluster(ctx context.Context, namespace string, kubeClient client.Client) (map[string]*runaiv1.RunaiJob, error) {
	data := make(map[string]*runaiv1.RunaiJob)
	jobList := &runaiv1.RunaiJobList{}
	listOptions := &client.ListOptions{
		Limit:     200,
		Namespace: namespace,
	}
	for {
		err := kubeClient.List(ctx, jobList, listOptions)
		if err != nil {
			return nil, err
		}
		for _, job := range jobList.Items {
			workloadName := job.Labels["workloadName"]
			data[workloadName] = &job
		}
		if jobList.Continue == "" {
			break
		}
		listOptions.Continue = jobList.Continue
	}
	return data, nil
}

func DeleteK8sNameSpaces(namespaces []string, kubeClient client.Client) error {
	for _, namespace := range namespaces {
		err := DeleteK8sNameSpace(namespace, kubeClient)
		if err != nil {
			fmt.Errorf("failed to delete namespace %s: %v", namespace, err)
		}
	}
	return nil
}

func DeleteK8sNameSpace(namespace string, kubeClient client.Client) error {
	k8sNamespace := &corev1.Namespace{}
	k8sNamespace.Name = namespace
	err := kubeClient.Delete(context.Background(), k8sNamespace)
	if err != nil {
		return err
	}
	return nil
}

func GetK8sPodsFromCluster(ctx context.Context, namespace string, kubeClient client.Client) (map[string][]corev1.Pod, error) {
	data := make(map[string][]corev1.Pod)
	pods := &corev1.PodList{}
	listOptions := &client.ListOptions{
		Limit:     200,
		Namespace: namespace,
	}
	for {
		err := kubeClient.List(ctx, pods, listOptions)
		if err != nil {
			return nil, err
		}
		for _, pod := range pods.Items {
			workloadName := pod.Labels["workloadName"]
			podSlice, ok := data[workloadName]
			if !ok {
				podSlice = []corev1.Pod{}
			}
			podSlice = append(podSlice, pod)
			data[workloadName] = podSlice
		}
		if pods.Continue == "" {
			break
		}

		listOptions.Continue = pods.Continue
	}
	fmt.Printf("Extracted: [%d] workloads pod data from cluster", len(data))
	return data, nil
}
