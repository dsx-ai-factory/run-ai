package handlers

import (
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	runaiBackendClient "github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/metrics"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	schedulingv1 "k8s.io/api/scheduling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PriorityClassesHandler struct {
	*CommonHandler
	lastUpdated time.Time
}

type PriorityClassObjectInspector bool

func (p PriorityClassObjectInspector) DeepCopySpec(from, to client.Object) {
	fromPC, ok := from.(*schedulingv1.PriorityClass)
	if !ok {
		// shouldn't happen because we convert objects to the correct type
		return
	}
	toPC, ok := to.(*schedulingv1.PriorityClass)
	if !ok {
		return
	}

	toPC.Value = fromPC.Value
	toPC.PreemptionPolicy = fromPC.PreemptionPolicy
	id, found := fromPC.Labels[constants.PriorityClassIdLabel]
	if found && id != "" {
		toPC.Labels[constants.PriorityClassIdLabel] = id
	}
}

func (p PriorityClassObjectInspector) IsUpToDate(left, right client.Object) bool {
	leftPC, ok := left.(*schedulingv1.PriorityClass)
	if !ok {
		// shouldn't happen because we convert objects to the correct type
		return false
	}
	rightPC, ok := right.(*schedulingv1.PriorityClass)
	if !ok {
		return false
	}

	leftPCLabels := leftPC.Labels
	rightPCLabels := rightPC.Labels

	return leftPC.Value == rightPC.Value &&
		leftPCLabels[constants.PriorityClassSourceLabel] == rightPCLabels[constants.PriorityClassSourceLabel] &&
		leftPCLabels[constants.PriorityClassIdLabel] == rightPCLabels[constants.PriorityClassIdLabel]
}

func NewPriorityClassesHandler(runaiBackendClient *runaiBackendClient.BackendClient, runaiClient client.Client,
	cachedClient client.Reader) (ResourceHandler, error) {
	commonHandler, err := NewCommonHandler(runaiBackendClient, runaiClient, cachedClient,
		&schedulingv1.PriorityClass{}, PriorityClassObjectInspector(true), HandlingOptions{})
	if err != nil {
		return nil, err
	}

	return &PriorityClassesHandler{
		CommonHandler: commonHandler,
		lastUpdated:   time.Time{},
	}, nil
}

func (h *PriorityClassesHandler) Handle() error {
	log.Debug().Msgf("Starting to sync priority classes")

	backendPC, err := h.getBackendPC()
	if err != nil {
		return err
	}

	priorityClassesUpsert, priorityClassesDelete, err := h.getClusterObjects(backendPC)

	metrics.GetMetricsExporter().AddResourceAction(h.MetricLabel(), metrics.ActionCreateOrUpdate, len(priorityClassesUpsert))
	metrics.GetMetricsExporter().AddResourceAction(h.MetricLabel(), metrics.ActionDelete, len(priorityClassesDelete))

	_, innerErr := h.DeleteObjectsById(priorityClassesDelete, constants.PriorityClassIdLabel)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
	}

	_, _, innerErr = h.CreateOrUpdateObjectsIgnoreConflicts(priorityClassesUpsert, constants.PriorityClassSourceLabel)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
	}

	if err != nil {
		log.Error().Msgf("Failed to handle prioirity classes, error: %s", err.Error())
		return err
	}

	log.Debug().Msgf("Successfully handled prioirity classes")
	return err
}

func (h *PriorityClassesHandler) MetricLabel() string {
	return "priority_classes"
}

func (h *PriorityClassesHandler) getBackendPC() ([]openapi.SyncPriority, error) {
	log.Debug().Msgf("Getting priority classes from workloads service")

	priorityClassRequest := h.backendClient.WorkloadsClient.WorkloadPropertiesAPI.
		GetSyncWorkloadPriorities(h.backendClient.Context()).ClusterId(config.Config().ClusterUuid)
	backendPriorities, response, err := priorityClassRequest.Execute()
	err = h.backendClient.HandleHttpResponse(response, err, "get priority classes")
	if err != nil {
		log.Error().Msgf("Failed to get priority classes: err: %v", err)
		return []openapi.SyncPriority{}, err
	}

	return backendPriorities, nil
}

func (h *PriorityClassesHandler) getClusterObjects(pcs []openapi.SyncPriority) ([]client.Object, map[string]types.NamespacedName, error) {
	var err error
	priorityClassesUpsert := []client.Object{}
	priorityClassesDelete := map[string]types.NamespacedName{}

	for _, pc := range pcs {
		if pc.DeletedAt.IsSet() && pc.DeletedAt.Get() != nil {
			priorityClassesDelete[pc.Id] = types.NamespacedName{Name: pc.Name, Namespace: ""}
		} else {
			labels := map[string]string{
				constants.PriorityClassIdLabel: pc.Id,
			}
			if pc.Source != nil {
				labels[constants.PriorityClassSourceLabel] = *pc.Source
			}
			priorityClassesUpsert = append(priorityClassesUpsert, &schedulingv1.PriorityClass{
				ObjectMeta: metav1.ObjectMeta{
					Name:   pc.Name,
					Labels: labels,
				},
				Value: pc.Value,
			})
		}
	}

	return priorityClassesUpsert, priorityClassesDelete, err
}
