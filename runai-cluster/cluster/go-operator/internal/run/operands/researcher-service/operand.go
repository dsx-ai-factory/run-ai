package researcher_service

import (
	"context"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/researcher-service/resources"
	kc "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
)

const (
	name = "researcher-service"
)

//+kubebuilder:rbac:namespace=runai,groups=apps,resources=deployments,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups=networking.k8s.io,resources=ingress,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups=route.openshift.io,resources=routes,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups=gateway.networking.k8s.io,resources=httproutes,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups=gateway.networking.k8s.io,resources=httproutes/status,verbs=get
//+kubebuilder:rbac:namespace=runai,groups="",resources=serviceaccounts,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="",resources=service,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="autoscaling.k8s.io",resources=verticalpodautoscalers,verbs=get;create;update;watch;patch;delete;list

type operand struct {
	runaiConfig *runv1.RunaiConfig
	resources   []operands.Resource
}

func (o *operand) Name() string {
	return name
}

func (o *operand) Resources() []operands.Resource {
	return o.resources
}

func (o *operand) Status(ctx context.Context) (operands.Status, error) {
	return utils.Status(ctx, o.resources)
}

func (o *operand) Enabled() bool {
	return *o.runaiConfig.Spec.ResearcherService.Enabled && !o.runaiConfig.Spec.Internal.Development.Headless
}

func Operand(k8sReader kc.HybridClient, runaiConfig *runv1.RunaiConfig) operands.Operand {
	return &operand{
		runaiConfig: runaiConfig,
		resources: []operands.Resource{
			resources.NewDeploymentResource(name, k8sReader, runaiConfig),
			resources.NewIngressResource(name, k8sReader, runaiConfig),
			resources.NewCliRouteResource(name, k8sReader, runaiConfig),
			resources.NewResearcherRouteResource(name, k8sReader, runaiConfig),
			resources.NewHTTPRouteResource(name, k8sReader, runaiConfig),
			resources.NewServiceResource(name, k8sReader, runaiConfig),
			resources.NewServiceAccountResource(name, k8sReader, runaiConfig),
			resources.NewVPAResource(name, k8sReader, runaiConfig),
		},
	}
}
