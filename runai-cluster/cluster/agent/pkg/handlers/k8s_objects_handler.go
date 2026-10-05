/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/concurrency"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	assetsApi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	runaiBackendClient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/backend-client"
	configmaps "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/config-maps"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/metrics"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// K8sObjectsHandler syncs the k8s-object asset family from the control
// plane into the cluster. One handler covers every kind in the family —
// resource-claim-template today, more to follow — via the shared
// /api/v1/k8s-objects/updates endpoint and the per-kind k8sObjectBuilder
// registry (k8s_objects_builders.go).
//
// Structurally the handler is a thin orchestrator that owns one
// kindHandler per registered k8sObjectBuilder. Each kindHandler bundles a
// CommonHandler bound to that kind's typed object, which keeps the
// upstream CommonHandler contract (one GVK per handler) intact while
// still exposing a single ResourceHandler to runai_agent.go.
//
// Lifecycle mirrors AssetsHandler: periodic sync tick with an inner
// catch-up loop, cursor persisted to the agent-config ConfigMap, per-asset
// report piggybacked on the next poll, concurrent worker pool. The split
// from AssetsHandler exists because the k8s-objects wire shape
// (K8sObjectSpec = objectMeta + untyped objectSpec) does not fit the
// typed per-group plumbing in AssetsHandler without noisy switch
// statements.
type K8sObjectsHandler struct {
	backendClient *runaiBackendClient.BackendClient
	runaiClient   client.Client
	cachedClient  client.Reader

	syncCounter  int64
	builders     map[assetsApi.AssetKind]k8sObjectBuilder
	kindHandlers map[assetsApi.AssetKind]*CommonHandler
}

// k8sObjectsSyncReport holds the pending ClusterSyncReport that will be
// posted on the NEXT poll. The report describes the outcome of the
// previous poll's updates — same piggyback pattern as credentials and
// datasources. Stored in a concurrent map (rather than a field) so the
// backend client lookup in GetK8sObjectsUpdatesFromBackend doesn't need a
// handler receiver, matching the assets_backend.go contract.
var (
	k8sObjectsSyncReportKey = "k8s-objects"
	k8sObjectsSyncReport    = concurrency.NewConcurrentMap[*AssetsSyncReportInfo](nil)
)

// k8sObjectsMetricLabel is what the metrics exporter buckets durations
// and failures by. Singular to match the credentials/datasource labels.
const k8sObjectsMetricLabel = "k8s-object"

// NewK8sObjectsHandler constructs the k8s-objects sync handler. One
// CommonHandler is built per registered builder so each kind has its own
// typed reconcile pipeline — the top-level K8sObjectsHandler only
// dispatches updates and owns the shared sync cursor + sync report.
func NewK8sObjectsHandler(backendClient *runaiBackendClient.BackendClient,
	runaiClient client.Client, cachedClient client.Reader,
) (ResourceHandler, error) {
	if len(k8sObjectBuilders) == 0 {
		return nil, fmt.Errorf("k8sObjectBuilders registry is empty; register at least one builder")
	}

	inspector := newK8sObjectsInspector(k8sObjectBuilders)
	handlingOptions := HandlingOptions{deletePropagationOrphan: false}

	kindHandlers := make(map[assetsApi.AssetKind]*CommonHandler, len(k8sObjectBuilders))
	for kind, builder := range k8sObjectBuilders {
		commonHandler, err := NewCommonHandler(backendClient, runaiClient, cachedClient, builder.NewEmpty(), inspector, handlingOptions)
		if err != nil {
			return nil, fmt.Errorf("init k8s-object common handler for kind %s: %w", kind, err)
		}
		kindHandlers[kind] = commonHandler
	}

	agentConfigData, _ := configmaps.GetAgentConfig(cachedClient)
	syncCounter := int64(0)
	if agentConfigData != nil {
		syncCounter = agentConfigData.K8sObjectsSyncCounter
	}

	return &K8sObjectsHandler{
		backendClient: backendClient,
		runaiClient:   runaiClient,
		cachedClient:  cachedClient,
		syncCounter:   syncCounter,
		builders:      k8sObjectBuilders,
		kindHandlers:  kindHandlers,
	}, nil
}

func (h *K8sObjectsHandler) MetricLabel() string {
	return k8sObjectsMetricLabel
}

