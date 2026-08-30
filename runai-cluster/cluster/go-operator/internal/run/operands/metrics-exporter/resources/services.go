package resources

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
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
		appName   = DeploymentName
		namespace = s.runaiConfig.GetNamespace()
		config    = s.runaiConfig.Spec.MetricsExporter
	)

	service := &v1.Service{}

	err := s.k8sReader.Get(ctx, types.NamespacedName{
		Name:      DeploymentName,
		Namespace: namespace,
	}, service)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	service.ObjectMeta.Name = DeploymentName
	service.ObjectMeta.Namespace = namespace
	service.ObjectMeta.Labels = map[string]string{
		"app":      appName,
		"jobLabel": DeploymentName,
	}
	service.ObjectMeta.Annotations = map[string]string{
		"prometheus.io/scrape": "true",
	}

	service.Spec.ClusterIP = "None"
	service.Spec.Ports = []v1.ServicePort{
		{
			Name:       "http-metrics",
			Port:       int32(*config.Service.Port),
			Protocol:   v1.ProtocolTCP,
			TargetPort: intstr.FromInt32(int32(*config.Service.TargetPort)),
		},
	}
	service.Spec.Selector = map[string]string{
		"app": appName,
	}
	service.Spec.SessionAffinity = v1.ServiceAffinityNone
	service.Spec.Type = config.Service.Type

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
