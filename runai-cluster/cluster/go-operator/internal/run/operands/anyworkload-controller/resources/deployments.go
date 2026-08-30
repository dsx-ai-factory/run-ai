package resources

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	outils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	anyworkload_controller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/anyworkload-controller"
	runv1_internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
		name                = d.operandName
		deploymentNamespace = d.runaiConfig.GetNamespace()

		internal = d.runaiConfig.Spec.Internal
		config   = d.runaiConfig.Spec.AnyworkloadController
		image    = config.Image
	)
	deployment := &appsv1.Deployment{}

	// Get the existing deployment if it exists to consume any cluster-set values
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

	deployment.Spec.Replicas = ptr.To(int32(1))
	deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": name,
		},
	}

	deployment.Spec.Template.Name = name
	deployment.Spec.Template.Labels = outils.InitMapIfNil(deployment.Spec.Template.Labels)
	deployment.Spec.Template.Labels["app"] = name
	deployment.Spec.Template.Labels["run.ai/legacy-ownership-webhook-serving"] = "true"
	deployment.Spec.Template.Labels["run.ai/workloadrunner-ownership-webhook-serving"] = "v1"

	webhookCertVolumeMounts := []v1.VolumeMount{
		{
			Name:      "webhook-certs",
			MountPath: "/etc/webhook/certs",
			ReadOnly:  true,
		},
	}
	allVolumeMounts := append(outils.GlobalVolumeMounts(d.runaiConfig), webhookCertVolumeMounts...)

	webhookCertVolumes := []v1.Volume{
		{
			Name: "webhook-certs",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: "runai-anyworkload-controller-tls-secret",
				},
			},
		},
	}
	allVolumes := append(outils.GlobalVolumes(d.runaiConfig), webhookCertVolumes...)
	envVars, err := d.buildEnvVars(config, internal)
	if err != nil {
		return nil, fmt.Errorf("build anyworkload-controller environment: %w", err)
	}

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
				VolumeMounts:    allVolumeMounts,
				Env:             envVars,
				Ports: []v1.ContainerPort{
					{
						ContainerPort: int32(*config.Metrics.TargetPort),
						Name:          "metrics",
					},
				},
				Resources: v1.ResourceRequirements(*config.Resources),
			},
		},
		ImagePullSecrets: outils.GetGlobalImagePullSecrets(d.runaiConfig),
		Volumes:          allVolumes,
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
			condition.Status == v1.ConditionStatus(metav1.ConditionTrue) {
			return nil
		}
	}

	return fmt.Errorf("deployment [%s] has not reached the Available condition", deployment.GetName())
}

func (d *DeploymentResource) Enabled() bool {
	return true
}

func (d *DeploymentResource) buildEnvVars(config *anyworkload_controller.AnyworkloadController, internal *runv1_internal.Internal) ([]v1.EnvVar, error) {
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
			Name:  "ENABLE_MNNVL_SUPPORT",
			Value: strconv.FormatBool(*config.GPUNetworkAccelerationEnabled),
		},
		{
			Name:  "ENABLE_ELASTIC_COMPUTE_DOMAIN",
			Value: strconv.FormatBool(*config.EnableElasticComputeDomain),
		},
		{
			Name:  "ENABLE_STORAGE_PLUGIN",
			Value: strconv.FormatBool(*config.StoragePluginEnabled),
		},
		{
			Name:  "METRICS_PORT",
			Value: strconv.Itoa(*config.Metrics.TargetPort),
		},
		{
			Name:  "MAX_MUTATION_ROUNDS",
			Value: strconv.Itoa(*config.MaxMutationRounds),
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
		{
			Name:  "GODEBUG",
			Value: outils.FipsGodebugValue(d.runaiConfig, "off"),
		},
	}

	if internal != nil && len(internal.Global.NMC.AllowedServiceAccounts) > 0 {
		envVars = append(envVars, v1.EnvVar{
			Name:  "RESTART_ALLOWED_SERVICE_ACCOUNTS",
			Value: strings.Join(internal.Global.NMC.AllowedServiceAccounts, ","),
		})
	}

	if internal != nil && internal.Global.ControlPlane.Config.GetResearcherService().OauthProxy.GetOidcIssuerUrl() != "" {
		envVars = append(envVars, v1.EnvVar{
			Name:  "OAUTH2_PROXY_OIDC_ISSUER_URL",
			Value: internal.Global.ControlPlane.Config.GetResearcherService().OauthProxy.GetOidcIssuerUrl(),
		})
	}

	if config.RunaiOauth2ProxyImage != nil {
		envVars = append(envVars, v1.EnvVar{
			Name: "RUNAI_OAUTH2_PROXY_IMAGE",
			Value: fmt.Sprintf("%s/%s:%s",
				*config.RunaiOauth2ProxyImage.Registry,
				*config.RunaiOauth2ProxyImage.Name,
				*config.RunaiOauth2ProxyImage.Tag),
		})
	}

	if config.RunaiOauth2ProxyRedisFQDN != nil {
		envVars = append(envVars, v1.EnvVar{
			Name:  "RUNAI_OAUTH2_PROXY_REDIS_FQDN",
			Value: *config.RunaiOauth2ProxyRedisFQDN,
		})
	}

	if config.RunaiOauth2ProxyWorkloadAuthSecretName != nil {
		envVars = append(envVars, v1.EnvVar{
			Name:  "RUNAI_OAUTH2_PROXY_WORKLOAD_AUTH_SECRET_NAME",
			Value: *config.RunaiOauth2ProxyWorkloadAuthSecretName,
		})
	}

	if config.RunaiOauth2ProxySessionStoreType != nil {
		envVars = append(envVars, v1.EnvVar{
			Name:  "RUNAI_OAUTH2_PROXY_SESSION_STORE_TYPE",
			Value: *config.RunaiOauth2ProxySessionStoreType,
		})
	}

	if config.RunaiOauth2ProxyUsernameLoaderImage != nil {
		envVars = append(envVars, v1.EnvVar{
			Name: "RUNAI_OAUTH2_PROXY_USERNAME_LOADER_IMAGE",
			Value: fmt.Sprintf("%s/%s:%s",
				*config.RunaiOauth2ProxyUsernameLoaderImage.Registry,
				*config.RunaiOauth2ProxyUsernameLoaderImage.Name,
				*config.RunaiOauth2ProxyUsernameLoaderImage.Tag),
		})
	}

	if d.runaiConfig.Spec.Global != nil && d.runaiConfig.Spec.Global.Ingress != nil {
		if d.runaiConfig.Spec.Global.Ingress.Class != "" {
			envVars = append(envVars, v1.EnvVar{
				Name:  "INGRESS_CLASS",
				Value: d.runaiConfig.Spec.Global.Ingress.Class,
			})
		}
	}

	if wc := d.runaiConfig.Spec.WorkloadController; wc != nil {
		envVars = append(envVars, v1.EnvVar{
			Name:  "OPENSHIFT_ROUTE_TERMINATION_POLICY",
			Value: string(wc.OpenshiftRouteConfiguration.GetTerminationPolicy()),
		})
	}

	if err := outils.AddWorkloadInjectedContainerResourcesEnvVarsIfConfigured(d.runaiConfig.Spec.Global, &envVars); err != nil {
		return nil, err
	}

	return envVars, nil
}