// Handle runs one sync tick: drain the control plane's pending updates in
// catch-up batches, apply them, then persist the local cursor so the next
// tick resumes after the batch we processed. On error we still advance
// the cursor — same policy as AssetsHandler, documented there — because
// retrying the exact same batch has no new information and would stall
// progress behind the first bad asset.
func (h *K8sObjectsHandler) Handle() error {
	var err error
	numWorkers := maxNumAssetWorkers

	report, found := k8sObjectsSyncReport.Load(k8sObjectsSyncReportKey)
	if !found {
		report = &AssetsSyncReportInfo{}
		k8sObjectsSyncReport.Store(k8sObjectsSyncReportKey, report)
	}

	for {
		backendUpdates, innerErr := h.GetK8sObjectsUpdatesFromBackend()
		if innerErr != nil {
			log.Error().Msgf("Had an error while getting backend k8s-objects, sync counter: %d, error: %s", h.syncCounter, innerErr.Message)
			return fmt.Errorf("%s", innerErr.Message)
		}

		if backendUpdates == nil {
			log.Error().Msgf("k8s-objects were not received from the backend for sync counter %d", h.syncCounter)
			return nil
		}

		var highestUpdate, highestSysUpdate int64
		if backendUpdates.ResponseInfo != nil && backendUpdates.ResponseInfo.HighestUpdate != nil {
			highestUpdate = *backendUpdates.ResponseInfo.HighestUpdate
		}
		if backendUpdates.SystemInfo != nil && backendUpdates.SystemInfo.HighestUpdate != nil {
			highestSysUpdate = *backendUpdates.SystemInfo.HighestUpdate
		}

		numUpdates := len(backendUpdates.Assets)
		log.Debug().Msgf("Backend k8s-objects response - highest update: %d, highest system update: %d, size: %d", highestUpdate, highestSysUpdate, numUpdates)

		report.numAssets = numUpdates

		if numUpdates == 0 {
			log.Debug().Msgf("No k8s-objects to handle while counter is %d", h.syncCounter)
			break
		}

		report.theReport.Updates = make([]assetsApi.ClusterSyncReportUpdate, numUpdates)

		if processErr := h.processUpdatesConcurrently(numUpdates, numWorkers, backendUpdates, report); processErr != nil {
			log.Error().Msgf("Had an error while processing k8s-objects, error: %s", processErr.Error())
			err = multierror.Append(err, processErr)
		}

		if setErr := h.setSyncCounter(highestUpdate); setErr != nil {
			log.Error().Msgf("Had an error while updating k8s-objects sync counter, error: %s", setErr.Error())
			err = multierror.Append(err, setErr)
		}

		if highestUpdate == highestSysUpdate {
			log.Debug().Msgf("No more k8s-objects to handle")
			break
		}
	}

	if err != nil {
		log.Error().Msgf("Failed to handle some of cluster k8s-objects, error: %s", err.Error())
		return err
	}
	log.Debug().Msgf("Finished handling k8s-objects")
	return nil
}

func (h *K8sObjectsHandler) setSyncCounter(highestUpdate int64) error {
	h.syncCounter = highestUpdate
	if err := configmaps.UpdateK8sObjectsAssetsSyncCounter(h.runaiClient, h.cachedClient, highestUpdate); err != nil {
		return err
	}
	log.Info().Msgf("Updated k8s-objects sync counter to: %d", highestUpdate)
	return nil
}

// processUpdatesConcurrently splits the batch across worker goroutines in
// the same chunking scheme used by AssetsHandler.processAssetsConcurrently.
// The sync report is filled in-place per-index so the next poll's piggyback
// report is accurate even when individual assets fail.
func (h *K8sObjectsHandler) processUpdatesConcurrently(numUpdates, numWorkers int, updates *K8sObjectsUpdateFromBackend, report *AssetsSyncReportInfo) error {
	var wg sync.WaitGroup
	var err error
	var errMu sync.Mutex

	if numWorkers <= 0 {
		log.Warn().Msgf("number of workers changed from %d to 1", numWorkers)
		numWorkers = 1
	}

	chunkSize := numUpdates / numWorkers
	if chunkSize == 0 {
		numWorkers = 1
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		start := i * chunkSize
		end := (i + 1) * chunkSize
		if i == numWorkers-1 {
			end = numUpdates
		}

		log.Debug().Msgf("k8s-objects worker %d: start processing, start: %v, end: %v", i, start, end)
		go func() {
			perAssetError, innerErr := h.processUpdates(updates.Assets[start:end], &wg)
			if innerErr != nil {
				errMu.Lock()
				err = multierror.Append(err, innerErr)
				errMu.Unlock()
			}

			for idx := start; idx < end; idx++ {
				report.theReport.Updates[idx] = assetsApi.ClusterSyncReportUpdate{
					Id:                convutils.Ref(updates.Assets[idx].Meta.Id),
					UpdateCount:       convutils.Ref(updates.Assets[idx].UpdateCount),
					UpdateAttemptTime: convutils.Ref(time.Now()),
				}
				if perAssetError[idx-start] != nil {
					report.theReport.Updates[idx].Message = *assetsApi.NewNullableString(convutils.Ref(perAssetError[idx-start].Error()))
				}
			}
		}()
	}

	wg.Wait()
	return err
}

