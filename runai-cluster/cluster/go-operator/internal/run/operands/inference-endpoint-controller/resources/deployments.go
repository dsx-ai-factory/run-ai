/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	outils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
)

const (
	DeploymentName = operands.RunaiInferenceEndpointControllerName
	appLabelKey    = "app"
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
	config := d.runaiConfig.Spec.InferenceEndpointController
	image := config.Image

	deployment := &appsv1.Deployment{}

	err := d.k8sReader.Get(ctx, types.NamespacedName{
		Name:      DeploymentName,
		Namespace: deploymentNamespace,
	}, deployment)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	deployment.Name = DeploymentName
	deployment.Namespace = deploymentNamespace

	deployment.Labels = outils.InitMapIfNil(deployment.Labels)
	deployment.Labels[appLabelKey] = DeploymentName

	deployment.Spec.Replicas = config.Replicas
	deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			appLabelKey: DeploymentName,
		},
	}

	deployment.Spec.Template.Name = DeploymentName
	deployment.Spec.Template.Labels = outils.InitMapIfNil(deployment.Spec.Template.Labels)
	deployment.Spec.Template.Labels[appLabelKey] = DeploymentName

	deployment.Spec.Template.Spec = v1.PodSpec{
		ServiceAccountName: DeploymentName,
		// Watches InferenceEndpoints and manages their routing, EPP, RBAC, and leader-election resources.
		AutomountServiceAccountToken: ptr.To(true),
		Affinity: outils.MergeAffinities(config.Affinity, d.runaiConfig.Spec.Global.Affinity,
			deployment.Spec.Selector.MatchLabels, d.runaiConfig.Spec.Global.RequireDefaultPodAntiAffinity),
		Tolerations: outils.GetTolerations(d.runaiConfig),
		Containers: []v1.Container{
			{
				Name:            DeploymentName,
				SecurityContext: controllerSecurityContext(d.runaiConfig),
				Image:           fmt.Sprintf("%s/%s:%s", *image.Registry, *image.Name, *image.Tag),
				ImagePullPolicy: image.PullPolicy,
				VolumeMounts:    outils.GlobalVolumeMounts(d.runaiConfig),
				Args:            deploymentArgs(d.runaiConfig),
				Env:             []v1.EnvVar{{Name: "GODEBUG", Value: outils.FipsGodebugValue(d.runaiConfig, "off")}},
				LivenessProbe:   &v1.Probe{ProbeHandler: v1.ProbeHandler{HTTPGet: &v1.HTTPGetAction{Path: "/healthz", Port: intstr.FromInt32(8081)}}, InitialDelaySeconds: 15, PeriodSeconds: 20},
				ReadinessProbe:  &v1.Probe{ProbeHandler: v1.ProbeHandler{HTTPGet: &v1.HTTPGetAction{Path: "/readyz", Port: intstr.FromInt32(8081)}}, InitialDelaySeconds: 5, PeriodSeconds: 10},
				Ports: []v1.ContainerPort{
					{
						ContainerPort: config.Metrics.Port,
						Name:          config.Metrics.PortName,
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

func deploymentArgs(runaiConfig *runaiv1.RunaiConfig) []string {
	config := runaiConfig.Spec.InferenceEndpointController
	logLevel := "info"
	if *config.UseDebugLogLevel {
		logLevel = "debug"
	}
	return []string{
		"--leader-elect",
		"--health-probe-bind-address=:8081",
		"--metrics-bind-address=:" + strconv.Itoa(int(config.Metrics.Port)),
		"--metrics-secure=false",
		"--zap-log-level=" + logLevel,
		"--zap-encoder=json",
		"--zap-devel=false",
	}
}

// controllerSecurityContext preserves administrator overrides and lets OpenShift
// assign a UID while supplying the controller's portable security defaults.
func controllerSecurityContext(config *runaiv1.RunaiConfig) *v1.SecurityContext {
	security := outils.GlobalSecurityContext(config)
	if security == nil {
		security = &v1.SecurityContext{}
	} else {
		security = security.DeepCopy()
	}
	if security.RunAsNonRoot == nil {
		security.RunAsNonRoot = ptr.To(true)
	}
	if security.ReadOnlyRootFilesystem == nil {
		security.ReadOnlyRootFilesystem = ptr.To(true)
	}
	if security.AllowPrivilegeEscalation == nil {
		security.AllowPrivilegeEscalation = ptr.To(false)
	}
	if security.Capabilities == nil {
		security.Capabilities = &v1.Capabilities{}
	}
	if len(security.Capabilities.Drop) == 0 {
		security.Capabilities.Drop = []v1.Capability{"ALL"}
	}
	if security.SeccompProfile == nil {
		security.SeccompProfile = &v1.SeccompProfile{Type: v1.SeccompProfileTypeRuntimeDefault}
	}
	return security
}
