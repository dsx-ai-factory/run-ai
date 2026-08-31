/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"fmt"
	"sync"
	"time"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	assetsApi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/metrics"
	v1 "k8s.io/api/core/v1"

	"github.com/hashicorp/go-multierror"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-application-packages/concurrency"
	runaiBackendClient "github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/backend-client"
	configmaps "github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/config-maps"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AssetsSyncReportInfo struct {
	numAssets int
	theReport assetsApi.ClusterSyncReport
}

var syncReportPerAssetType = concurrency.NewConcurrentMap[*AssetsSyncReportInfo](nil)

type clusterUpdatesOfAssets struct {
	ToUpdate  []client.Object
	UpdateRef []int // the reference of this update to the array of asset updates that were received from the CP

	ToDelete  []types.NamespacedName
	DeleteRef []int // the reference of this deletion to the array of asset updates that were received from the CP
}

const (
	maxNumAssetWorkers = 4

	AssetGroupDatasources = "datasources"
	AssetGroupCredentials = "credentials"
)

type AssetsHandler struct {
	*CommonHandler
	syncCounter int64
	assetGroup  string
	metricLabel string
}

func NewCredentialAssetsHandler(runaiBackendClient *runaiBackendClient.BackendClient, runaiClient client.Client,
	cachedClient client.Reader) (ResourceHandler, error) {
	return NewAssetsHandler(runaiBackendClient, AssetGroupCredentials, runaiClient, cachedClient, "credential")
}

func NewDatasourceAssetsHandler(runaiBackendClient *runaiBackendClient.BackendClient, runaiClient client.Client,
	cachedClient client.Reader) (ResourceHandler, error) {
	return NewAssetsHandler(runaiBackendClient, AssetGroupDatasources, runaiClient, cachedClient, "datasource")
}

func NewAssetsHandler(runaiBackendClient *runaiBackendClient.BackendClient, assetGroup string, runaiClient client.Client,
	cachedClient client.Reader, metricLabel string,
) (ResourceHandler, error) {
	var object client.Object
	var objectInspector AssetsObjectInspector

	switch assetGroup {
	case AssetGroupDatasources:
		object = &v1.PersistentVolumeClaim{}
	case AssetGroupCredentials:
		object = &v1.Secret{}
	default:
		return nil, fmt.Errorf("unsupported asset group: %s", assetGroup)
	}

	handlingOptions := HandlingOptions{deletePropagationOrphan: false}
	commonHandler, err := NewCommonHandler(runaiBackendClient, runaiClient, cachedClient, object, objectInspector, handlingOptions)
	if err != nil {
		return nil, err
	}

	agentConfigData, _ := configmaps.GetAgentConfig(cachedClient)

	switch assetGroup {
	case AssetGroupDatasources:
		return &AssetsHandler{
			CommonHandler: commonHandler,
			assetGroup:    assetGroup,
			syncCounter:   agentConfigData.DatasourcesSyncCounter,
			metricLabel:   metricLabel,
		}, nil
	case AssetGroupCredentials:
		return &AssetsHandler{
			CommonHandler: commonHandler,
			assetGroup:    assetGroup,
			syncCounter:   agentConfigData.CredentialsSyncCounter,
			metricLabel:   metricLabel,
		}, nil
	}

	return nil, fmt.Errorf("unsupported asset group: %s", assetGroup)
}

func (ah *AssetsHandler) Handle() error {
	var err error
	var prevHighestUpdate int64
	numWorkers := maxNumAssetWorkers

	//
	//  make sure each workload type has an entry in the sync report
	//
	syncReport, found := syncReportPerAssetType.Load(ah.assetGroup)
	if !found {
		syncReport = &AssetsSyncReportInfo{}
		syncReportPerAssetType.Store(ah.assetGroup, syncReport)
	}

	for {
		backendAssets, innerErr := ah.GetUpdatesFromBackend()
		if innerErr != nil {
			log.Error().Msgf("Had an error while getting backend %s assets, sync counter: %d, error: %s", ah.assetGroup, ah.syncCounter, innerErr.Message)
			return fmt.Errorf("%s", innerErr.Message)
		}

		if backendAssets == nil {
			log.Error().Msgf("%s assets were not received from the backend for sync counter %d", ah.assetGroup, ah.syncCounter)
			return nil
		}

		var highestUpdate, highestSysUpdate int64
		if backendAssets.ResponseInfo != nil && backendAssets.ResponseInfo.HighestUpdate != nil {
			highestUpdate = *backendAssets.ResponseInfo.HighestUpdate
		}

		if backendAssets.SystemInfo != nil && backendAssets.SystemInfo.HighestUpdate != nil {
			highestSysUpdate = *backendAssets.SystemInfo.HighestUpdate
		}

		numAssets := len(backendAssets.Assets)
		log.Debug().Msgf("Backend %s assets response - prev highest update: %d, highest update: %d, highest system update: %d, size: %d", ah.assetGroup, prevHighestUpdate, highestUpdate, highestSysUpdate, numAssets)

		syncReport.numAssets = numAssets

		if numAssets == 0 {
			log.Debug().Msgf("No %s assets to handle while counter is %d", ah.assetGroup, ah.syncCounter)
			break
		}

		syncReport.theReport.Updates = make([]assetsApi.ClusterSyncReportUpdate, numAssets)

		processErr := ah.processAssetsConcurrently(numAssets, numWorkers, backendAssets)
		if processErr != nil {
			log.Error().Msgf("Had an error while processing %s assetd, error: %s", ah.assetGroup, processErr.Error())
			err = multierror.Append(err, processErr)
		}

		// update the highest update counter
		// NOTE! we're updating the sync counter even if there was an error in processing the assets
		// because there is no point to attempting again.
		setError := ah.setSyncCounter(highestUpdate)
		if setError != nil {
			log.Error().Msgf("Had an error while updating %s assets sync counter, error: %s", ah.assetGroup, setError.Error())
			err = multierror.Append(err, setError)
		}

		// if the highest update counter is the same as the highest system update counter, it means that there are no more cluster assets to handle
		if highestUpdate == highestSysUpdate {
			log.Debug().Msgf("No more %s assets to handle", ah.assetGroup)
			break
		}
	}

	if err != nil {
		log.Error().Msgf("Failed to handle some of cluster assets (create or update), error: %s", err.Error())
		return err
	}

	log.Debug().Msgf("Finished handling %s assets", ah.assetGroup)
	return nil
}

