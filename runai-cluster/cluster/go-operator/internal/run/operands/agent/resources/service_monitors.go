package resources

import (
	"context"
	"fmt"

	monitorv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ServiceMonitorResource struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runv1.RunaiConfig
}

func NewServiceMonitorResource(operandName string, k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) *ServiceMonitorResource {
	return &ServiceMonitorResource{operandName: operandName, k8sReader: k8sReader, runaiConfig: runaiConfig}
}

func (s *ServiceMonitorResource) OperandName() string { return s.operandName }

func (s *ServiceMonitorResource) Object(ctx context.Context) (client.Object, error) {
	namespace := s.runaiConfig.GetNamespace()
	serviceMonitor := &monitorv1.ServiceMonitor{}
	err := s.k8sReader.Get(ctx, types.NamespacedName{Name: DeploymentName, Namespace: namespace}, serviceMonitor)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}
	serviceMonitor.ObjectMeta.Name = DeploymentName
	serviceMonitor.ObjectMeta.Namespace = namespace
	serviceMonitor.ObjectMeta.Labels = map[string]string{"app": DeploymentName}
	serviceMonitor.Spec.JobLabel = DeploymentName
	serviceMonitor.Spec.NamespaceSelector = monitorv1.NamespaceSelector{MatchNames: []string{namespace}}
	serviceMonitor.Spec.Selector = v1.LabelSelector{MatchLabels: map[string]string{"app": DeploymentName}}
	serviceMonitor.Spec.Endpoints = []monitorv1.Endpoint{
		{
			Port:            s.runaiConfig.Spec.Agent.Metrics.PortName,
			BearerTokenFile: "/var/run/secrets/kubernetes.io/serviceaccount/token",
			MetricRelabelConfigs: []monitorv1.RelabelConfig{
				{Replacement: ptr.To("stats"), TargetLabel: "type", Action: "replace"},
			},
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

func (s *ServiceMonitorResource) Enabled() bool { return true }
