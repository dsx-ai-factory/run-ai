/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"

	metrics_exporter "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/metrics-exporter"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	outils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
)

const (
	DeploymentName = "metrics-exporter"
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
	deploymentNamespace := d.runaiConfig.GetNamespace()

	deployment := &appsv1.Deployment{}

	err := d.k8sReader.Get(ctx, types.NamespacedName{
		Name:      DeploymentName,
		Namespace: deploymentNamespace,
	}, deployment)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	deploymentForRunaiConfig(d.runaiConfig, deployment)

	return deployment, nil
}

func deploymentForRunaiConfig(runaiConfig *runaiv1.RunaiConfig, deployment *appsv1.Deployment) {
	deployment.Name = DeploymentName
	deployment.Namespace = runaiConfig.GetNamespace()
	config := runaiConfig.Spec.MetricsExporter
	image := config.Image

	deployment.Spec.Replicas = config.Replicas
	deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "metrics-exporter",
		},
	}

	deployment.Spec.Template.Name = DeploymentName
	deployment.Spec.Template.Labels = outils.InitMapIfNil(deployment.Spec.Template.Labels)
	deployment.Spec.Template.Labels["app"] = DeploymentName

	deployment.Spec.Template.Spec = v1.PodSpec{
		ServiceAccountName: outils.GetNMCServiceAccountNameOrDefault(runaiConfig, "metrics-exporter"),
		Affinity: outils.MergeAffinities(config.Affinity, runaiConfig.Spec.Global.Affinity,
			deployment.Spec.Selector.MatchLabels, runaiConfig.Spec.Global.RequireDefaultPodAntiAffinity),
		Tolerations: outils.GetTolerations(runaiConfig),
		Containers: []v1.Container{
			{
				Name:            "metrics-exporter",
				SecurityContext: outils.GlobalSecurityContext(runaiConfig),
				VolumeMounts:    outils.GlobalVolumeMounts(runaiConfig),
				Image:           fmt.Sprintf("%s/%s:%s", *image.Registry, *image.Name, *image.Tag),
				ImagePullPolicy: image.PullPolicy,
				Args:            handleDeploymentArgs(config),
				Env:             getEnvVars(runaiConfig),
				Ports:           []v1.ContainerPort{},
				Resources:       v1.ResourceRequirements(*config.Resources),
			},
		},
		ImagePullSecrets: outils.GetGlobalImagePullSecrets(runaiConfig),
		Volumes:          outils.GlobalVolumes(runaiConfig),
	}
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

	deploymentStatus := deployment.Status
	if len(deploymentStatus.Conditions) == 0 {
		return fmt.Errorf("deployment [%s] has no conditions", deployment.GetName())
	}

	conditions := deploymentStatus.Conditions
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

func getEnvVars(runaiConfig *runaiv1.RunaiConfig) []v1.EnvVar {
	envVars := []v1.EnvVar{
		{
			Name:  "CLUSTER_ID",
			Value: runaiConfig.Spec.Internal.Global.ControlPlane.ClusterUUID,
		},
		{
			Name:  "GODEBUG",
			Value: outils.FipsGodebugValue(runaiConfig, "off"),
		},
	}

	outils.AddNMCRemoteClusterAPIServerURLEnvVarIfConfigured(runaiConfig, &envVars)

	return envVars
}

func handleDeploymentArgs(config *metrics_exporter.MetricsExporter) []string {
	var args []string
	if *config.Replicas > 1 {
		args = append(args, "--leader-elect")
	}
	return args
}