func (ah *AssetsHandler) MetricLabel() string {
	return ah.metricLabel
}

func (ah *AssetsHandler) processAssetsConcurrently(numUpdates int, numWorkers int, updates *AssetsUpdateFromBackend) error {
	var wg sync.WaitGroup
	var err error

	syncReport, _ := syncReportPerAssetType.Load(ah.assetGroup)

	// to be on the safe side make sure at least one worker is running
	if numWorkers <= 0 {
		log.Warn().Msgf("number of workers changed from %d to 1", numWorkers)
		numWorkers = 1
	}

	chunkSize := numUpdates / numWorkers

	if chunkSize == 0 {
		// if there are fewer updates than workers, we will have only one worker who will handle all the updates
		numWorkers = 1
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		// Calculate the start and end indices for each worker's chunk
		start := i * chunkSize
		end := (i + 1) * chunkSize

		// The last worker may get extra items if numUpdates is not divisible by numWorkers
		if i == numWorkers-1 {
			end = numUpdates
		}

		log.Debug().Msgf("Worker %d: start processing, start: %v, end: %v", i, start, end)
		go func() {
			perAssetError, innerErr := ah.ProcessAssets(updates.Assets[start:end], &wg)
			if innerErr != nil {
				err = multierror.Append(err, innerErr)
			}

			for idx := start; idx < end; idx++ {
				syncReport.theReport.Updates[idx] = assetsApi.ClusterSyncReportUpdate{
					Id:                convutils.Ref(updates.Assets[idx].Meta.Id),
					UpdateCount:       convutils.Ref(updates.Assets[idx].UpdateCount),
					UpdateAttemptTime: convutils.Ref(time.Now()),
				}
				if perAssetError[idx-start] != nil {
					syncReport.theReport.Updates[idx].Message = *assetsApi.NewNullableString(convutils.Ref(perAssetError[idx-start].Error()))
				}
			}
		}()
	}

	// Wait for all workers to finish
	wg.Wait()
	return err
}

func (ah *AssetsHandler) ProcessAssets(backendAsset []AssetsFromBackend, wg *sync.WaitGroup) ([]error, error) {
	defer wg.Done()
	var err error

	var clusterUpdates clusterUpdatesOfAssets
	var innerErr error

	switch ah.assetGroup {
	case AssetGroupDatasources:
		clusterUpdates, innerErr = ah.composeClusterUpdatesOfDatasources(backendAsset)
	case AssetGroupCredentials:
		clusterUpdates, innerErr = ah.composeClusterUpdatesOfCredentials(backendAsset)
	}
	if innerErr != nil {
		log.Error().Msgf("Had an error while convering %s assets: %s", ah.assetGroup, innerErr.Error())
		err = multierror.Append(err, innerErr)
	}

	metrics.GetMetricsExporter().AddResourceAction(ah.MetricLabel(), metrics.ActionCreateOrUpdate, len(clusterUpdates.ToUpdate))
	metrics.GetMetricsExporter().AddResourceAction(ah.MetricLabel(), metrics.ActionDelete, len(clusterUpdates.ToDelete))

	allObjectsErrors := make([]error, len(backendAsset))

	if len(clusterUpdates.ToUpdate) > 0 {
		_, perObjectError, innerErr := ah.CreateOrUpdateObjects(clusterUpdates.ToUpdate)
		if innerErr != nil {
			err = multierror.Append(err, innerErr)
		}

		for idx, objectErr := range perObjectError {
			toIndex := clusterUpdates.UpdateRef[idx]
			allObjectsErrors[toIndex] = objectErr
		}
	}

	if len(clusterUpdates.ToDelete) > 0 {
		perObjectError, innerErr := ah.DeleteObjects(clusterUpdates.ToDelete)
		if innerErr != nil {
			err = multierror.Append(err, innerErr)
		}

		for idx, objectErr := range perObjectError {
			toIndex := clusterUpdates.DeleteRef[idx]
			allObjectsErrors[toIndex] = objectErr
		}
	}

	return allObjectsErrors, err
}

func (ah *AssetsHandler) setSyncCounter(highestUpdate int64) error {
	ah.syncCounter = highestUpdate
	switch ah.assetGroup {
	case AssetGroupDatasources:
		err := configmaps.UpdateDatasourceAssetsSyncCounter(ah.runaiClient, ah.cachedClient, highestUpdate)
		if err != nil {
			return err
		}
		log.Info().Msgf("Updated datasources sync counter to: %d", highestUpdate)
	case AssetGroupCredentials:
		err := configmaps.UpdateCredentialsAssetsSyncCounter(ah.runaiClient, ah.cachedClient, highestUpdate)
		if err != nil {
			return err
		}
		log.Info().Msgf("Updated credentials sync counter to: %d", highestUpdate)
	}
	return nil
}
