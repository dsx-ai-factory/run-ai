package resources

import (
	"context"
	"fmt"

	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	runaiv1_internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ServiceResource struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runaiv1.RunaiConfig
}

func NewServiceResource(operandName string, k8sReader client.Reader, runaiConfig *runaiv1.RunaiConfig) *ServiceResource {
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
		name             = s.operandName
		serviceNamespace = s.runaiConfig.GetNamespace()
		config           = s.runaiConfig.Spec.AnyworkloadController
	)

	service := &v1.Service{}
	err := s.k8sReader.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: serviceNamespace,
	}, service)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	service.ObjectMeta.Name = name
	service.ObjectMeta.Namespace = serviceNamespace

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
			Name:       webhookPortName,
			Port:       443,
			TargetPort: intstr.FromInt32(9443),
		},
		{
			Protocol:   v1.ProtocolTCP,
			Name:       metricPortName,
			Port:       int32(*config.Metrics.Port),
			TargetPort: intstr.FromInt32(int32(*config.Metrics.TargetPort)),
		},
	}

	if s.runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor == runaiv1_internal.KubernetesFlavorOpenShift {
		if service.ObjectMeta.Annotations == nil {
			service.ObjectMeta.Annotations = map[string]string{}
		}
		service.ObjectMeta.Annotations["service.beta.openshift.io/serving-cert-secret-name"] =
			"runai-anyworkload-controller-tls-secret"
	}

	return service, nil
}

func (s *ServiceResource) Status(ctx context.Context) error {
	resource, err := s.Object(ctx)
	if err != nil {
		return err
	}

	service := resource.(*v1.Service)
	if service.GetUID() == "" {
		return fmt.Errorf("service [%s] is missing", service.GetName())
	}

	return nil
}

func (s *ServiceResource) Enabled() bool {
	return true
}
