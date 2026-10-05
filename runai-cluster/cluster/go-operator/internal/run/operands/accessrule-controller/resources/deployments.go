/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	outils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	runaiv1accessrule "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/accessrule-controller"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	name          = operands.RunaiAccessRuleControllerName
	certsDir      = "/etc/webhook/certs"
	tlsSecretName = "runai-accessrule-controller-tls-secret"
)

func DeploymentForRunaiConfig(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runaiv1.RunaiConfig) (client.Object, error) {
	var (
		deploymentNamespace = runaiConfig.GetNamespace()

		config = runaiConfig.Spec.AccessRuleController
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

	deployment.Spec.Replicas = config.Replicas
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
		Affinity: outils.MergeAffinities(config.Affinity, runaiConfig.Spec.Global.Affinity,
			deployment.Spec.Selector.MatchLabels, runaiConfig.Spec.Global.RequireDefaultPodAntiAffinity),
		Tolerations: outils.GetTolerations(runaiConfig),
		Containers: []v1.Container{
			{
				Name:            name,
				SecurityContext: outils.GlobalSecurityContext(runaiConfig),
				Image: fmt.Sprintf("%s/%s:%s",
					*image.Registry, *image.Name, *image.Tag),
				ImagePullPolicy: image.PullPolicy,
				VolumeMounts: append(outils.GlobalVolumeMounts(runaiConfig),
					v1.VolumeMount{
						MountPath: certsDir,
						Name:      "webhook-certs",
						ReadOnly:  true,
					},
				),
				Env: []v1.EnvVar{
					{Name: "GODEBUG", Value: outils.FipsGodebugValue(runaiConfig, "off")},
				},
				Args:      buildContainerArgs(config, runaiConfig),
				Resources: v1.ResourceRequirements(*config.Resources),
			},
		},
		ImagePullSecrets: outils.GetGlobalImagePullSecrets(runaiConfig),
		Volumes: append(outils.GlobalVolumes(runaiConfig),
			v1.Volume{
				Name: "webhook-certs",
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: tlsSecretName,
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

func buildContainerArgs(config *runaiv1accessrule.AccessRuleController, runaiConfig *runaiv1.RunaiConfig) []string {
	args := []string{
		"--webhook-port",
		strconv.Itoa(*config.Service.Webhook.TargetPort),
	}
	if *config.Replicas > 1 {
		args = append(args, "--leader-elect")
	}
	if config.Args.DebugLog {
		args = append(args, "--debug")
	}

	if config.Args.LegacyProjectPermissions != nil && *config.Args.LegacyProjectPermissions {
		args = append(args, "--legacy-project-permissions")
	}
	if config.Args.LegacyProjectPermissionsIntervalSeconds != nil {
		args = append(args,
			"--legacy-project-permissions-interval-seconds",
			strconv.Itoa(*config.Args.LegacyProjectPermissionsIntervalSeconds))
	}

	if allowedServiceAccounts := outils.GetNMCAllowedServiceAccountUsernames(runaiConfig); len(allowedServiceAccounts) > 0 {
		args = append(args, "--allowed-service-accounts")
		args = append(args, strings.Join(allowedServiceAccounts, ","))
	}

	return args
}
