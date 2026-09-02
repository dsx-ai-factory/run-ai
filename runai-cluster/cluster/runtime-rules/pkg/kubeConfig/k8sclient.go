/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package kubeConfig

import (
	"context"
	"fmt"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetWLR(ctx context.Context, k8sClient client.Client, namespace string, workloadRunnerName string) (dromosv1alpha1.WorkloadRunner, error) {
	workloadRunner := dromosv1alpha1.WorkloadRunner{}
	err := k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: workloadRunnerName}, &workloadRunner)
	if err != nil {
		zap.S().Errorf("err %+v ", err)
		return workloadRunner, err
	}
	return workloadRunner, nil
}

func ListWLR(ctx context.Context, k8sClient client.Client, namespace string) (dromosv1alpha1.WorkloadRunnerList, error) {
	workloadRunnerList := dromosv1alpha1.WorkloadRunnerList{}
	err := k8sClient.List(ctx, &workloadRunnerList, client.InNamespace(namespace))
	if err != nil {
		zap.S().Errorf("err %+v ", err)
		return workloadRunnerList, err
	}
	return workloadRunnerList, nil
}

// ListProjects lists the kai.resources projects.
func ListProjects(ctx context.Context, k8sClient client.Client) (kaiorg.ProjectList, error) {
	projectList := kaiorg.ProjectList{}
	err := k8sClient.List(ctx, &projectList)
	if err != nil {
		zap.S().Errorf("err %+v ", err)
		return kaiorg.ProjectList{}, err
	}
	return projectList, nil
}

func ProjectNameOfNamespace(ctx context.Context, k8sClient client.Client, namespace string, workloadLabels map[string]string) (string, error) {
	if workloadLabels == nil {
		zap.S().Warnf("workloadLabels is nil, getting workload project using namespace %s", namespace)
	} else {
		workloadProject := workloadLabels[constants.RunaiProjectLabel]
		if workloadProject == "" {
			zap.S().Warnf("runai-project label not found on workload, getting workload project using namespace %s", namespace)
		} else {
			return workloadProject, nil
		}
	}

	var ns corev1.Namespace
	err := k8sClient.Get(ctx, client.ObjectKey{Name: namespace}, &ns)
	if err != nil {
		return "", err
	}

	if ns.Labels[constants.RunaiQueueLabel] == "" {
		return "", fmt.Errorf("runai-queue label not found on namespace %s", namespace)
	}

	return ns.Labels[constants.RunaiQueueLabel], nil
}