// processUpdates converts and reconciles one slice of the current batch.
// Updates are bucketed per-kind so each kind's typed CommonHandler sees
// only its own objects — this keeps the "one GVK per CommonHandler"
// contract intact while still letting the top-level handler drain a
// mixed-kind feed. Per-asset conversion failures (unknown kind, bad spec,
// missing namespace) are recorded in the returned per-asset error slice
// so the sync report surfaces them to the control plane; apply/delete
// failures from the cluster side come from each kind's CommonHandler.
func (h *K8sObjectsHandler) processUpdates(updates []K8sObjectFromBackend, wg *sync.WaitGroup) ([]error, error) {
	defer wg.Done()
	var err error

	projectToNamespaceMap, nsErr := h.anyKindHandler().ProjectToNamespaceList()
	if nsErr != nil {
		return make([]error, len(updates)), nsErr
	}

	perAssetError := make([]error, len(updates))

	type kindBucket struct {
		toUpdate    []client.Object
		updateRef   []int
		toRecreate  []client.Object
		recreateRef []int
		toDelete    []types.NamespacedName
		deleteRef   []int
	}
	buckets := make(map[assetsApi.AssetKind]*kindBucket, len(h.kindHandlers))

	for idx, update := range updates {
		builder, ok := h.builders[update.Meta.Kind]
		if !ok {
			perAssetError[idx] = fmt.Errorf("unsupported k8s-object kind %q (agent build is older than the control plane; register a k8sObjectBuilder for this kind)", update.Meta.Kind)
			continue
		}

		namespace, nsErr := h.namespaceFor(&update, projectToNamespaceMap)
		if nsErr != nil {
			perAssetError[idx] = nsErr
			continue
		}

		bucket, ok := buckets[update.Meta.Kind]
		if !ok {
			bucket = &kindBucket{}
			buckets[update.Meta.Kind] = bucket
		}

		if update.IsDeleted {
			bucket.toDelete = append(bucket.toDelete, types.NamespacedName{Namespace: namespace, Name: update.Meta.Name})
			bucket.deleteRef = append(bucket.deleteRef, idx)
			continue
		}

		obj, buildErr := builder.Build(update, namespace)
		if buildErr != nil {
			perAssetError[idx] = buildErr
			continue
		}

		// Kinds whose Spec the k8s API server marks immutable (RCT in
		// resource.k8s.io/v1) must be handled with a delete-and-recreate
		// pass instead of Update, otherwise the API server returns
		// "spec is immutable" and the asset is stuck. The ShouldReCreate
		// probe runs against the current cluster state (via the cached
		// reader) and only reroutes when there is actually a cluster
		// object whose immutable fields diverge from the freshly-built
		// one — the first-time create path still goes through toUpdate.
		if h.needsRecreation(builder, obj) {
			bucket.toRecreate = append(bucket.toRecreate, obj)
			bucket.recreateRef = append(bucket.recreateRef, idx)
			continue
		}

		bucket.toUpdate = append(bucket.toUpdate, obj)
		bucket.updateRef = append(bucket.updateRef, idx)
	}

	for kind, bucket := range buckets {
		kindHandler := h.kindHandlers[kind]

		metrics.GetMetricsExporter().AddResourceAction(h.MetricLabel(), metrics.ActionCreateOrUpdate, len(bucket.toUpdate)+len(bucket.toRecreate))
		metrics.GetMetricsExporter().AddResourceAction(h.MetricLabel(), metrics.ActionDelete, len(bucket.toDelete))

		if len(bucket.toUpdate) > 0 {
			_, perObjectErr, innerErr := kindHandler.CreateOrUpdateObjects(bucket.toUpdate)
			if innerErr != nil {
				err = multierror.Append(err, innerErr)
			}
			for i, objErr := range perObjectErr {
				perAssetError[bucket.updateRef[i]] = objErr
			}
		}

		if len(bucket.toRecreate) > 0 {
			perObjectErr, innerErr := h.recreateObjects(kind, bucket.toRecreate)
			if innerErr != nil {
				err = multierror.Append(err, innerErr)
			}
			for i, objErr := range perObjectErr {
				perAssetError[bucket.recreateRef[i]] = objErr
			}
		}

		if len(bucket.toDelete) > 0 {
			perObjectErr, innerErr := kindHandler.DeleteObjects(bucket.toDelete)
			if innerErr != nil {
				err = multierror.Append(err, innerErr)
			}
			for i, objErr := range perObjectErr {
				perAssetError[bucket.deleteRef[i]] = objErr
			}
		}
	}

	return perAssetError, err
}

