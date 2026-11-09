/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gpu_pci_inventory

import (
	"context"
	"fmt"

	outils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type daemonSetResource struct {
	reader client.Reader
	config *runv1.RunaiConfig
}

func (r *daemonSetResource) OperandName() string          { return name }
func (r *daemonSetResource) Enabled() bool                { return true }
func (r *daemonSetResource) Status(context.Context) error { return nil }
func (r *daemonSetResource) Object(ctx context.Context) (client.Object, error) {
	ds := &appsv1.DaemonSet{}
	key := client.ObjectKey{Namespace: r.config.Namespace, Name: name}
	if err := r.reader.Get(ctx, key, ds); err != nil && !apierrors.IsNotFound(err) {
		return nil, fmt.Errorf("get GPU PCI inventory daemonset: %w", err)
	}
	ds.Name = name
	ds.Namespace = r.config.Namespace
	ds.Spec.Selector = &metav1.LabelSelector{MatchLabels: map[string]string{"app": name}}
	ds.Spec.Template.Labels = outils.InitMapIfNil(ds.Spec.Template.Labels)
	ds.Spec.Template.Labels["app"] = name
	image := r.config.Spec.ClusterSync.Image
	tolerations := []corev1.Toleration{
		{Key: "CriticalAddonsOnly", Operator: corev1.TolerationOpExists},
		{Key: "nvidia.com/gpu", Operator: corev1.TolerationOpExists, Effect: corev1.TaintEffectNoSchedule},
		{Key: "kubernetes.azure.com/scalesetpriority", Operator: corev1.TolerationOpEqual, Value: "spot", Effect: corev1.TaintEffectNoSchedule},
	}
	tolerations = append(tolerations, outils.GetTolerations(r.config)...)
	tolerations = append(tolerations, r.config.Spec.Global.DaemonsetsTolerations...)
	ds.Spec.Template.Spec = corev1.PodSpec{
		ServiceAccountName: name,
		NodeSelector:       map[string]string{"kubernetes.io/os": "linux"},
		Affinity:           outils.GPUNodeAffinity(r.config, nil, nil, false),
		Tolerations:        tolerations,
		ImagePullSecrets:   outils.GetGlobalImagePullSecrets(r.config),
		Containers: []corev1.Container{{
			Name: name, Image: fmt.Sprintf("%s/%s:%s", *image.Registry, *image.Name, *image.Tag), ImagePullPolicy: image.PullPolicy,
			Env: []corev1.EnvVar{
				{Name: "GPU_PCI_COLLECTOR", Value: "true"},
				{Name: "NODE_NAME", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.nodeName"}}},
				{Name: "NVIDIA_VISIBLE_DEVICES", Value: "void"},
				{Name: "GODEBUG", Value: outils.FipsGodebugValue(r.config, "off")},
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("10m"), corev1.ResourceMemory: resource.MustParse("32Mi")},
				Limits:   corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("100m"), corev1.ResourceMemory: resource.MustParse("128Mi")},
			},
			SecurityContext: &corev1.SecurityContext{
				RunAsNonRoot: ptr.To(true), RunAsUser: ptr.To[int64](65532), RunAsGroup: ptr.To[int64](65532),
				Privileged: ptr.To(false), AllowPrivilegeEscalation: ptr.To(false), ReadOnlyRootFilesystem: ptr.To(true),
				Capabilities:   &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}},
				SeccompProfile: &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault},
			},
			VolumeMounts: []corev1.VolumeMount{
				{Name: "sys", MountPath: "/host/sys", ReadOnly: true},
				{Name: "nvidia-gpus", MountPath: "/host/nvidia-gpus", ReadOnly: true},
			},
		}},
		Volumes: []corev1.Volume{
			{Name: "sys", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/sys", Type: ptr.To(corev1.HostPathDirectory)}}},
			{Name: "nvidia-gpus", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/proc/driver/nvidia/gpus", Type: ptr.To(corev1.HostPathDirectory)}}},
		},
	}
	return ds, nil
}

type serviceAccountResource struct {
	reader client.Reader
	config *runv1.RunaiConfig
}

func (r *serviceAccountResource) OperandName() string          { return name }
func (r *serviceAccountResource) Enabled() bool                { return true }
func (r *serviceAccountResource) Status(context.Context) error { return nil }
func (r *serviceAccountResource) Object(ctx context.Context) (client.Object, error) {
	sa := &corev1.ServiceAccount{}
	key := client.ObjectKey{Namespace: r.config.Namespace, Name: name}
	if err := r.reader.Get(ctx, key, sa); err != nil && !apierrors.IsNotFound(err) {
		return nil, fmt.Errorf("get GPU PCI inventory service account: %w", err)
	}
	sa.Name = name
	sa.Namespace = r.config.Namespace
	return sa, nil
}
