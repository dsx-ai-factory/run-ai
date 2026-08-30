package resources

import (
	"context"
	"fmt"

	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	runv1_internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ServiceResource struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runv1.RunaiConfig
}

func NewServiceResource(operandName string, k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) *ServiceResource {
	return &ServiceResource{
		operandName: operandName,
		k8sReader:   k8sReader,
		runaiConfig: runaiConfig,
	}
}

func (s *ServiceResource) OperandName() string {
	return s.operandName
}

func (s *ServiceResource) Object(ctx context.Context) (client.Object, error) {
	var (
		namespace = s.runaiConfig.Namespace
		config    = s.runaiConfig.Spec.PodGroupAssigner
	)

	service := &v1.Service{}

	err := s.k8sReader.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, service)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	service.ObjectMeta.Name = name
	service.ObjectMeta.Namespace = namespace

	if s.runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor == runv1_internal.KubernetesFlavorOpenShift {
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

func (s *ServiceResource) Status(ctx context.Context) error {
	resource, err := s.Object(ctx)
	if err != nil {
		return err
	}

	if resource.GetUID() == "" {
		return fmt.Errorf("service [%s] is missing", resource.GetName())
	}

	return nil
}

func (s *ServiceResource) Enabled() bool {
	return true
}
