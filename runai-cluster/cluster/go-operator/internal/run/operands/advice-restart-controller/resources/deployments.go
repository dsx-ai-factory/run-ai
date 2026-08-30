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
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	outils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
)

const (
	DeploymentName = operands.RunaiAdviceRestartControllerName
	tlsSecretName  = "runai-advice-restart-controller-tls-secret"
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
	config := d.runaiConfig.Spec.AdviceRestartController
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

	replicas := int32(1)
	deployment.Spec.Replicas = &replicas
	deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": DeploymentName,
		},
	}

	deployment.Spec.Template.Name = DeploymentName
	deployment.Spec.Template.Labels = outils.InitMapIfNil(deployment.Spec.Template.Labels)
	deployment.Spec.Template.Labels["app"] = DeploymentName

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
					SecretName: tlsSecretName,
				},
			},
		},
	}
	allVolumes := append(outils.GlobalVolumes(d.runaiConfig), webhookCertVolumes...)

	deployment.Spec.Template.Spec = v1.PodSpec{
		ServiceAccountName: outils.GetNMCServiceAccountNameOrDefault(d.runaiConfig, DeploymentName),
		Affinity: outils.MergeAffinities(nil, d.runaiConfig.Spec.Global.Affinity,
			deployment.Spec.Selector.MatchLabels, d.runaiConfig.Spec.Global.RequireDefaultPodAntiAffinity),
		Tolerations: outils.GetTolerations(d.runaiConfig),
		Containers: []v1.Container{
			{
				Name:            DeploymentName,
				SecurityContext: outils.GlobalSecurityContext(d.runaiConfig),
				Image:           fmt.Sprintf("%s/%s:%s", *image.Registry, *image.Name, *image.Tag),
				ImagePullPolicy: image.PullPolicy,
				VolumeMounts:    allVolumeMounts,
				Env:             handleEnvVars(d.runaiConfig),
				Resources:       v1.ResourceRequirements(*config.Resources),
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

	for _, condition := range deploymentStatus.Conditions {
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
	config := runaiConfig.Spec.AdviceRestartController
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
			Name:  "WEBHOOK_PORT",
			Value: strconv.Itoa(*config.Service.Webhook.TargetPort),
		},
		{
			Name:  "WEBHOOK_CERT_DIR",
			Value: "/etc/webhook/certs",
		},
		{
			Name:  "MODE",
			Value: *config.Mode,
		},
		{
			Name:  "RESTART_LIMIT",
			Value: strconv.Itoa(*config.RestartLimit),
		},
		{
			Name:  "MAX_RETRIES_PER_ADVICE",
			Value: strconv.Itoa(*config.MaxRetriesPerAdvice),
		},
		{
			Name:  "METRICS_BIND_ADDRESS",
			Value: fmt.Sprintf(":%d", *config.Metric.TargetPort),
		},
		{
			Name:  "GODEBUG",
			Value: outils.FipsGodebugValue(runaiConfig, "off"),
		},
	}

	outils.AddNMCRemoteClusterAPIServerURLEnvVarIfConfigured(runaiConfig, &envVars)

	return envVars
}
