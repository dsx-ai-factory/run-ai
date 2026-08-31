/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	outils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func DaemonSetForRunaiConfig(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runaiv1.RunaiConfig) (client.Object, error) {
	var (
		name      = "runai-mps-server"
		namespace = runaiConfig.GetNamespace()

		config = runaiConfig.Spec.MPSServer
		image  = config.Image
	)

	daemonSet := &appsv1.DaemonSet{}

	// Get the existing deployment if it exists to consume any cluster-set values
	err := k8sReader.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, daemonSet)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	daemonSet.SetName(name)
	daemonSet.SetNamespace(namespace)

	daemonSet.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": name,
		},
	}

	daemonSet.Spec.Template.Name = name
	daemonSet.Spec.Template.Labels = outils.InitMapIfNil(daemonSet.Spec.Template.Labels)
	daemonSet.Spec.Template.Labels["app"] = name

	daemonSet.Spec.Template.Spec = v1.PodSpec{
		ServiceAccountName: "runai-mps",
		Affinity:           outils.GPUNodeAffinity(runaiConfig, nil, nil, false),
		Tolerations: []v1.Toleration{
			{
				Key:      "CriticalAddonsOnly",
				Operator: v1.TolerationOpExists,
			},
			{
				Key:      "nvidia.com/gpu",
				Operator: v1.TolerationOpExists,
				Effect:   v1.TaintEffectNoSchedule,
			},
			{
				Key:      "kubernetes.azure.com/scalesetpriority",
				Operator: v1.TolerationOpEqual,
				Value:    "spot",
				Effect:   v1.TaintEffectNoSchedule,
			},
		},
		PriorityClassName: "runai-critical",
		HostIPC:           true,
		Containers: []v1.Container{
			{
				Name:    "mps-server-dev",
				Image:   fmt.Sprintf("%s/%s:%s", *image.Registry, *image.Name, *image.Tag),
				Command: []string{"nvidia-cuda-mps-control"},
				Args:    []string{"-f"},
				Env: []v1.EnvVar{
					{
						Name:  "NVIDIA_VISIBLE_DEVICES",
						Value: "all",
					},
					{
						Name:  "NVIDIA_DRIVER_CAPABILITIES",
						Value: "all",
					},
					{
						Name:  "CUDA_MPS_PIPE_DIRECTORY",
						Value: "/runai/mps/",
					},
					{
						Name:  "GODEBUG",
						Value: outils.FipsGodebugValue(runaiConfig, "off"),
					},
				},
				VolumeMounts: append(outils.GlobalVolumeMounts(runaiConfig),
					v1.VolumeMount{
						Name:      "runai-mps",
						MountPath: "/runai/mps/",
					},
				),
			},
		},
		Volumes: append(outils.GlobalVolumes(runaiConfig),
			v1.Volume{
				Name: "runai-mps",
				VolumeSource: v1.VolumeSource{
					HostPath: &v1.HostPathVolumeSource{
						Path: config.RunAIMPSDirectoryHostPath,
						Type: func() *v1.HostPathType {
							doc := v1.HostPathDirectoryOrCreate
							return &doc
						}(),
					},
				},
			},
		),
	}

	if runaiConfig.Spec.Global.DaemonsetsTolerations != nil {
		daemonSet.Spec.Template.Spec.Tolerations = append(daemonSet.Spec.Template.Spec.Tolerations, runaiConfig.Spec.Global.DaemonsetsTolerations...)
	}

	return daemonSet, nil
}

func DaemonSetStatus(ctx context.Context,
	k8sReader client.Reader, runaiConfig *runaiv1.RunaiConfig) (operands.Status, error) {
	resource, err := DaemonSetForRunaiConfig(ctx, k8sReader, runaiConfig)
	if err != nil {
		return operands.ReadyStatus(), err
	}

	daemonSet := resource.(*appsv1.DaemonSet)

	if daemonSet.GetUID() == "" {
		return operands.NotReadyStatus(
			fmt.Sprintf("daemonset [%s] is missing", daemonSet.GetName()),
		), nil
	}

	status := daemonSet.Status

	if status.DesiredNumberScheduled != status.NumberAvailable {
		return operands.NotReadyStatus(
			fmt.Sprintf("daemonset [%s] is not fully rolled-out",
				daemonSet.GetName())), nil
	}

	return operands.ReadyStatus(), nil
}
