/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"testing"

	legacycleanup "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/legacy-workload-cleanup"
	schemeinit "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	clientfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestJobResourceBuildsCleanupJobFromOperatorDeployment(t *testing.T) {
	operator := operatorDeployment()
	reader := clientfake.NewClientBuilder().WithScheme(schemeinit.Scheme()).WithObjects(operator).Build()
	object, err := NewJobResource(reader).Object(context.Background())
	require.NoError(t, err)
	job := object.(*batchv1.Job)

	require.Equal(t, legacycleanup.CleanupJobName, job.Name)
	require.Nil(t, job.Spec.ActiveDeadlineSeconds)
	require.Equal(t, int32(3), *job.Spec.BackoffLimit)
	require.Equal(t, corev1.RestartPolicyNever, job.Spec.Template.Spec.RestartPolicy)
	require.Equal(t, legacycleanup.CleanupJobName, job.Spec.Template.Spec.ServiceAccountName)
	require.Equal(t, []string{"--legacy-workload-cleanup"}, job.Spec.Template.Spec.Containers[0].Args)
	require.Equal(t, operator.Spec.Template.Spec.Containers[0].Image, job.Spec.Template.Spec.Containers[0].Image)
	require.Equal(t, operator.Spec.Template.Spec.Containers[0].Resources, job.Spec.Template.Spec.Containers[0].Resources)
	require.Equal(t, operator.Spec.Template.Spec.ImagePullSecrets, job.Spec.Template.Spec.ImagePullSecrets)
	require.Equal(t, operator.Spec.Template.Spec.Affinity, job.Spec.Template.Spec.Affinity)
	require.Equal(t, operator.Spec.Template.Spec.Tolerations, job.Spec.Template.Spec.Tolerations)
	require.Nil(t, job.Annotations)
}

func TestJobResourceReturnsExistingFailedJobUnchanged(t *testing.T) {
	existing := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{Namespace: "runai", Name: legacycleanup.CleanupJobName, UID: "job-uid"},
		Spec: batchv1.JobSpec{Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers:    []corev1.Container{{Name: "existing", Image: "existing:image"}},
		}}},
		Status: batchv1.JobStatus{Failed: 1},
	}
	reader := clientfake.NewClientBuilder().WithScheme(schemeinit.Scheme()).WithObjects(existing).Build()
	object, err := NewJobResource(reader).Object(context.Background())
	require.NoError(t, err)
	job := object.(*batchv1.Job)
	require.Equal(t, existing.UID, job.UID)
	require.Equal(t, existing.Status, job.Status)
	require.Equal(t, "existing:image", job.Spec.Template.Spec.Containers[0].Image)
}

func operatorDeployment() *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Namespace: "runai", Name: "runai-operator"},
		Spec: appsv1.DeploymentSpec{Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
			ServiceAccountName: "runai-operator",
			ImagePullSecrets:   []corev1.LocalObjectReference{{Name: "registry"}},
			Affinity:           &corev1.Affinity{},
			Tolerations:        []corev1.Toleration{{Key: "gpu"}},
			Volumes:            []corev1.Volume{{Name: "custom-ca"}},
			Containers: []corev1.Container{{
				Name:            "runai-go-operator",
				Image:           "registry/runai-operator:3.0.0",
				ImagePullPolicy: corev1.PullIfNotPresent,
				SecurityContext: &corev1.SecurityContext{RunAsNonRoot: ptr.To(true)},
				Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("10m"),
				}},
				VolumeMounts: []corev1.VolumeMount{{Name: "custom-ca", MountPath: "/ca"}},
			}},
		}}},
	}
}
