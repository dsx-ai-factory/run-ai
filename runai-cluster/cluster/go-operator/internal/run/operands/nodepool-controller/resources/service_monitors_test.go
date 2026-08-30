package resources

import (
	"context"
	"testing"

	kaiconstants "github.com/kai-scheduler/KAI-scheduler/pkg/common/constants"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestNodepoolControllerServiceMonitorIsScrapedByRunaiPrometheus(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, monitoringv1.AddToScheme(scheme))

	runaiConfig := &runv1.RunaiConfig{}
	runaiConfig.Name = runv1.DefaultName
	runaiConfig.Namespace = runv1.DefaultNamespace
	runaiConfig.SetDefaultsWhereNeeded()

	resource := NewServiceMonitorResource(
		DeploymentName,
		fake.NewClientBuilder().WithScheme(scheme).Build(),
		runaiConfig,
	)
	object, err := resource.Object(context.Background())
	require.NoError(t, err)

	serviceMonitor := object.(*monitoringv1.ServiceMonitor)
	require.Equal(t, DeploymentName, serviceMonitor.Name)
	require.Equal(t, DeploymentName, serviceMonitor.Labels["app"])
	// Must NOT carry the kai.scheduler/accounting label: the runai Prometheus
	// excludes accounting-labeled monitors, and runai_node_nodepool must be
	// scraped by the runai Prometheus so its recording rules can consume it.
	require.NotContains(t, serviceMonitor.Labels, kaiconstants.DefaultAccountingLabelKey)
	require.Equal(t, runaiConfig.Spec.NodePoolController.Metrics.PortName, serviceMonitor.Spec.Endpoints[0].Port)
	require.Equal(t, DeploymentName, serviceMonitor.Spec.Selector.MatchLabels["app"])
}

func TestNodepoolControllerAccountingServiceMonitorIsScrapedByKaiAccountingPrometheus(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, monitoringv1.AddToScheme(scheme))

	runaiConfig := &runv1.RunaiConfig{}
	runaiConfig.Name = runv1.DefaultName
	runaiConfig.Namespace = runv1.DefaultNamespace
	runaiConfig.SetDefaultsWhereNeeded()

	resource := NewAccountingServiceMonitorResource(
		DeploymentName,
		fake.NewClientBuilder().WithScheme(scheme).Build(),
		runaiConfig,
	)
	object, err := resource.Object(context.Background())
	require.NoError(t, err)

	serviceMonitor := object.(*monitoringv1.ServiceMonitor)
	// Distinct object name from the primary ServiceMonitor so both can coexist.
	require.Equal(t, accountingServiceMonitorName, serviceMonitor.Name)
	require.NotEqual(t, DeploymentName, serviceMonitor.Name)
	require.Equal(t, DeploymentName, serviceMonitor.Labels["app"])
	// Carries the kai.scheduler/accounting label so the KAI accounting Prometheus
	// scrapes it (In "true"), while the runai Prometheus ignores it (NotIn "true").
	// This routes runai_node_nodepool to the accounting Prometheus for
	// runai-accounting / resources_per_nodepool.
	require.Equal(t,
		kaiconstants.DefaultAccountingLabelValue,
		serviceMonitor.Labels[kaiconstants.DefaultAccountingLabelKey],
	)
	// Scrapes the same nodepool-controller endpoint as the primary ServiceMonitor.
	require.Equal(t, runaiConfig.Spec.NodePoolController.Metrics.PortName, serviceMonitor.Spec.Endpoints[0].Port)
	require.Equal(t, DeploymentName, serviceMonitor.Spec.Selector.MatchLabels["app"])
}
