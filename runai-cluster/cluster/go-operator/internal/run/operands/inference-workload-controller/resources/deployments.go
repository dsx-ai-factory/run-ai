/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	custom_ca "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/common/custom-ca"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"gopkg.in/yaml.v2"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"

	runv1internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	outils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	name = operands.RunaiInferenceWorkloadControllerName
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
	global := d.runaiConfig.Spec.Global
	subdomainSupport := true
	if global != nil {
		subdomainSupport = ptr.Deref(global.SubdomainSupport, true)
	}
	internal := d.runaiConfig.Spec.Internal
	workloadsAuthClient := d.runaiConfig.Spec.Internal.Global.ControlPlane.Config.GetWorkloads().Client
	config := d.runaiConfig.Spec.InferenceWorkloadController
	image := config.Image
	replicas := config.Replicas

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

	envVars := []v1.EnvVar{
		{
			Name:  "IS_OPENSHIFT",
			Value: strconv.FormatBool(d.runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor == runv1internal.KubernetesFlavorOpenShift),
		},
		{
			Name:  "CLUSTER_DOMAIN",
			Value: d.runaiConfig.Spec.Internal.Global.ClusterURL,
		},
		{
			Name:  "SUBDOMAIN_SUPPORT",
			Value: strconv.FormatBool(subdomainSupport),
		},
		{
			Name:  "CLUSTER_VERSION",
			Value: d.runaiConfig.Spec.Internal.Global.Version,
		},
		{
			Name:  "ENDPOINT",
			Value: internal.Global.ControlPlane.URL,
		},
		{
			Name:  "WORKLOADS_AUTH_CLIENT_ID",
			Value: workloadsAuthClient.GetId(),
		},
		{
			Name:  "WORKLOADS_AUTH_CLIENT_NAME",
			Value: workloadsAuthClient.GetName(),
		},
		{
			Name: "WORKLOADS_AUTH_CLIENT_SECRET",
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: constants.WorkloadAuthSecretName,
					},
					Key: constants.WorkloadAuthClientSecretKey,
				},
			},
		},
		{
			Name: "WORKLOADS_AUTH_PROXY_COOKIE_SECRET",
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: constants.WorkloadAuthSecretName,
					},
					Key: constants.WorkloadAuthCookieSecretKey,
				},
			},
		},
		{
			Name:  "OAUTH2_PROXY_OIDC_ISSUER_URL",
			Value: internal.Global.ControlPlane.Config.GetResearcherService().OauthProxy.GetOidcIssuerUrl(),
		},
		{
			Name:  "CONTROL_PLANE_POLICIES_ENABLED",
			Value: strconv.FormatBool(*global.ControlPlanePoliciesEnabled),
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
		{Name: "GODEBUG", Value: outils.FipsGodebugValue(d.runaiConfig, "off")},
	}

	knativeInstalled, lwsInstalled := getKnativeAndLwsStatusFromRunaiPublic(ctx, d.k8sReader, d.runaiConfig.Namespace)

	envVars = append(envVars, v1.EnvVar{
		Name:  constants.KnativeInstalledEnvVar,
		Value: strconv.FormatBool(knativeInstalled),
	})

	envVars = append(envVars, v1.EnvVar{
		Name:  constants.LwsInstalledEnvVar,
		Value: strconv.FormatBool(lwsInstalled),
	})

	if err := outils.AddWorkloadInjectedContainerResourcesEnvVarsIfConfigured(global, &envVars); err != nil {
		return nil, err
	}

	if global != nil && global.Ingress != nil {
		if len(global.Ingress.Annotations) > 0 {
			annotationsJson, err := json.Marshal(global.Ingress.Annotations)
			if err != nil {
				return nil, err
			}
			envVars = append(envVars, v1.EnvVar{
				Name:  "INGRESS_ANNOTATIONS",
				Value: string(annotationsJson),
			})
		}
		if global.Ingress.Class != "" {
			envVars = append(envVars, v1.EnvVar{
				Name:  "INGRESS_CLASS",
				Value: global.Ingress.Class,
			})
		}
	}

	if global != nil && global.Ingress != nil && global.Ingress.Enabled != nil {
		envVars = append(envVars, v1.EnvVar{
			Name:  "INGRESS_ENABLED",
			Value: strconv.FormatBool(*global.Ingress.Enabled),
		})
	}

	if global != nil && global.GatewayAPI != nil && global.GatewayAPI.Enabled != nil {
		envVars = append(envVars, v1.EnvVar{
			Name:  "GATEWAY_API_ENABLED",
			Value: strconv.FormatBool(*global.GatewayAPI.Enabled),
		})
		if *global.GatewayAPI.Enabled {
			if global.GatewayAPI.Name == "" || global.GatewayAPI.Namespace == "" {
				return nil, fmt.Errorf("gatewayAPI is enabled but name or namespace is empty")
			}
			envVars = append(envVars, v1.EnvVar{
				Name:  "GATEWAY_NAME",
				Value: global.GatewayAPI.Name,
			})
			envVars = append(envVars, v1.EnvVar{
				Name:  "GATEWAY_NAMESPACE",
				Value: global.GatewayAPI.Namespace,
			})
		}
	}

	if config.RunaiOauth2ProxyImage != nil {
		runaiOauth2ProxyImage := config.RunaiOauth2ProxyImage
		envVars = append(envVars, v1.EnvVar{Name: "RUNAI_OAUTH2_PROXY_IMAGE", Value: fmt.Sprintf("%s/%s:%s",
			*runaiOauth2ProxyImage.Registry, *runaiOauth2ProxyImage.Name, *runaiOauth2ProxyImage.Tag)})
	}

	if config.RunaiOauth2ProxyRedisImage != nil {
		runaiOauth2ProxyRedisImage := config.RunaiOauth2ProxyRedisImage
		envVars = append(envVars, v1.EnvVar{Name: "RUNAI_OAUTH2_PROXY_REDIS_IMAGE", Value: fmt.Sprintf("%s/%s:%s",
			*runaiOauth2ProxyRedisImage.Registry, *runaiOauth2ProxyRedisImage.Name, *runaiOauth2ProxyRedisImage.Tag)})
	}

	if config.RunaiOauth2ProxyRedisFQDN != nil {
		runaiOauth2ProxyRedisFQDN := config.RunaiOauth2ProxyRedisFQDN
		envVars = append(envVars, v1.EnvVar{Name: "RUNAI_OAUTH2_PROXY_REDIS_FQDN", Value: *runaiOauth2ProxyRedisFQDN})
	}

	if config.RunaiOauth2ProxyWorkloadAuthSecretName != nil {
		runaiOauth2ProxyWorkloadAuthSecretName := config.RunaiOauth2ProxyWorkloadAuthSecretName
		envVars = append(envVars, v1.EnvVar{Name: "RUNAI_OAUTH2_PROXY_WORKLOAD_AUTH_SECRET_NAME", Value: *runaiOauth2ProxyWorkloadAuthSecretName})
	}

	if config.RunaiOauth2ProxySessionStoreType != nil {
		runaiOauth2ProxySessionStoreType := config.RunaiOauth2ProxySessionStoreType
		envVars = append(envVars, v1.EnvVar{Name: "RUNAI_OAUTH2_PROXY_SESSION_STORE_TYPE", Value: *runaiOauth2ProxySessionStoreType})
	}

	if config.RunaiOauth2ProxyUsernameLoaderImage != nil {
		runaiOauth2ProxyUsernameLoaderImage := config.RunaiOauth2ProxyUsernameLoaderImage
		envVars = append(envVars, v1.EnvVar{Name: "RUNAI_OAUTH2_PROXY_USERNAME_LOADER_IMAGE", Value: fmt.Sprintf("%s/%s:%s",
			*runaiOauth2ProxyUsernameLoaderImage.Registry, *runaiOauth2ProxyUsernameLoaderImage.Name, *runaiOauth2ProxyUsernameLoaderImage.Tag)})
	}

	if config.GitSyncImage != nil {
		gitSyncImage := config.GitSyncImage
		envVars = append(envVars, v1.EnvVar{Name: "GIT_SYNC_IMAGE", Value: fmt.Sprintf("%s/%s:%s",
			*gitSyncImage.Registry, *gitSyncImage.Name, *gitSyncImage.Tag)})
	}

	if config.RunaiLlmGatewayImage != nil {
		runaiLlmGatewayImage := config.RunaiLlmGatewayImage
		envVars = append(envVars, v1.EnvVar{Name: "RUNAI_LLM_GATEWAY_IMAGE", Value: fmt.Sprintf("%s/%s:%s",
			*runaiLlmGatewayImage.Registry, *runaiLlmGatewayImage.Name, *runaiLlmGatewayImage.Tag)})
	}

	if config.RunaiLlmGatewayResourcesCpuCore != nil {
		envVars = append(envVars, v1.EnvVar{Name: "RUNAI_LLM_GATEWAY_CPU_CORE", Value: *config.RunaiLlmGatewayResourcesCpuCore})
	}

	if config.RunaiLlmGatewayResourcesCpuMemory != nil {
		envVars = append(envVars, v1.EnvVar{Name: "LLM_GATEWAY_RESOURCES_CPU_MEMORY", Value: *config.RunaiLlmGatewayResourcesCpuMemory})
	}

	if config.K8sClientBurst != nil {
		envVars = append(envVars, v1.EnvVar{Name: "K8S_CLIENT_BURST", Value: *config.K8sClientBurst})
	}

	if config.K8sClientQPS != nil {
		envVars = append(envVars, v1.EnvVar{Name: "K8S_CLIENT_QPS", Value: *config.K8sClientQPS})
	}

	externalAuthUrlEnabled := "true"
	if config.ExternalAuthUrlEnabled != nil && *config.ExternalAuthUrlEnabled == false {
		externalAuthUrlEnabled = "false"
	}
	envVars = append(envVars, v1.EnvVar{Name: "EXTERNAL_AUTH_URL_ENABLED", Value: externalAuthUrlEnabled})

	if wc := d.runaiConfig.Spec.WorkloadController; wc != nil {
		envVars = append(envVars, v1.EnvVar{
			Name:  "OPENSHIFT_ROUTE_TERMINATION_POLICY",
			Value: string(wc.OpenshiftRouteConfiguration.GetTerminationPolicy()),
		})
	}

	if config.ApplyControlPlanePolicies != nil {
		envVars = append(envVars, v1.EnvVar{Name: "APPLY_CONTROL_PLANE_POLICIES", Value: strconv.FormatBool(*config.ApplyControlPlanePolicies)})
	}

	customCAs := map[string]*common.CustomCA{
		custom_ca.EnvVarPrefix:    internal.Global.CustomCA,
		custom_ca.EnvVarPrefixGit: global.CustomCAGit,
		custom_ca.EnvVarPrefixS3:  global.CustomCAS3,
	}

	envVars = append(envVars, custom_ca.ConfigEnvVars(customCAs)...)

	deployment.Spec.Replicas = replicas
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
		Affinity: outils.MergeAffinities(config.Affinity, d.runaiConfig.Spec.Global.Affinity,
			deployment.Spec.Selector.MatchLabels, d.runaiConfig.Spec.Global.RequireDefaultPodAntiAffinity),
		Tolerations: outils.GetTolerations(d.runaiConfig),
		Containers: []v1.Container{
			{
				Name:            name,
				SecurityContext: outils.GlobalSecurityContext(d.runaiConfig),
				Image:           fmt.Sprintf("%s/%s:%s", *image.Registry, *image.Name, *image.Tag),
				ImagePullPolicy: image.PullPolicy,
				VolumeMounts: append(outils.GlobalVolumeMounts(d.runaiConfig),
					v1.VolumeMount{
						MountPath: "/etc/webhook/certs",
						Name:      "webhook-certs",
						ReadOnly:  true,
					},
				),
				Env: envVars,
				Args: func() []string {
					args := []string{}

					if global.EnableWorkloadOwnershipProtection != nil && *global.EnableWorkloadOwnershipProtection {
						args = append(args, "--enable-workloads-ownership-protection")
					}
					if config.HealthProbe != nil && config.HealthProbe.Port != nil {
						args = append(args, fmt.Sprintf("--health-probe-bind-address=:%d", *config.HealthProbe.Port))
					}

					if config.Metric.Port != nil {
						args = append(args, fmt.Sprintf("--metrics-bind-address=:%d", *config.Metric.Port))
					}

					if *replicas > 1 {
						args = append(args, "--leader-elect")
					}

					if config.EnableTopology != nil && *config.EnableTopology {
						args = append(args, "--enable-topology-support")
					}

					if config.EnableElasticComputeDomain != nil && *config.EnableElasticComputeDomain {
						args = append(args, "--enable-elastic-compute-domain")
					}

					if config.GPUNetworkAccelerationEnabled != nil && *config.GPUNetworkAccelerationEnabled {
						args = append(args, "--enable-mnnvl-support")
					}

					if len(d.runaiConfig.Spec.Internal.Global.NMC.AllowedServiceAccounts) > 0 {
						args = append(args, "--allowed-service-accounts")
						args = append(args, strings.Join(d.runaiConfig.Spec.Internal.Global.NMC.AllowedServiceAccounts, ","))
					}
					return args
				}(),
				Resources: v1.ResourceRequirements(*config.Resources),
			},
		},
		ImagePullSecrets: outils.GetGlobalImagePullSecrets(d.runaiConfig),
		Volumes: append(outils.GlobalVolumes(d.runaiConfig),
			v1.Volume{
				Name: "webhook-certs",
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: "runai-inference-workload-controller-tls-secret",
					},
				},
			},
		),
	}

	return deployment, nil
}

func getKnativeAndLwsStatusFromRunaiPublic(ctx context.Context, k8sReader client.Reader, namespace string) (bool, bool) {
	configMap := &v1.ConfigMap{}
	err := k8sReader.Get(ctx, types.NamespacedName{
		Name:      runv1.RunaiPublicConfigMapName,
		Namespace: namespace,
	}, configMap)
	if err != nil {
		return false, false
	}

	runaiPublicData, exists := configMap.Data["runai-public"]
	if !exists {
		return false, false
	}

	var runaiPublic runv1.RunaiPublic
	err = yaml.Unmarshal([]byte(runaiPublicData), &runaiPublic)
	if err != nil {
		return false, false
	}

	if runaiPublic.Dependencies == nil || runaiPublic.Dependencies.Optional == nil {
		return false, false
	}

	knativeStatus, knativeExists := runaiPublic.Dependencies.Optional["knative"]
	knativeInstalled := knativeExists && knativeStatus.Available != nil && *knativeStatus.Available

	lwsStatus, lwsExists := runaiPublic.Dependencies.Optional["lws"]
	lwsInstalled := lwsExists && lwsStatus.Available != nil && *lwsStatus.Available

	return knativeInstalled, lwsInstalled
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
