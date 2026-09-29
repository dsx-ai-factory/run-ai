/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"
	"strconv"

	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/secrets"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	outils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	name            = operands.RunaiApplicationControllerName
	tlsSecretName   = "application-controller-tls-secret"
	webhookCertName = "webhook-certs"
)

func DeploymentForRunaiConfig(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runaiv1.RunaiConfig) (client.Object, error) {
	var (
		deploymentNamespace = runaiConfig.GetNamespace()

		config = runaiConfig.Spec.ApplicationController
		image  = config.Image
	)
	deployment := &appsv1.Deployment{}

	// Get the existing deployment if it exists to consume any cluster-set values
	err := k8sReader.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: deploymentNamespace,
	}, deployment)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	deployment.Name = name
	deployment.Namespace = deploymentNamespace

	deployment.Labels = outils.InitMapIfNil(deployment.Labels)
	deployment.Labels["app"] = name

	volumes := outils.GlobalVolumes(runaiConfig)
	volumeMounts := outils.GlobalVolumeMounts(runaiConfig)

	volume := v1.Volume{
		Name: webhookCertName,
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName:  tlsSecretName,
				DefaultMode: ptr.To[int32](420),
			},
		},
	}
	volumeMount := v1.VolumeMount{
		Name:      webhookCertName,
		ReadOnly:  true,
		MountPath: "/etc/webhook/certs",
	}
	volumes = append(volumes, volume)
	volumeMounts = append(volumeMounts, volumeMount)

	deployment.Spec.Replicas = pointer.Int32(1)
	deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": name,
		},
	}

	deployment.Spec.Template.Name = name
	deployment.Spec.Template.Labels = outils.InitMapIfNil(deployment.Spec.Template.Labels)
	deployment.Spec.Template.Labels["app"] = name

	deployment.Spec.Template.Spec = v1.PodSpec{
		ServiceAccountName: name,
		// The controller talks to the Kubernetes API, so pin the token mount on even if a cluster
		// baseline defaults automountServiceAccountToken to false.
		AutomountServiceAccountToken: ptr.To(true),
		Affinity: outils.MergeAffinities(config.Affinity, runaiConfig.Spec.Global.Affinity,
			deployment.Spec.Selector.MatchLabels, runaiConfig.Spec.Global.RequireDefaultPodAntiAffinity),
		Tolerations: outils.GetTolerations(runaiConfig),
		Containers: []v1.Container{
			{
				Name:            name,
				SecurityContext: outils.GlobalSecurityContext(runaiConfig),
				Image:           fmt.Sprintf("%s/%s:%s", *image.Registry, *image.Name, *image.Tag),
				ImagePullPolicy: image.PullPolicy,
				VolumeMounts:    volumeMounts,
				Env:             handleEnvVars(runaiConfig),
				Resources:       v1.ResourceRequirements(*config.Resources),
				Ports: []v1.ContainerPort{
					{
						ContainerPort: int32(*config.Service.Metrics.Port),
						Name:          config.Service.Metrics.PortName,
					},
				},
			},
		},
		ImagePullSecrets: outils.GetGlobalImagePullSecrets(runaiConfig),
		Volumes:          volumes,
	}

	// Mount whisper-application-controller-secret (per-cluster agent credentials) so the controller
	// can obtain a control-plane bearer token for authenticated chart downloads.
	if err := secrets.MountToDeployment(ctx, k8sReader, deployment); err != nil {
		return nil, err
	}

	return deployment, nil
}

func handleEnvVars(runaiConfig *runaiv1.RunaiConfig) []v1.EnvVar {
	config := runaiConfig.Spec.ApplicationController
	envVars := []v1.EnvVar{
		{
			Name:  "K8S_CLIENT_CONFIG_QPS",
			Value: strconv.Itoa(*config.K8sClientConfig.QPS),
		},
		{
			Name:  "K8S_CLIENT_CONFIG_BURST",
			Value: strconv.Itoa(*config.K8sClientConfig.Burst),
		},
		{
			Name:  "USE_DEBUG_LOG_LEVEL",
			Value: strconv.FormatBool(*config.UseDebugLogLevel),
		},
		{
			Name:  "METRICS_PORT",
			Value: strconv.Itoa(*config.Service.Metrics.Port),
		},
		{
			Name:  "GODEBUG",
			Value: outils.FipsGodebugValue(runaiConfig, "off"),
		},
		{
			Name: constants.RunaiAdsEncryptionKeyEnvVar,
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: constants.RunaiAdsSecretName,
					},
					Key: constants.RunaiAdsEncryptionKeyName,
				},
			},
		},
	}
	// In headless installs the control-plane URL is the placeholder "headless", which is not a
	// reachable endpoint. Setting ENDPOINT there makes the controller treat control-plane
	// credentials as configured and fatally fail its startup token fetch. Omit it so
	// the controller falls back to the unauthenticated path.
	if !runaiConfig.Spec.Internal.Development.Headless {
		envVars = append(envVars, v1.EnvVar{
			Name:  "ENDPOINT",
			Value: runaiConfig.Spec.Internal.Global.ControlPlane.URL,
		})
	}
	if ptr.Deref(config.Replicas, 1) > 1 {
		envVars = append(envVars, v1.EnvVar{
			Name:  "LEADER_ELECT",
			Value: "true",
		})
	}
	return envVars
}

func DeploymentStatus(ctx context.Context,
	k8sReader client.Reader, runaiConfig *runaiv1.RunaiConfig) (operands.Status, error) {
	resource, err := DeploymentForRunaiConfig(ctx, k8sReader, runaiConfig)
	if err != nil {
		return operands.NotReadyStatus(), err
	}

	deployment := resource.(*appsv1.Deployment)

	if deployment.GetUID() == "" {
		return operands.NotReadyStatus(fmt.Sprintf("deployment [%s] is missing",
			deployment.GetName())), nil
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
