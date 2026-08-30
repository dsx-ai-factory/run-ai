package handlers

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"sync"
	"sync/atomic"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	swagger "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/metrics"
	"k8s.io/utils/ptr"

	runaiBackendClient "github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	_ "k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KartaHandler struct {
	resourceName  constants.ClusterResource
	runaiClient   client.Client
	backendClient *runaiBackendClient.BackendClient

	// syncMutex serializes all SyncWorkloadTypesExecute calls so that Keycloak
	// only sees one in-flight JWT validation for this endpoint at a time.
	syncMutex sync.Mutex
	// syncWaiting allows at most one waiter behind syncMutex for HandleSyncEvent;
	// additional full-sync callers are dropped because the in-flight or queued
	// sync subsumes their work.
	syncWaiting atomic.Bool
}

func NewKartaHandler(runaiBackendClient *runaiBackendClient.BackendClient, runaiClient client.Client) *KartaHandler {
	return &KartaHandler{
		resourceName:  constants.KartaResourceName,
		runaiClient:   runaiClient,
		backendClient: runaiBackendClient,
	}
}

func (h *KartaHandler) HandleUpdateItem(key client.ObjectKey) error {
	karta := &kartav1alpha1.Karta{}
	err := h.runaiClient.Get(context.Background(), key, karta)
	if err != nil {
		return err
	}

	return h.HandleAddUpdate(karta)
}

func (h *KartaHandler) HandleAddUpdate(karta *kartav1alpha1.Karta) error {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(h.resourceName, constants.UpsertOperation)()
	backendKarta, err := parsers.ParseBackendKarta(karta, false)
	if err != nil {
		return err
	}

	h.syncMutex.Lock()
	defer h.syncMutex.Unlock()

	request := h.backendClient.WorkloadsClient.WorkloadPropertiesAPI.
		SyncWorkloadTypes(h.backendClient.Context()).
		ClusterId(config.Config().ClusterUuid).
		WorkloadTypesForSync(swagger.WorkloadTypesForSync{
			FullSync:      ptr.To(false),
			WorkloadTypes: []swagger.WorkloadTypeForSync{backendKarta},
		})
	res, err := h.backendClient.WorkloadsClient.WorkloadPropertiesAPI.SyncWorkloadTypesExecute(request)
	return h.backendClient.HandleHttpResponse(res, err, fmt.Sprintf("Upsert Karta %s", karta.Name))
}

func (h *KartaHandler) HandleDelete(karta *kartav1alpha1.Karta) error {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(h.resourceName, constants.DeleteOperation)()
	backendKarta, err := parsers.ParseBackendKarta(karta, true)
	if err != nil {
		return err
	}

	h.syncMutex.Lock()
	defer h.syncMutex.Unlock()

	request := h.backendClient.WorkloadsClient.WorkloadPropertiesAPI.
		SyncWorkloadTypes(h.backendClient.Context()).
		ClusterId(config.Config().ClusterUuid).
		WorkloadTypesForSync(swagger.WorkloadTypesForSync{
			FullSync:      ptr.To(false),
			WorkloadTypes: []swagger.WorkloadTypeForSync{backendKarta},
		})
	res, err := h.backendClient.WorkloadsClient.WorkloadPropertiesAPI.SyncWorkloadTypesExecute(request)
	return h.backendClient.HandleHttpResponse(res, err, fmt.Sprintf("Delete Karta %s", karta.Name))
}

func (h *KartaHandler) HandleSyncEvent(waitGroup *sync.WaitGroup) {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(h.resourceName, constants.SyncOperation)()
	defer waitGroup.Done()

	// Allow at most 1 waiting goroutine; additional callers are dropped because
	// a full sync subsumes them.
	if !h.syncWaiting.CompareAndSwap(false, true) {
		zap.S().Debugf("Skipping Kartas sync - another goroutine is already waiting")
		return
	}
	h.syncMutex.Lock()
	h.syncWaiting.Store(false)
	defer h.syncMutex.Unlock()

	zap.S().Infof("Start Kartas Sync")

	kartas := &kartav1alpha1.KartaList{}
	err := h.runaiClient.List(context.Background(), kartas)
	if err != nil {
		zap.S().Errorf("Failed to list Kartas, error: %s", err.Error())
		return
	}

	backendKartas := make([]swagger.WorkloadTypeForSync, 0, len(kartas.Items))
	for _, karta := range kartas.Items {
		backendKarta, parseErr := parsers.ParseBackendKarta(&karta, false)
		if parseErr != nil {
			zap.S().Errorf("Failed to parse Karta %s, error: %s", karta.Name, parseErr.Error())
			continue
		}
		backendKartas = append(backendKartas, backendKarta)
	}

	request := h.backendClient.WorkloadsClient.WorkloadPropertiesAPI.
		SyncWorkloadTypes(h.backendClient.Context()).
		ClusterId(config.Config().ClusterUuid).
		WorkloadTypesForSync(swagger.WorkloadTypesForSync{
			FullSync:      ptr.To(true),
			WorkloadTypes: backendKartas,
		})
	_, err = h.backendClient.WorkloadsClient.WorkloadPropertiesAPI.SyncWorkloadTypesExecute(request)
	if err != nil {
		zap.S().Errorf("Failed to sync Kartas, error: %s", err.Error())
	}
}
