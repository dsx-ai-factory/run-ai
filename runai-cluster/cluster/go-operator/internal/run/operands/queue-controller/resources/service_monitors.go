package resources

import (
	"context"
	"fmt"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
)

const (
	Name = "queue-controller"
)

type ServiceMonitorResource struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runv1.RunaiConfig
}

func NewServiceMonitorResource(operandName string, k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) *ServiceMonitorResource {
	return &ServiceMonitorResource{
		operandName: operandName,
		k8sReader:   k8sReader,
		runaiConfig: runaiConfig,
	}
}

func (s *ServiceMonitorResource) OperandName() string {
	return s.operandName
}

func (s *ServiceMonitorResource) Object(ctx context.Context) (client.Object, error) {
	var (
		namespace = s.runaiConfig.Namespace
		config    = s.runaiConfig.Spec.QueueController
	)

	serviceMonitor := &monitoringv1.ServiceMonitor{}

	err := s.k8sReader.Get(ctx, types.NamespacedName{
		Name:      Name,
		Namespace: namespace,
	}, serviceMonitor)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	serviceMonitor.ObjectMeta.Name = Name
	serviceMonitor.ObjectMeta.Namespace = namespace
	if serviceMonitor.Labels == nil {
		serviceMonitor.Labels = map[string]string{}
	}
	serviceMonitor.Labels["app"] = Name
	serviceMonitor.Spec.JobLabel = Name

	serviceMonitor.Spec.Selector = metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": Name,
		},
	}
	serviceMonitor.Spec.NamespaceSelector = monitoringv1.NamespaceSelector{
		MatchNames: []string{
			namespace,
		},
	}
	serviceMonitor.Spec.Endpoints = []monitoringv1.Endpoint{
		{
			Port:            *config.Service.Metrics.Name,
			BearerTokenFile: "/var/run/secrets/kubernetes.io/serviceaccount/token",
		},
	}

	return serviceMonitor, nil
}

func (s *ServiceMonitorResource) Status(ctx context.Context) error {
	resource, err := s.Object(ctx)
	if err != nil {
		return err
	}

	if resource.GetUID() == "" {
		return fmt.Errorf("service monitor [%s] is missing", resource.GetName())
	}

	return nil
}

func (s *ServiceMonitorResource) Enabled() bool {
	return true
}
