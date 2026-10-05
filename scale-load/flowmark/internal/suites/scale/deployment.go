/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package scale

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/measure"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/submit"
)

type deploymentFlow struct {
	kube        k8sclient.WithWatch
	concurrency int
}

func (*deploymentFlow) name() string {
	return "deployment_to_external_workload"
}

func (f *deploymentFlow) watch(ctx context.Context, namespace string) (*measure.Watcher, error) {
	return measure.WatchExternalWorkloads(ctx, f.kube, namespace)
}

func (f *deploymentFlow) create(ctx context.Context, workload item) (string, error) {
	deployment, err := submit.K8sDeployment(ctx, f.kube, submit.DeploymentSpec{
		Name: workload.name, Namespace: workload.namespace, Image: workload.image, Replicas: workload.pods, Labels: workload.labels,
	})
	if err != nil {
		return "", err
	}
	return string(deployment.UID), nil
}

func (f *deploymentFlow) teardown(ctx context.Context, namespace string, labels map[string]string) (int, error) {
	var deployments appsv1.DeploymentList
	if err := f.kube.List(ctx, &deployments, k8sclient.InNamespace(namespace), k8sclient.MatchingLabels(labels)); err != nil {
		return 0, fmt.Errorf("listing the round's deployments: %w", err)
	}
	targets := make([]k8sclient.Object, 0, len(deployments.Items))
	for i := range deployments.Items {
		if deployments.Items[i].DeletionTimestamp == nil {
			targets = append(targets, &deployments.Items[i])
		}
	}
	if err := deleteAll(ctx, f.concurrency, targets, func(ctx context.Context, obj k8sclient.Object) error {
		return k8sclient.IgnoreNotFound(f.kube.Delete(ctx, obj))
	}); err != nil {
		return len(deployments.Items), err
	}
	external, err := countAllInNamespace(ctx, f.kube, measure.ExternalWorkloadKind, namespace)
	if err != nil {
		return len(deployments.Items), err
	}
	return len(deployments.Items) + external, nil
}