// needsRecreation returns true when the freshly-built object differs on
// immutable fields from the current cluster state. A missing cluster
// object or a lookup error short-circuits to false: a first-time create
// belongs in the normal toUpdate pass, and lookup errors (non-NotFound)
// are left for CreateOrUpdateObjects to surface in its own per-object
// error slice so the sync report sees them.
func (h *K8sObjectsHandler) needsRecreation(builder k8sObjectBuilder, obj client.Object) bool {
	existing := builder.NewEmpty()
	key := client.ObjectKey{Namespace: obj.GetNamespace(), Name: obj.GetName()}
	if err := h.cachedClient.Get(context.Background(), key, existing); err != nil {
		return false
	}
	return builder.ShouldReCreate(obj, existing)
}

// recreateObjects deletes then re-creates each object in the slice, used
// for kinds whose Spec is immutable at the API-server level. The two
// operations must happen in the same sync tick because the k8s-objects
// feed is cursor-based: an asset only reaches the agent at the counter
// it changed on, so a Delete without a same-tick Create would leave the
// asset gone from the cluster until the user edits it again.
//
// Create is wrapped in retry.OnError for IsAlreadyExists because the
// typed client may observe stale cache state immediately after Delete;
// the default backoff is bounded and short so a stuck finalizer surfaces
// as an error rather than an infinite loop.
func (h *K8sObjectsHandler) recreateObjects(kind assetsApi.AssetKind, objects []client.Object) ([]error, error) {
	perObjectError := make([]error, len(objects))
	var err error

	for i, obj := range objects {
		name := types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}

		log.Info().Msgf("Recreating k8s-object %s <%s/%s> because its spec diverges on an immutable field",
			kind, name.Namespace, name.Name)

		if delErr := h.runaiClient.Delete(context.Background(), obj); delErr != nil && !errors.IsNotFound(delErr) {
			perObjectError[i] = fmt.Errorf("delete %s/%s for recreate: %w", name.Namespace, name.Name, delErr)
			err = multierror.Append(err, perObjectError[i])
			continue
		}

		// Create is retried via DeepCopy because the typed client sets
		// ResourceVersion on the input when a Create returns, which would
		// make a subsequent retry fail with "resourceVersion should not
		// be set on objects to be created".
		createErr := retry.OnError(retry.DefaultBackoff, errors.IsAlreadyExists, func() error {
			fresh := obj.DeepCopyObject().(client.Object)
			fresh.SetResourceVersion("")
			return h.runaiClient.Create(context.Background(), fresh)
		})
		if createErr != nil {
			perObjectError[i] = fmt.Errorf("create %s/%s after delete: %w", name.Namespace, name.Name, createErr)
			err = multierror.Append(err, perObjectError[i])
			continue
		}
	}

	return perObjectError, err
}

// namespaceFor resolves the cluster namespace for a k8s-object asset. The
// convention matches credentials/datasources: project-scoped assets land
// in the project namespace via the KAI Project CR lookup; cluster- and
// department-scoped assets go to the `runai` namespace. Tenant scope is
// explicitly rejected — the control plane constrains the k8s-objects
// family to cluster scope and below, and a tenant-scoped asset here means
// drift that is better surfaced than silently defaulted.
func (h *K8sObjectsHandler) namespaceFor(update *K8sObjectFromBackend, projectToNamespaceMap map[string]string) (string, error) {
	switch update.Meta.Scope {
	case assetsApi.SCOPE_PROJECT:
		projectIdPtr := update.Meta.ProjectId.Get()
		if projectIdPtr == nil {
			return "", fmt.Errorf("project-scoped k8s-object %s is missing projectId", update.Meta.Name)
		}
		projectId := strconv.Itoa(int(*projectIdPtr))
		namespace, found := projectToNamespaceMap[projectId]
		if !found || namespace == "" {
			return "", fmt.Errorf("project %s namespace not found for k8s-object %s", projectId, update.Meta.Name)
		}
		return namespace, nil
	case assetsApi.SCOPE_CLUSTER, assetsApi.SCOPE_DEPARTMENT:
		return constants.NamespaceRunAi, nil
	case assetsApi.SCOPE_TENANT:
		return "", fmt.Errorf("tenant-scoped k8s-objects are not supported for %s (control plane should constrain this family to cluster scope and below)", update.Meta.Name)
	default:
		return "", fmt.Errorf("unknown scope %q for k8s-object %s", update.Meta.Scope, update.Meta.Name)
	}
}

// anyKindHandler returns an arbitrary per-kind CommonHandler for use with
// helpers that do not depend on the specific GVK (ProjectToNamespaceList
// today). Avoids wiring a duplicate client plumbing path on
// K8sObjectsHandler just to read KAI Project CRs.
func (h *K8sObjectsHandler) anyKindHandler() *CommonHandler {
	for _, handler := range h.kindHandlers {
		return handler
	}
	// Unreachable: NewK8sObjectsHandler refuses to construct with an empty
	// builder registry. Keep a defensive panic rather than a nil deref so a
	// future regression surfaces as a loud message in the agent log.
	panic("K8sObjectsHandler initialized without any kind handlers")
}
