/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	outils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func DeploymentForRunaiConfig(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runaiv1.RunaiConfig) (client.Object, error) {
	var (
		deploymentName      = "runai-admission-controller"
		deploymentNamespace = runaiConfig.GetNamespace()

		global   = runaiConfig.Spec.Global
		config   = runaiConfig.Spec.AdmissionController
		image    = config.Image
		replicas = config.Replicas
	)

	deployment := &appsv1.Deployment{}

	err := k8sReader.Get(ctx, types.NamespacedName{
		Name:      deploymentName,
		Namespace: deploymentNamespace,
	}, deployment)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	deployment.SetName(deploymentName)
	deployment.SetNamespace(deploymentNamespace)

	if deployment.GetLabels() == nil {
		deployment.SetLabels(map[string]string{})
	}

	deployment.Labels["app"] = deploymentName

	deployment.Spec.Replicas = replicas
	deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": deploymentName,
		},
	}

	deployment.Spec.Template.Name = deploymentName
	deployment.Spec.Template.Labels = outils.InitMapIfNil(deployment.Spec.Template.Labels)
	deployment.Spec.Template.Labels["app"] = deploymentName

	deployment.Spec.Template.Spec = v1.PodSpec{
		Affinity: outils.MergeAffinities(config.Affinity, runaiConfig.Spec.Global.Affinity,
			deployment.Spec.Selector.MatchLabels, runaiConfig.Spec.Global.RequireDefaultPodAntiAffinity),
		Tolerations:        outils.GetTolerations(runaiConfig),
		ServiceAccountName: deploymentName,
		Containers: []v1.Container{
			{
				Name:            deploymentName,
				SecurityContext: outils.GlobalSecurityContext(runaiConfig),
				Image:           fmt.Sprintf("%s/%s:%s", *image.Registry, *image.Name, *image.Tag),
				ImagePullPolicy: image.PullPolicy,
				VolumeMounts: append(outils.GlobalVolumeMounts(runaiConfig),
					v1.VolumeMount{
						Name:      "webhook-certs",
						MountPath: "/etc/webhook/certs",
						ReadOnly:  true,
					},
				),
				Env: []v1.EnvVar{
					{
						Name:  "WEBHOOK_SERVER_READ_TIMEOUT",
						Value: strconv.Itoa(*config.Server.ReadTimeout),
					},
					{
						Name:  "WEBHOOK_SERVER_WRITE_TIMEOUT",
						Value: strconv.Itoa(*config.Server.WriteTimeout),
					},
					{
						Name:  "GODEBUG",
						Value: outils.FipsGodebugValue(runaiConfig, "off"),
					},
				},
				Args: func() []string {
					args := []string{}

					if global.KubevirtCluster.Enabled {
						args = append(args, "--runai-kubevirt-enabled-cluster")
					}

					if global.ScheduleWorkflowPods.Enabled != nil && *global.ScheduleWorkflowPods.Enabled {
						args = append(args, "--runai-schedule-workflow-pods")
					}

					if runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor == internal.KubernetesFlavorGKE {
						args = append(args, "--specific-cluster-type", "gke")
					}

					if runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor == internal.KubernetesFlavorOKE {
						args = append(args, "--specific-cluster-type", "oke")
					}
					return args
				}(),
				Resources: v1.ResourceRequirements(*config.Resources),
			},
		},
		ImagePullSecrets: outils.GetGlobalImagePullSecrets(runaiConfig),
		Volumes: append(outils.GlobalVolumes(runaiConfig),
			v1.Volume{
				Name: "webhook-certs",
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: "runai-tls-secret",
					},
				},
			},
		),
	}

	return deployment, nil
}

func DeploymentStatus(ctx context.Context,
	k8sReader client.Reader, runaiConfig *runaiv1.RunaiConfig) (operands.Status, error) {
	resource, err := DeploymentForRunaiConfig(ctx, k8sReader, runaiConfig)
	if err != nil {
		return operands.NotReadyStatus(), err
	}

	deployment := resource.(*appsv1.Deployment)

	if deployment.GetUID() == "" {
		return operands.NotReadyStatus(fmt.Sprintf("deployment [%s] is missing", deployment.GetName())), nil
	}

	deplopymentStatus := deployment.Status
	if len(deplopymentStatus.Conditions) == 0 {
		return operands.NotReadyStatus(fmt.Sprintf("deployment [%s] has no conditions",
			deployment.GetName())), nil
	}

	conditions := deplopymentStatus.Conditions
	for _, condition := range conditions {
		if condition.Type == appsv1.DeploymentAvailable &&
			condition.Status == v1.ConditionStatus(metav1.ConditionTrue) {
			return operands.ReadyStatus(), nil
		}
	}

	return operands.NotReadyStatus(fmt.Sprintf("deployment [%s] has not reached the Available condition",
		deployment.GetName())), nil
}
