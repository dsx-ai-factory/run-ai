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

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	runv1internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/secrets"

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
	DeploymentName = operands.RunaiAgentName
	certsDir       = "/etc/webhook/certs"
	tlsSecretName  = "runai-agent-tls-secret"
)

type DeploymentResource struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runaiv1.RunaiConfig
}

func NewDeploymentResource(operandName string, k8sReader client.Reader, runaiConfig *runaiv1.RunaiConfig) *DeploymentResource {
	return &DeploymentResource{operandName: operandName, k8sReader: k8sReader, runaiConfig: runaiConfig}
}

func (d *DeploymentResource) OperandName() string { return d.operandName }

func (d *DeploymentResource) Object(ctx context.Context) (client.Object, error) {
	var (
		deploymentNamespace = d.runaiConfig.GetNamespace()
		config              = d.runaiConfig.Spec.Agent
		image               = config.Image
	)

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

	var replicas *int32
	if d.runaiConfig.Spec.Internal.Development.Headless {
		replicas = ptr.To(int32(0))
	} else {
		replicas = config.Replicas
	}

	deployment.Spec.Replicas = replicas
	deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{"app": DeploymentName},
	}

	deployment.Spec.Template.Name = DeploymentName
	deployment.Spec.Template.Labels = outils.InitMapIfNil(deployment.Spec.Template.Labels)
	deployment.Spec.Template.Labels["app"] = DeploymentName
	deployment.Spec.Template.Spec = v1.PodSpec{
		ServiceAccountName: outils.GetNMCServiceAccountNameOrDefault(d.runaiConfig, DeploymentName),
		Affinity: outils.MergeAffinities(config.Affinity, d.runaiConfig.Spec.Global.Affinity,
			deployment.Spec.Selector.MatchLabels, d.runaiConfig.Spec.Global.RequireDefaultPodAntiAffinity),
		Tolerations: outils.GetTolerations(d.runaiConfig),
		Containers: []v1.Container{
			{
				Name:            DeploymentName,
				SecurityContext: outils.GlobalSecurityContext(d.runaiConfig),
				Image:           fmt.Sprintf("%s/%s:%s", *image.Registry, *image.Name, *image.Tag),
				ImagePullPolicy: image.PullPolicy,
				VolumeMounts: append(append(outils.GlobalVolumeMounts(d.runaiConfig), volumeMountsForHTTPProxy(d.runaiConfig)...),
					v1.VolumeMount{
						MountPath: certsDir,
						Name:      "webhook-certs",
						ReadOnly:  true,
					},
				),
				Env: handleEnvVars(d.runaiConfig),
				Ports: []v1.ContainerPort{
					{ContainerPort: config.Metrics.Port, Name: config.Metrics.PortName},
				},
				Resources: v1.ResourceRequirements(*config.Resources),
			},
		},
		ImagePullSecrets: outils.GetGlobalImagePullSecrets(d.runaiConfig),
		Volumes: append(append(outils.GlobalVolumes(d.runaiConfig), volumesForHTTPProxy(d.runaiConfig)...),
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

	err = secrets.MountToDeployment(ctx, d.k8sReader, deployment)
	if err != nil {
		return nil, err
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
	for _, condition := range deplopymentStatus.Conditions {
		if condition.Type == appsv1.DeploymentAvailable &&
			condition.Status == v1.ConditionStatus(metav1.ConditionTrue) {
			return nil
		}
	}
	return fmt.Errorf("deployment [%s] has not reached the Available condition", deployment.GetName())
}

func (d *DeploymentResource) Enabled() bool { return true }

func handleEnvVars(runaiConfig *runaiv1.RunaiConfig) []v1.EnvVar {
	config := runaiConfig.Spec.Agent
	envVars := []v1.EnvVar{
		{
			Name:  "IS_OPENSHIFT",
			Value: strconv.FormatBool(runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor == runv1internal.KubernetesFlavorOpenShift),
		},
		{
			Name:  "METRICS_PORT",
			Value: strconv.Itoa(int(config.Metrics.Port)),
		},
		{
			Name:  "AGENT_SYNC_INTERVAL_MILLIS",
			Value: strconv.Itoa(*config.SyncIntervalMillis),
		},
		{
			Name:  "ENDPOINT",
			Value: runaiConfig.Spec.Internal.Global.ControlPlane.URL,
		},
		{
			Name:  "API_VERSION",
			Value: config.APIVersion,
		},
		{
			Name:  "TLS_REJECT_UNAUTHORIZED",
			Value: strconv.FormatBool(*config.RejectUnauthorized),
		},
		{
			Name:  "ACCESS_RULES_LIMIT",
			Value: strconv.Itoa(*config.AccessRulesLimit),
		},
		{
			Name:  "CLUSTER_DOMAIN",
			Value: runaiConfig.Spec.Internal.Global.ClusterURL,
		},
		{
			Name:  "CLUSTER_VERSION",
			Value: runaiConfig.Spec.Internal.Global.Version,
		},
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
			Name:  "POLICY_SYNC_ENABLED",
			Value: strconv.FormatBool(*runaiConfig.Spec.Global.ControlPlanePoliciesEnabled),
		},
		{
			Name: constants.RunaiAdsEncryptionKeyEnvVar,
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{Name: constants.RunaiAdsSecretName},
					Key:                  constants.RunaiAdsEncryptionKeyName,
				},
			},
		},
		{
			Name:  "GODEBUG",
			Value: outils.FipsGodebugValue(runaiConfig, "off"),
		},
	}
	envVars = append(envVars,
		v1.EnvVar{Name: "WEBHOOK_PORT", Value: strconv.Itoa(*runaiConfig.Spec.ProjectController.Service.Webhook.Port)},
		v1.EnvVar{Name: "CERT_DIR", Value: certsDir},
	)
	if len(runaiConfig.Spec.Internal.Global.NMC.AllowedServiceAccounts) > 0 {
		envVars = append(envVars, v1.EnvVar{
			Name:  "ALLOWED_SERVICE_ACCOUNTS",
			Value: strings.Join(runaiConfig.Spec.Internal.Global.NMC.AllowedServiceAccounts, ","),
		})
	}
	envVars = append(envVars, envVarsForHTTPProxy(runaiConfig)...)
	if ptr.Deref(config.Replicas, 1) > 1 {
		envVars = append(envVars, v1.EnvVar{Name: "LEADER_ELECT", Value: "true"})
	}
	outils.AddNMCRemoteClusterAPIServerURLEnvVarIfConfigured(runaiConfig, &envVars)
	return envVars
}
