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
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	outils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
)

const (
	DeploymentName = "runai-job-controller"
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
	config := d.runaiConfig.Spec.JobController
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

	if deployment.ObjectMeta.Labels == nil {
		deployment.ObjectMeta.Labels = map[string]string{}
	}

	deployment.ObjectMeta.Labels["app"] = DeploymentName

	deployment.Spec.Replicas = ptr.To[int32](1)
	deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": DeploymentName,
		},
	}

	var envVars = []v1.EnvVar{
		{Name: "GODEBUG", Value: outils.FipsGodebugValue(d.runaiConfig, "off")},
	}

	if config.K8sClientBurst != nil {
		envVars = append(envVars, v1.EnvVar{Name: "K8S_CLIENT_BURST", Value: *config.K8sClientBurst})
	}

	if config.K8sClientQPS != nil {
		envVars = append(envVars, v1.EnvVar{Name: "K8S_CLIENT_QPS", Value: *config.K8sClientQPS})
	}

	deployment.Spec.Template.Name = DeploymentName
	deployment.Spec.Template.Labels = outils.InitMapIfNil(deployment.Spec.Template.Labels)
	deployment.Spec.Template.Labels["app"] = DeploymentName
	deployment.Spec.Template.Spec = v1.PodSpec{
		Affinity: outils.MergeAffinities(config.Affinity, d.runaiConfig.Spec.Global.Affinity,
			deployment.Spec.Selector.MatchLabels, d.runaiConfig.Spec.Global.RequireDefaultPodAntiAffinity),
		Tolerations:        outils.GetTolerations(d.runaiConfig),
		ServiceAccountName: DeploymentName,
		Containers: []v1.Container{
			{
				Name:            DeploymentName,
				SecurityContext: outils.GlobalSecurityContext(d.runaiConfig),
				VolumeMounts:    outils.GlobalVolumeMounts(d.runaiConfig),
				Image:           fmt.Sprintf("%s/%s:%s", *image.Registry, *image.Name, *image.Tag),
				ImagePullPolicy: image.PullPolicy,
				Args: []string{
					"--logtostderr",
					"--v", strconv.Itoa(*config.Args.Verbosity),
					"--prom-service-url=" + d.runaiConfig.Spec.Global.PrometheusService.Address,
				},
				Env:       envVars,
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
