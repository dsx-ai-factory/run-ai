package resources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/secrets"
	outils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DeploymentName = "cluster-sync"
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
		config              = d.runaiConfig.Spec.ClusterSync
		image               = config.Image
	)

	deployment := &appsv1.Deployment{}
	err := d.k8sReader.Get(ctx, types.NamespacedName{Name: DeploymentName, Namespace: deploymentNamespace}, deployment)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	deployment.ObjectMeta.Name = DeploymentName
	deployment.ObjectMeta.Namespace = deploymentNamespace

	if d.runaiConfig.Spec.Internal.Development.Headless {
		deployment.Spec.Replicas = ptr.To[int32](0)
	} else {
		deployment.Spec.Replicas = config.Replicas
	}

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
				Args:            []string{"--log-level", strconv.Itoa(*config.LogLevel)},
				VolumeMounts:    outils.GlobalVolumeMounts(d.runaiConfig),
				Env:             handleEnvVars(d.runaiConfig),
				Ports: []v1.ContainerPort{
					{ContainerPort: int32(config.Metrics.Port), Name: config.Metrics.PortName},
				},
				Resources: v1.ResourceRequirements(*config.Resources),
			},
		},
		ImagePullSecrets: outils.GetGlobalImagePullSecrets(d.runaiConfig),
		Volumes:          outils.GlobalVolumes(d.runaiConfig),
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

func (d *DeploymentResource) Enabled() bool {
	return true
}

func handleEnvVars(runaiConfig *runaiv1.RunaiConfig) []v1.EnvVar {
	config := runaiConfig.Spec.ClusterSync
	global := runaiConfig.Spec.Internal.Global
	envVars := []v1.EnvVar{
		{Name: "METRICS_PORT", Value: strconv.Itoa(int(config.Metrics.Port))},
		{Name: "ENDPOINT", Value: global.ControlPlane.URL},
		{Name: "SSL_CERTIFICATE_VERIFICATION", Value: strconv.FormatBool(*config.SSLCertificateVerification)},
		{Name: "MAX_NUM_OF_EVENTS_PER_REQUEST", Value: strconv.Itoa(*config.MaxNumOfEventsPerRequest)},
		{Name: "MAX_NUM_OF_PODS_PER_REQUEST", Value: strconv.Itoa(*config.MaxNumOfPodsPerRequest)},
		{Name: "MAX_NUM_OF_WORKLOADS_PER_REQUEST", Value: strconv.Itoa(*config.MaxNumOfWorkloadsPerRequest)},
		{Name: "MAX_NUM_OF_DATAVOLUMES_PER_REQUEST", Value: strconv.Itoa(*config.MaxNumOfDatavolumesPerRequest)},
		{Name: "SYNC_EVENTS_INTERVAL_SECONDS", Value: strconv.Itoa(*config.SyncEventsIntervalSeconds)},
		{Name: "LIVENESS_INTERVAL_SECONDS", Value: strconv.Itoa(*config.LivenessIntervalSeconds)},
		{Name: "MAX_NUMBER_OF_RETRIES", Value: strconv.Itoa(*config.MaxNumberOfRetries)},
		{Name: "SYNC_INTERVAL_MINUTES", Value: strconv.Itoa(*config.SyncIntervalMinutes)},
		{Name: "EXISTING_SECRET_TO_SYNC_LABEL", Value: config.ExistingSecretToSyncLabel},
		{Name: "CONFIG_MAP_TO_SYNC_LABEL", Value: config.ConfigMapToSyncLabel},
		{Name: "RETRIES_INTERVAL_SECONDS", Value: strconv.Itoa(*config.RetriesIntervalSeconds)},
		{Name: "CLUSTER_DOMAIN", Value: global.ClusterURL},
		{Name: "ENABLE_PROFILER", Value: strconv.FormatBool(*config.Profiling.Enabled)},
		{Name: "PROFILER_API_PORT", Value: config.Profiling.ApiPort},
		{Name: "K8S_CLIENT_CONFIG_QPS", Value: strconv.Itoa(*config.K8sClientConfig.QPS)},
		{Name: "K8S_CLIENT_CONFIG_BURST", Value: strconv.Itoa(*config.K8sClientConfig.Burst)},
		{Name: "NVIDIA_GPU_LABEL", Value: *config.NvidiaGPULabel},
		{Name: "NODE_LABEL_NVIDIA_CONTAINER_TOOLKIT", Value: *config.NodeLabelNvidiaContainerToolkit},
		{Name: "NODE_LABEL_NVIDIA_DCGM_EXPORTER", Value: *config.NodeLabelNvidiaDcgmExporter},
		{Name: "ENABLE_NODE_NVIDIA_COMPONENTS_READINESS", Value: strconv.FormatBool(*config.EnableNodeNvidiaComponentsReadiness)},
		{Name: "RATE_LIMITER_WORKLOADS_DELAY_SECONDS", Value: strconv.Itoa(*config.RateLimiterWorkloadsDelaySeconds)},
		{Name: "RATE_LIMITER_WORKLOADS_NUM_OF_WORKERS", Value: strconv.Itoa(*config.RateLimiterWorkloadsNumOfWorkers)},
		{Name: "RATE_LIMITER_REGULAR_DELAY_SECONDS", Value: strconv.Itoa(*config.RateLimiterRegularDelaySeconds)},
		{Name: "RATE_LIMITER_REGULAR_NUM_OF_WORKERS", Value: strconv.Itoa(*config.RateLimiterRegularNumOfWorkers)},
		{Name: "SYNC_PODS_WITH_WORKLOADS", Value: strconv.FormatBool(*config.SyncPodsWithWorkloads)},
		{Name: "GODEBUG", Value: outils.FipsGodebugValue(runaiConfig, "off")},
	}
	if ptr.Deref(config.Replicas, 1) > 1 {
		envVars = append(envVars, v1.EnvVar{Name: "LEADER_ELECT", Value: "true"})
	}
	outils.AddNMCRemoteClusterAPIServerURLEnvVarIfConfigured(runaiConfig, &envVars)
	return envVars
}
