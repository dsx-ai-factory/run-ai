package resources

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	runaiv1_internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ServiceForRunaiConfig(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runv1.RunaiConfig) (client.Object, error) {
	var (
		namespace = runaiConfig.Namespace
		config    = runaiConfig.Spec.AccessRuleController
	)

	service := &v1.Service{}

	// Get the existing service if it exists to consume any cluster-set values
	err := k8sReader.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, service)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	service.ObjectMeta.Name = name
	service.ObjectMeta.Namespace = namespace

	if runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor == runaiv1_internal.KubernetesFlavorOpenShift {
		if service.ObjectMeta.Annotations == nil {
			service.ObjectMeta.Annotations = map[string]string{}
		}

		service.ObjectMeta.Annotations["service.beta.openshift.io/serving-cert-secret-name"] = tlsSecretName
	}

	if service.ObjectMeta.Labels == nil {
		service.ObjectMeta.Labels = map[string]string{}
	}

	service.ObjectMeta.Labels["app"] = name

	service.Spec.Selector = map[string]string{
		"app": name,
	}
	service.Spec.Ports = []v1.ServicePort{
		{
			Protocol:   v1.ProtocolTCP,
			Port:       int32(*config.Service.Webhook.Port),
			TargetPort: intstr.FromInt32(int32(*config.Service.Webhook.TargetPort)),
		},
	}
	return service, nil
}

func ServiceStatus(ctx context.Context,
	k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) (operands.Status, error) {
	service, err := ServiceForRunaiConfig(ctx, k8sReader, runaiConfig)
	if err != nil {
		return operands.NotReadyStatus(), err
	}

	if service.GetUID() == "" {
		return operands.NotReadyStatus(fmt.Sprintf("service [%s] is missing",
			service.GetName())), nil
	}

	return operands.ReadyStatus(), nil
}
