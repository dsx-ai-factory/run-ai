package resources

import (
	"context"
	"fmt"
	"strconv"

	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	outils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	name = operands.Runai{{cookiecutter.__app_name_camel_case_upper}}Name
)

func DeploymentForRunaiConfig(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runaiv1.RunaiConfig) (client.Object, error) {
	var (
		deploymentNamespace = runaiConfig.GetNamespace()

		config = runaiConfig.Spec.{{cookiecutter.__app_name_camel_case_upper}}
		image  = config.Image
	)
	deployment := &appsv1.Deployment{}

	// Get the existing deployment if it exists to consume any cluster-set values
	err := k8sReader.Get(ctx, types.NamespacedName{
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

	deployment.Spec.Replicas = pointer.Int32(1)
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
		Affinity:           outils.GetAffinity(runaiConfig),
		Tolerations:        outils.GetTolerations(runaiConfig),
		Containers: []v1.Container{
			{
				Name:            name,
				SecurityContext: outils.GlobalSecurityContext(runaiConfig),
				Image:           fmt.Sprintf("%s/%s:%s", *image.Registry, *image.Name, *image.Tag),
				ImagePullPolicy: image.PullPolicy,
				VolumeMounts:    outils.GlobalVolumeMounts(runaiConfig),
				Env: []v1.EnvVar{
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
				},
				Resources: v1.ResourceRequirements(*config.Resources),
			},
		},
		ImagePullSecrets: outils.GetGlobalImagePullSecrets(runaiConfig),
		Volumes: outils.GlobalVolumes(runaiConfig),
	}

	return deployment, nil
}

func DeploymentStatus(ctx context.Context,
	k8sReader client.Reader, runaiConfig *runaiv1.RunaiConfig) (operands.Status, error) {
	resource, err := DeploymentForRunaiConfig(ctx, k8sReader, runaiConfig)
	if err != nil {
		return operands.NotReadyStatus(), err
	}

	deployment := resource.(*appsv1.Deployment)

	if deployment.GetUID() == "" {
		return operands.NotReadyStatus(fmt.Sprintf("deployment [%s] is missing",
			deployment.GetName())), nil
	}

	deploymentStatus := deployment.Status
	if len(deploymentStatus.Conditions) == 0 {
		return operands.NotReadyStatus(fmt.Sprintf("deployment [%s] has no conditions",
			deployment.GetName())), nil
	}

	conditions := deploymentStatus.Conditions
	for _, condition := range conditions {
		if condition.Type == appsv1.DeploymentAvailable &&
			condition.Status == v1.ConditionStatus(metav1.ConditionTrue) {
			return operands.ReadyStatus(), nil
		}
	}

	return operands.NotReadyStatus(fmt.Sprintf("deployment [%s] has not reached the Available condition",
		deployment.GetName())), nil
}
