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

	"k8s.io/utils/ptr"

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
	DeploymentName                     = "runtime-rules"
	deploymentName                     = "runtime-rules"
	serviceAccountName                 = "runtime-rules"
	numberOfWorkersPerQueue            = "NUMBER_OF_WORKERS_PER_QUEUE"
	workerSleepSeconds                 = "WORKER_SLEEP_SECONDS"
	numberOfWorkersAutoDeleteCompleted = "NUMBER_OF_WORKERS_AUTO_DELETE_COMPLETED"
	eventList                          = "EVENT_LIST"
	deploymentSelectorKey              = "app"
	deploymentSelectorValue            = "runtime-rules"
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
	config := d.runaiConfig.Spec.RuntimeRules
	image := config.Image

	deployment := &appsv1.Deployment{}

	err := d.k8sReader.Get(ctx, types.NamespacedName{
		Name:      DeploymentName,
		Namespace: deploymentNamespace,
	}, deployment)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	deployment.ObjectMeta.Name = DeploymentName
	deployment.ObjectMeta.Namespace = deploymentNamespace

	replicas := config.Replicas
	deployment.Spec.Replicas = replicas
	deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			deploymentSelectorKey: deploymentSelectorValue,
		},
	}

	deployment.Spec.Template.Name = DeploymentName
	deployment.Spec.Template.Labels = outils.InitMapIfNil(deployment.Spec.Template.Labels)
	deployment.Spec.Template.Labels["app"] = DeploymentName
	deployment.Spec.Template.Spec = v1.PodSpec{
		ServiceAccountName: serviceAccountName,
		Affinity: outils.MergeAffinities(config.Affinity, d.runaiConfig.Spec.Global.Affinity,
			deployment.Spec.Selector.MatchLabels, d.runaiConfig.Spec.Global.RequireDefaultPodAntiAffinity),
		Tolerations: outils.GetTolerations(d.runaiConfig),
		Containers: []v1.Container{
			{
				Name:            deploymentName,
				SecurityContext: outils.GlobalSecurityContext(d.runaiConfig),
				Image:           fmt.Sprintf("%s/%s:%s", *image.Registry, *image.Name, *image.Tag),
				ImagePullPolicy: image.PullPolicy,
				VolumeMounts:    outils.GlobalVolumeMounts(d.runaiConfig),
				Args:            handleDeploymentArgs(d.runaiConfig),
				Env: []v1.EnvVar{
					{
						Name:  workerSleepSeconds,
						Value: strconv.Itoa(*config.WorkerSleepSeconds),
					},
					{
						Name:  numberOfWorkersPerQueue,
						Value: strconv.Itoa(*config.NumberOfWorkersPerQueue),
					},
					{
						Name:  numberOfWorkersAutoDeleteCompleted,
						Value: strconv.Itoa(*config.NumberOfWorkersAutoDeleteCompleted),
					},
					{
						Name:  eventList,
						Value: strings.Trim(strings.Join(strings.Fields(fmt.Sprint(*config.EventList)), ","), "[]"),
					},
					{
						Name:  "GODEBUG",
						Value: outils.FipsGodebugValue(d.runaiConfig, "off"),
					},
				},
				Resources: v1.ResourceRequirements(*config.Resources),
			},
		},
		ImagePullSecrets: outils.GetGlobalImagePullSecrets(d.runaiConfig),
		Volumes:          outils.GlobalVolumes(d.runaiConfig),
	}

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

	deploymentStatus := deployment.Status
	if len(deploymentStatus.Conditions) == 0 {
		return fmt.Errorf("deployment [%s] has no conditions", deployment.GetName())
	}

	conditions := deploymentStatus.Conditions
	for _, condition := range conditions {
		if condition.Type == appsv1.DeploymentAvailable &&
			condition.Status == v1.ConditionTrue {
			return nil
		}
	}

	return fmt.Errorf("deployment [%s] has not reached the Available condition", deployment.GetName())
}

func (d *DeploymentResource) Enabled() bool {
	return true
}

func handleDeploymentArgs(runaiConfig *runaiv1.RunaiConfig) []string {
	config := runaiConfig.Spec.RuntimeRules
	args := []string{
		"--log-level",
		strconv.Itoa(*config.LogLevel),
		"--prom-service-url=" + runaiConfig.Spec.Global.PrometheusService.Address,
		"--framework-polling-interval-sec=" + strconv.Itoa(*config.DistributedFrameworkStatusPollingIntervalSeconds),
	}
	if ptr.Deref(config.Replicas, 1) > 1 {
		args = append(args, "--leader-elect")
	}
	if ptr.Deref(config.DisableWorkloadRunnerRules, false) {
		args = append(args, "--disable-workload-runner-rules")
	}
	return args
}
