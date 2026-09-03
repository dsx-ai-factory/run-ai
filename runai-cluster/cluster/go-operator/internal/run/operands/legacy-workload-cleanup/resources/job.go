/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"

	legacycleanup "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/legacy-workload-cleanup"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const operandName = "legacy-workload-cleanup"

type JobResource struct {
	reader client.Reader
}

func NewJobResource(reader client.Reader) *JobResource {
	return &JobResource{reader: reader}
}

func (r *JobResource) OperandName() string { return operandName }

func (r *JobResource) Enabled() bool { return true }

func (r *JobResource) Status(context.Context) error { return nil }

func (r *JobResource) Object(ctx context.Context) (client.Object, error) {
	key := client.ObjectKey{Namespace: runaiv1.DefaultNamespace, Name: legacycleanup.CleanupJobName}
	existing := &batchv1.Job{}
	if err := r.reader.Get(ctx, key, existing); err == nil {
		return existing.DeepCopy(), nil
	} else if !apierrors.IsNotFound(err) {
		return nil, fmt.Errorf("get cleanup Job %s: %w", key, err)
	}

	operator := &appsv1.Deployment{}
	operatorKey := client.ObjectKey{Namespace: runaiv1.DefaultNamespace, Name: "runai-operator"}
	if err := r.reader.Get(ctx, operatorKey, operator); err != nil {
		return nil, fmt.Errorf("get go-operator Deployment %s: %w", operatorKey, err)
	}
	var source *corev1.Container
	for i := range operator.Spec.Template.Spec.Containers {
		container := &operator.Spec.Template.Spec.Containers[i]
		if container.Name == "runai-go-operator" {
			source = container
			break
		}
	}
	if source == nil {
		return nil, fmt.Errorf("go-operator Deployment %s has no runai-go-operator container", operatorKey)
	}

	container := source.DeepCopy()
	container.Args = []string{"--legacy-workload-cleanup"}
	container.Ports = nil
	container.LivenessProbe = nil
	container.ReadinessProbe = nil
	container.StartupProbe = nil
	container.Lifecycle = nil

	podSpec := operator.Spec.Template.Spec.DeepCopy()
	podSpec.ServiceAccountName = legacycleanup.CleanupJobName
	podSpec.RestartPolicy = corev1.RestartPolicyNever
	podSpec.InitContainers = nil
	podSpec.Containers = []corev1.Container{*container}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: runaiv1.DefaultNamespace,
			Name:      legacycleanup.CleanupJobName,
			Labels:    map[string]string{"app": legacycleanup.CleanupJobName},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: ptr.To[int32](3),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": legacycleanup.CleanupJobName}},
				Spec:       *podSpec,
			},
		},
	}, nil
}
