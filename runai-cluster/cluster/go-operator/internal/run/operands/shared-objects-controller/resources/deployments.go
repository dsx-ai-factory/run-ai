/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	outils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	name = operands.RunaiSharedObjectsControllerName
)

type DeploymentResource struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runaiv1.RunaiConfig
}

func NewDeploymentResource(operandName string, k8sReader client.Reader, runaiConfig *runaiv1.RunaiConfig) *DeploymentResource {
	return &DeploymentResource{
		operandName: operandName,
		k8sReader:   k8sReader,
		runaiConfig: runaiConfig,
	}
}

func (d *DeploymentResource) OperandName() string {
	return d.operandName
}

func (d *DeploymentResource) Object(ctx context.Context) (client.Object, error) {
	var (
		deploymentNamespace     = d.runaiConfig.GetNamespace()
		config                  = d.runaiConfig.Spec.SharedObjectsController
		projectControllerConfig = d.runaiConfig.Spec.ProjectController
		image                   = config.Image
	)
	deployment := &appsv1.Deployment{}

	err := d.k8sReader.Get(ctx, types.NamespacedName{
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
		ServiceAccountName: outils.GetNMCServiceAccountNameOrDefault(d.runaiConfig, name),
		Affinity: outils.MergeAffinities(config.Affinity, d.runaiConfig.Spec.Global.Affinity,
			deployment.Spec.Selector.MatchLabels, d.runaiConfig.Spec.Global.RequireDefaultPodAntiAffinity),
		Tolerations: outils.GetTolerations(d.runaiConfig),
		Containers: []v1.Container{
			{
				Name:            name,
				SecurityContext: outils.GlobalSecurityContext(d.runaiConfig),
				Image:           fmt.Sprintf("%s/%s:%s", *image.Registry, *image.Name, *image.Tag),
				ImagePullPolicy: image.PullPolicy,
				VolumeMounts:    outils.GlobalVolumeMounts(d.runaiConfig),
				Env: []v1.EnvVar{
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
						Name:  "BASE_CONTROLLER_DELAY_MILLISECONDS",
						Value: strconv.Itoa(*config.BaseControllerDelayMilliseconds),
					},
					{
						Name:  "MAX_CONTROLLER_DELAY_SECONDS",
						Value: strconv.Itoa(*config.MaxControllerDelaySeconds),
					},
					{
						Name:  "GODEBUG",
						Value: outils.FipsGodebugValue(d.runaiConfig, "off"),
					},
					{
						Name:  "CREATE_SERVICE_MONITORS",
						Value: strconv.FormatBool(*projectControllerConfig.CreateServiceMonitors),
					},
					{
						Name:  "CLUSTER_WIDE_SECRETS",
						Value: strconv.FormatBool(*projectControllerConfig.ClusterWideSecret),
					},
					{
						Name:  "CLUSTER_WIDE_PVCS",
						Value: strconv.FormatBool(*projectControllerConfig.ClusterWidePvc),
					},
					{
						Name:  "CLUSTER_WIDE_CONFIG_MAPS",
						Value: strconv.FormatBool(*projectControllerConfig.ClusterWideConfigMap),
					},
					{
						Name:  "CLUSTER_WIDE_RESOURCE_CLAIM_TEMPLATES",
						Value: strconv.FormatBool(*projectControllerConfig.ClusterWideResourceClaimTemplate),
					},
				},
				Args: func() []string {
					args := []string{}
					if *config.Replicas > 1 {
						args = append(args, "--leader-elect")
					}
					return args
				}(),
				Resources: v1.ResourceRequirements(*config.Resources),
			},
		},
		ImagePullSecrets: outils.GetGlobalImagePullSecrets(d.runaiConfig),
		Volumes:          outils.GlobalVolumes(d.runaiConfig),
	}

	outils.AddNMCRemoteClusterAPIServerURLEnvVarIfConfigured(d.runaiConfig, &deployment.Spec.Template.Spec.Containers[0].Env)

	return deployment, nil
}

func (d *DeploymentResource) Status(ctx context.Context) error {
	resource, err := d.Object(ctx)
	if err != nil {
		return err
	}

	deployment := resource.(*appsv1.Deployment)

	if deployment.GetUID() == "" {
		return fmt.Errorf("deployment [%s] is missing", deployment.GetName())
	}

	deplopymentStatus := deployment.Status
	if len(deplopymentStatus.Conditions) == 0 {
		return fmt.Errorf("deployment [%s] has no conditions", deployment.GetName())
	}

	conditions := deplopymentStatus.Conditions
	for _, condition := range conditions {
		if condition.Type == appsv1.DeploymentAvailable &&
			condition.Status == v1.ConditionStatus(metav1.ConditionTrue) {
			return nil
		}
	}

	return fmt.Errorf("deployment [%s] has not reached the Available condition", deployment.GetName())
}

func (d *DeploymentResource) Enabled() bool {
	return true
}
