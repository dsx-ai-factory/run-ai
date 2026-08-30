package resources

import (
	"context"
	"fmt"
	"strconv"

	clustercommonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	controller_utils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	outils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DeploymentName = "nodepool-controller"
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
		deploymentNamespace = d.runaiConfig.GetNamespace()
		config              = d.runaiConfig.Spec.NodePoolController
		image               = config.Image
	)

	shardArgs, err := marshalShardArgs(d.runaiConfig.Spec.RunAIScheduler, d.runaiConfig.Spec.Global.NodeAffinity)
	if err != nil {
		return nil, err
	}

	dcgmExporterServiceNamespace := getDcgmExporterNamespaceForScheduler(ctx, d.k8sReader)

	deployment := &appsv1.Deployment{}

	err = d.k8sReader.Get(ctx, types.NamespacedName{
		Name:      DeploymentName,
		Namespace: deploymentNamespace,
	}, deployment)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	deployment.ObjectMeta.Name = DeploymentName
	deployment.ObjectMeta.Namespace = deploymentNamespace

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
				Env: []v1.EnvVar{
					{
						Name:  "METRICS_PORT",
						Value: strconv.Itoa(int(config.Metrics.Port)),
					},
					{
						Name:  "GODEBUG",
						Value: outils.FipsGodebugValue(d.runaiConfig, "off"),
					},
				},
				Ports: []v1.ContainerPort{
					{
						ContainerPort: config.Metrics.Port,
						Name:          config.Metrics.PortName,
					},
				},
				VolumeMounts: outils.GlobalVolumeMounts(d.runaiConfig),
				Args: func() []string {
					args := []string{
						// runai overrides: nodepool-controller binary defaults are
						// KAI-style; preserve runai vocabulary explicitly so runai
						// deployment behavior is unchanged.
						"--nodepool-label-key", v1alpha1.NodePoolNameLabel,
						"--default-nodepool-name", v1alpha1.DefaultNodePoolName,
						"--scheduler-name", commonconstants.RunaiSchedulerName,
						"--scheduler-namespace", clustercommonconstants.NamespaceRunAi,
						"--metrics-namespace", clustercommonconstants.NamespaceRunAi,
						"--finalizer-domain", clustercommonconstants.RunaiOrgDomain,
						"--scheduling-shard-args", shardArgs,
						"--managed-nodes-config-name", clustercommonconstants.ManagedNodesConfigName,
						"--excluded-nodepool-name", clustercommonconstants.ExcludedNodepoolName,
						"--to-exclude-label", clustercommonconstants.ShouldBeExcludedLabelKey,
						"--unschedulable-label", clustercommonconstants.LabelMarkedUnschedulableByNodePoolController,
						"--grove-topology-annotation", clustercommonconstants.GroveTopologyAnnotation,
						"--grove-topology-resource-version-annotation", clustercommonconstants.GroveTopologyResourceVersionAnnotation,
						"--metrics-port", strconv.Itoa(int(config.Metrics.Port)),
						"--dcgm-exporter-namespace", dcgmExporterServiceNamespace,
						"--enable-nodepool-validation-webhook=false",
					}
					if config.Args.DebugLog {
						args = append(args, "--debug")
					}

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

func getDcgmExporterNamespaceForScheduler(ctx context.Context, k8sReader client.Reader) string {
	dcgmExporterService, err := controller_utils.GetDCGMExporterService(ctx, k8sReader)
	if err != nil {
		return ""
	}

	return dcgmExporterService.Namespace
}
