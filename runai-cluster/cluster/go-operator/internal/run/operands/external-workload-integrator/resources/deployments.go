package resources

import (
	"context"
	"fmt"
	"strconv"

	"k8s.io/utils/ptr"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	outils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	runv1_internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
)

const (
	DeploymentName = operands.RunaiExternalWorkloadIntegratorName
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
	config := d.runaiConfig.Spec.ExternalWorkloadIntegrator
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
	deployment.Labels["app"] = DeploymentName

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
				VolumeMounts:    outils.GlobalVolumeMounts(d.runaiConfig),
				Env:             handleEnvVars(d.runaiConfig),
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

func handleEnvVars(runaiConfig *runaiv1.RunaiConfig) []v1.EnvVar {
	config := runaiConfig.Spec.ExternalWorkloadIntegrator
	envVars := []v1.EnvVar{
		{
			Name:  "IS_OPENSHIFT",
			Value: strconv.FormatBool(runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor == runv1_internal.KubernetesFlavorOpenShift),
		},
		{
			Name:  "METRICS_PORT",
			Value: strconv.Itoa(int(config.Metrics.Port)),
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
			Name:  "ALLOW_NON_SUPPORTED_TYPES",
			Value: strconv.FormatBool(true),
		},
		{
			Name:  "WORKLOAD_TYPE_RBAC_HANDLER_ENABLED",
			Value: strconv.FormatBool(*config.RIDRbacConfig.Enabled),
		},
		{
			Name:  "CLUSTER_ROLE_READ_NAME_FOR_RBAC",
			Value: config.RIDRbacConfig.ClusterRoleReadName,
		},
		{
			Name:  "CLUSTER_ROLE_CRUD_NAME_FOR_RBAC",
			Value: config.RIDRbacConfig.ClusterRoleCrudName,
		},
		{
			Name:  "CLUSTER_DOMAIN",
			Value: runaiConfig.Spec.Internal.Global.ClusterURL,
		},
		{
			Name:  "LB_SERVICE_ENDPOINTS_TLS_BY_EXPLICIT_ANNOTATION",
			Value: strconv.FormatBool(*config.Network.LbServiceEndpointsTlsByExplicitAnnotation),
		},
		{Name: "GODEBUG", Value: outils.FipsGodebugValue(runaiConfig, "off")},
	}
	if ptr.Deref(config.Replicas, 1) > 1 {
		envVars = append(envVars, v1.EnvVar{
			Name:  "LEADER_ELECT",
			Value: "true",
		})
	}
	outils.AddNMCRemoteClusterAPIServerURLEnvVarIfConfigured(runaiConfig, &envVars)
	return envVars
}
