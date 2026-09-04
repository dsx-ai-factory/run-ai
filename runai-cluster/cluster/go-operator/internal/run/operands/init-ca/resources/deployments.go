/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"
	"hash/crc64"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	outils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	deploymentName = "init-ca"
)

var webhooksChecksumTable = crc64.MakeTable(crc64.ISO) // checksum uses that table to hash the config bytes quickly

func DeploymentForRunaiConfig(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runaiv1.RunaiConfig) (client.Object, error) {
	var (
		namespace = runaiConfig.GetNamespace()

		config = runaiConfig.Spec.InitCA
		image  = config.Image
	)

	deployment := &appsv1.Deployment{}

	// Get the existing job if it exists to consume any cluster-set values
	err := k8sReader.Get(ctx, types.NamespacedName{
		Name:      deploymentName,
		Namespace: namespace,
	}, deployment)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	deployment.Name = deploymentName
	deployment.Namespace = namespace

	deployment.Spec.Replicas = config.Replicas
	deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": deploymentName,
		},
	}
	deployment.Spec.Strategy = appsv1.DeploymentStrategy{
		Type: appsv1.RecreateDeploymentStrategyType,
	}

	deployment.Spec.Template.Labels = outils.InitMapIfNil(deployment.Spec.Template.Labels)
	deployment.Spec.Template.Labels["app"] = deploymentName

	deployment.Spec.Template.Annotations = outils.InitMapIfNil(deployment.Spec.Template.Annotations)
	// if webhooks specification changes (in configmap), it will change the annotation and cause the deployment to restart
	deployment.Spec.Template.Annotations["webhooks-checksum"] = getWebhooksDataChecksum(ctx, k8sReader, namespace)

	deployment.Spec.Template.Spec = v1.PodSpec{
		Affinity: outils.MergeAffinities(config.Affinity, runaiConfig.Spec.Global.Affinity,
			deployment.Spec.Selector.MatchLabels, runaiConfig.Spec.Global.RequireDefaultPodAntiAffinity),
		Tolerations: outils.GetTolerations(runaiConfig),
		// init-ca accesses Secrets, Pods, Deployments, cert-manager Certificates,
		// ValidatingWebhookConfigurations, and MutatingWebhookConfigurations.
		AutomountServiceAccountToken: ptr.To(true),
		ServiceAccountName:           outils.GetNMCServiceAccountNameOrDefault(runaiConfig, deploymentName),
		Volumes: append(outils.GlobalVolumes(runaiConfig),
			v1.Volume{
				Name: "webhooks",
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: deploymentName,
						},
					},
				},
			},
		),
		Containers: []v1.Container{
			{
				Name:            "create-ca",
				Image:           fmt.Sprintf("%s/%s:%s", *image.Registry, *image.Name, *image.Tag),
				ImagePullPolicy: image.PullPolicy,
				Args:            handleDeploymentArgs(),
				Resources:       v1.ResourceRequirements(*config.Resources),
				VolumeMounts: append(outils.GlobalVolumeMounts(runaiConfig),
					v1.VolumeMount{
						Name:      "webhooks",
						MountPath: "/webhooks",
					},
				),
				SecurityContext: initCAContainerSecurityContext(runaiConfig),
				Env: []v1.EnvVar{
					{Name: "GODEBUG", Value: outils.FipsGodebugValue(runaiConfig, "off")},
				},
			},
		},
		ImagePullSecrets: outils.GetGlobalImagePullSecrets(runaiConfig),
	}
	outils.AddNMCRemoteClusterAPIServerURLEnvVarIfConfigured(runaiConfig, &deployment.Spec.Template.Spec.Containers[0].Env)

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

func getWebhooksDataChecksum(ctx context.Context, k8sReader client.Reader, namespace string) string {
	configMap := &v1.ConfigMap{}
	err := k8sReader.Get(ctx, types.NamespacedName{
		Name:      deploymentName,
		Namespace: namespace,
	}, configMap)
	if err != nil {
		return ""
	}

	webhooksData, found := configMap.Data["webhooks.yaml"]
	if !found {
		return ""
	}

	checksum := crc64.Checksum([]byte(webhooksData), webhooksChecksumTable)
	return fmt.Sprintf("%016x", checksum)
}

func handleDeploymentArgs() []string {
	return []string{"--leader-elect"}
}

func initCAContainerSecurityContext(runaiConfig *runaiv1.RunaiConfig) *v1.SecurityContext {
	securityContext := outils.GlobalSecurityContext(runaiConfig)
	if securityContext == nil {
		securityContext = &v1.SecurityContext{}
	}

	var addedCapabilities []v1.Capability
	if securityContext.Capabilities != nil {
		addedCapabilities = append(addedCapabilities, securityContext.Capabilities.Add...)
	}
	securityContext.AllowPrivilegeEscalation = ptr.To(false)
	securityContext.Capabilities = &v1.Capabilities{
		Add:  addedCapabilities,
		Drop: []v1.Capability{"ALL"},
	}
	securityContext.ReadOnlyRootFilesystem = ptr.To(true)
	securityContext.RunAsNonRoot = ptr.To(true)
	securityContext.SeccompProfile = &v1.SeccompProfile{Type: v1.SeccompProfileTypeRuntimeDefault}

	return securityContext
}
