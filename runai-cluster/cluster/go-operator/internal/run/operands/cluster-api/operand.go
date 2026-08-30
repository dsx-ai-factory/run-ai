package cluster_api

import (
	"context"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/cluster-api/resources"
	kc "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
)

const (
	Name = operands.RunaiClusterApiName
)

//+kubebuilder:rbac:namespace=runai,groups=apps,resources=deployments,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="",resources=serviceaccounts,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups=gateway.networking.k8s.io,resources=httproutes,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups=gateway.networking.k8s.io,resources=httproutes/status,verbs=get
//+kubebuilder:rbac:namespace=runai,groups="autoscaling.k8s.io",resources=verticalpodautoscalers,verbs=get;create;update;watch;patch;delete;list

type operand struct {
	runaiConfig *runv1.RunaiConfig
	resources   []operands.Resource
}

func (o *operand) Name() string {
	return Name
}

func (o *operand) Resources() []operands.Resource {
	return o.resources
}

func (o *operand) Status(ctx context.Context) (operands.Status, error) {
	return utils.Status(ctx, o.resources)
}

func (o *operand) Enabled() bool {
	return *o.runaiConfig.Spec.ClusterApi.Enabled && !o.runaiConfig.Spec.Internal.Development.Headless
}

func Operand(k8sReader kc.HybridClient, runaiConfig *runv1.RunaiConfig) operands.Operand {
	return &operand{
		runaiConfig: runaiConfig,
		resources: []operands.Resource{
			resources.NewDeploymentResource(Name, k8sReader, runaiConfig),
			resources.NewServiceAccountResource(Name, k8sReader, runaiConfig),
			resources.NewRouteResource(Name, k8sReader, runaiConfig),
			resources.NewIngressResource(Name, k8sReader, runaiConfig),
			resources.NewHTTPRouteResource(Name, k8sReader, runaiConfig),
			resources.NewServiceResource(Name, k8sReader, runaiConfig),
			resources.NewServiceMonitorResource(Name, k8sReader, runaiConfig),
			resources.NewVPAResource(Name, k8sReader, runaiConfig),
		},
	}
}
