package inference

import (
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/builders"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/controller"
	inferenceaggregator "github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/status/aggregator"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/status/issue"
	statusupdater "github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/status/updater"
	utilsconverter "github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/pkg/utils/converter"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/common"
	workloadCommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/managers"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/status/update"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WorkloadControllerFactory struct {
	Client                   client.Client
	APIReadClient            client.Reader
	ComputeBuilderCreationFn func(client.Client) common.ComputeBuilder
	IsOpenshift              bool
	WorkloadType             workloadCommon.WorkloadType
}

func NewInferenceWorkloadControllerFactory(k8sClient client.Client, apiReadClient client.Reader, isOpenshift bool) *WorkloadControllerFactory {
	return &WorkloadControllerFactory{
		Client:                   k8sClient,
		APIReadClient:            apiReadClient,
		IsOpenshift:              isOpenshift,
		WorkloadType:             controller.TypeInferenceWorkload,
		ComputeBuilderCreationFn: builders.NewKnativeBuilder,
	}
}

func (wcf *WorkloadControllerFactory) CreateController() (*Controller, error) {
	knativeBuilder := wcf.ComputeBuilderCreationFn(wcf.Client)
	serviceBuilder := common.NewServiceBuilder(wcf.Client)
	externalUrlBuilder := common.NewExternalsUrlsBuilder(wcf.Client, wcf.IsOpenshift)
	pvcBuilder := common.NewPvcBuilder(wcf.Client)
	credentialsBuilder := common.NewCredentialsBuilder(wcf.Client)
	commonResourceBuilder := common.NewShardResourcesBuilder(serviceBuilder, externalUrlBuilder, pvcBuilder, credentialsBuilder)

	statusUpdater := statusupdater.NewInferenceWorkloadUpdater(wcf.Client, controller.TypeInferenceWorkload, utilsconverter.NewInferenceWorkloadConvertor())

	statusAggregator := inferenceaggregator.NewStatusCalculator()
	statusManger := update.NewSimpleStatusUpdaterManager(statusUpdater)
	k8sManager := managers.NewResourceManager(wcf.Client, wcf.APIReadClient, &issue.InferenceResourceGrouper{})
	return NewController(wcf.Client,
		knativeBuilder,
		commonResourceBuilder,
		statusManger, nil,
		statusAggregator, k8sManager,
		managers.ResourceOptions{}), nil
}
