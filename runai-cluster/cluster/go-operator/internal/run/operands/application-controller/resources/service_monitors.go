package resources

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
)

func ServiceMonitorForRunaiConfig(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runv1.RunaiConfig) (client.Object, error) {
	var (
		name      = operands.RunaiApplicationControllerName
		namespace = runaiConfig.Namespace
		appName   = operands.RunaiApplicationControllerName
	)

	serviceMonitor := &monitoringv1.ServiceMonitor{}

	// Get the existing service monitor if it exists to consume any cluster-set values
	err := k8sReader.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, serviceMonitor)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	serviceMonitor.ObjectMeta.Name = name
	serviceMonitor.ObjectMeta.Namespace = namespace
	serviceMonitor.ObjectMeta.Labels = map[string]string{
		"app": appName,
	}

	serviceMonitor.Spec.JobLabel = appName
	serviceMonitor.Spec.NamespaceSelector = monitoringv1.NamespaceSelector{
		MatchNames: []string{
			"runai",
		},
	}
	serviceMonitor.Spec.Selector = metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": appName,
		},
	}
	serviceMonitor.Spec.Endpoints = []monitoringv1.Endpoint{
		{
			BearerTokenFile: "/var/run/secrets/kubernetes.io/serviceaccount/token",
			Port:            runaiConfig.Spec.ApplicationController.Service.Metrics.PortName,
			MetricRelabelConfigs: []monitoringv1.RelabelConfig{
				{
					Replacement: ptr.To("stats"),
					TargetLabel: "type",
					Action:      "replace",
				},
			},
		},
	}

	return serviceMonitor, nil
}

func ServiceMonitorStatus(ctx context.Context,
	k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) (operands.Status, error) {
	serviceMonitor, err := ServiceMonitorForRunaiConfig(ctx, k8sReader, runaiConfig)
	if err != nil {
		return operands.NotReadyStatus(), err
	}

	if serviceMonitor.GetUID() == "" {
		return operands.NotReadyStatus(fmt.Sprintf("service monitor [%s] is missing",
			serviceMonitor.GetName())), nil
	}

	return operands.ReadyStatus(), nil
}
