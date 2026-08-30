package resources

import (
	"context"
	"fmt"
	"strconv"

	clustercommonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
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
	name          = "pod-group-assigner"
	certsDir      = "/etc/webhook/certs"
	tlsSecretName = "runai-pod-group-assigner-tls-secret"
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
		config              = d.runaiConfig.Spec.PodGroupAssigner
		image               = config.Image
	)
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

	deployment.Labels = outils.InitMapIfNil(deployment.Labels)
	deployment.Labels["app"] = name

	deployment.Spec.Replicas = config.Replicas
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
						MountPath: certsDir,
						Name:      "webhook-certs",
						ReadOnly:  true,
					},
				),
				Env: []v1.EnvVar{
					{
						Name:  "GODEBUG",
						Value: outils.FipsGodebugValue(d.runaiConfig, "off"),
					},
				},
				Args: func() []string {
					args := []string{
						// runai overrides: binary defaults are KAI-style; preserve runai vocabulary explicitly.
						// queue-label-key and namespace-project-label-key MUST share runai/queue to preserve the
						// historical overload of RunaiQueueLabel on both workload and namespace labels.
						"--nodepool-label-key", v1alpha1.NodePoolNameLabel,
						"--queue-label-key", clustercommonconstants.RunaiQueueLabel,
						"--namespace-project-label-key", clustercommonconstants.RunaiQueueLabel,
						"--project-label-key", clustercommonconstants.RunaiProjectLabel,
						"--unexisting-nodepool-sentinel", v1alpha1.UnexistingRunaiNodePool,
						"--default-nodepool-name", v1alpha1.DefaultNodePoolName,
						"--annotation-nodepools-key", commonconstants.AnnotationNodepools,
						"--enforce-scheduler-annotation-key", clustercommonconstants.EnforceSchedulerAnnotationName,
						"--scheduler-name", commonconstants.RunaiSchedulerName,
						"--webhook-port",
						strconv.Itoa(*config.Service.Webhook.TargetPort),
						"--qps",
						strconv.Itoa(*config.K8sClientConfig.QPS),
						"--burst",
						strconv.Itoa(*config.K8sClientConfig.Burst),
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
		Volumes: append(outils.GlobalVolumes(d.runaiConfig),
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
