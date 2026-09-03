/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"
	"strconv"

	clustercommonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	outils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	project_controller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/project-controller"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DeploymentName = operands.RunaiProjectControllerName
)

func DeploymentForRunaiConfig(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runaiv1.RunaiConfig) (client.Object, error) {
	var (
		deploymentNamespace = runaiConfig.GetNamespace()

		config = runaiConfig.Spec.ProjectController
		image  = config.Image
	)

	deployment := &appsv1.Deployment{}

	// Get the existing deployment if it exists to consume any cluster-set values
	err := k8sReader.Get(ctx, types.NamespacedName{
		Name:      DeploymentName,
		Namespace: deploymentNamespace,
	}, deployment)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	deployment.Name = DeploymentName
	deployment.Namespace = deploymentNamespace

	deployment.Spec.Replicas = config.Replicas
	deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": DeploymentName,
		},
	}

	deployment.Spec.Template.Name = DeploymentName
	deployment.Spec.Template.Labels = outils.InitMapIfNil(deployment.Spec.Template.Labels)
	deployment.Spec.Template.Labels["app"] = DeploymentName

	deployment.Spec.Template.Spec = v1.PodSpec{
		ServiceAccountName: DeploymentName,
		Affinity: outils.MergeAffinities(config.Affinity, runaiConfig.Spec.Global.Affinity,
			deployment.Spec.Selector.MatchLabels, runaiConfig.Spec.Global.RequireDefaultPodAntiAffinity),
		Tolerations: outils.GetTolerations(runaiConfig),
		Volumes:     outils.GlobalVolumes(runaiConfig),
		Containers: []v1.Container{
			{
				Name:            DeploymentName,
				SecurityContext: outils.GlobalSecurityContext(runaiConfig),
				Image:           fmt.Sprintf("%s/%s:%s", *image.Registry, *image.Name, *image.Tag),
				ImagePullPolicy: image.PullPolicy,
				Env: []v1.EnvVar{
					{
						Name:  "METRICS_PORT",
						Value: strconv.Itoa(*config.Service.Metrics.Port),
					},
					{
						Name:  "GODEBUG",
						Value: outils.FipsGodebugValue(runaiConfig, "off"),
					},
				},
				Args: buildContainerArgs(config, deploymentNamespace),
				Ports: []v1.ContainerPort{
					{
						ContainerPort: int32(*config.Service.Metrics.Port),
						Name:          config.Service.Metrics.PortName,
					},
				},
				Resources:    v1.ResourceRequirements(*config.Resources),
				VolumeMounts: outils.GlobalVolumeMounts(runaiConfig),
			},
		},
		TerminationGracePeriodSeconds: ptr.To[int64](10),
		ImagePullSecrets:              outils.GetGlobalImagePullSecrets(runaiConfig),
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
		return operands.NotReadyStatus(
			fmt.Sprintf("deployment [%s] has no conditions",
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

func buildContainerArgs(config *project_controller.ProjectController, installNamespace string) []string {
	args := []string{
		// runai overrides — project-controller binary defaults are KAI-style;
		// preserve the runai vocabulary explicitly so runtime behavior is unchanged.
		"--nodepool-label-key", v1alpha1.NodePoolNameLabel,
		"--default-nodepool-name", v1alpha1.DefaultNodePoolName,
		"--queue-label-key", clustercommonconstants.RunaiQueueLabel,
		"--namespace-project-label-key", clustercommonconstants.RunaiQueueLabel,
		"--project-label-key", clustercommonconstants.RunaiProjectLabel,
		"--project-id-label-key", clustercommonconstants.ProjectIdLabel,
		"--queue-department-name-label-key", clustercommonconstants.QueueDepartmentNameLabel,
		"--namespace-version-label-key", clustercommonconstants.RunaiNamespaceVersionLabel,
		"--enforce-scheduler-annotation-key", clustercommonconstants.EnforceSchedulerAnnotationName,
		"--finalizer-domain", clustercommonconstants.RunaiOrgDomain,
		"--install-namespace", installNamespace,
		"--project-name-prefix", clustercommonconstants.NamespaceRunAi,
		"--rolebindings-configmap-name", clustercommonconstants.ProjectControllerRoleBindingCm,
		"--rolebindings-configmap-namespace", clustercommonconstants.NamespaceRunAi,
		"--resource-manual-override-label-key", clustercommonconstants.RunaiResourceManualOverrideLabel,
		"--project-delete-blockers-namespace", clustercommonconstants.NamespaceRunAi,
		"--enable-project-validation-webhook=false",
		"--enable-department-validation-webhook=false",
		"--metrics-port",
		strconv.Itoa(*config.Service.Metrics.Port),
		"--profiler-api-port",
		config.Profiling.ApiPort,
		"--qps",
		strconv.Itoa(*config.K8sClientConfig.QPS),
		"--burst",
		strconv.Itoa(*config.K8sClientConfig.Burst),
	}

	if config.IsOpenShift {
		args = append(args, "--openshift")
	}

	if *config.CreateNamespaces {
		args = append(args, "--namespaces")
	}

	if *config.CreateRoleBindings {
		args = append(args, "--role-bindings")
	}

	if *config.ClusterWideSecret {
		args = append(args, "--cluster-wide-secrets")
	}

	if *config.ClusterWidePvc {
		args = append(args, "--cluster-wide-pvcs")
	}

	if *config.ClusterWideConfigMap {
		args = append(args, "--cluster-wide-config-maps")
	}

	if *config.LimitRange {
		args = append(args, "--limit-range",
			"--limit-range-name", clustercommonconstants.RunaiLimitRangeName)
	}

	if *config.Profiling.Enabled {
		args = append(args, "--enable-profiling")
	}

	if *config.Args.Verbosity < 0 {
		args = append(args, "--debug")
	}

	if *config.Replicas > 1 {
		args = append(args, "--leader-elect")
	}

	return args
}
