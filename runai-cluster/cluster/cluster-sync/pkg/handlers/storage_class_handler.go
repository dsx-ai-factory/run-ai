/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"
	"sync"

	trackerclient "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client/generated"
	runaiBackendClient "github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	"go.uber.org/zap"
	storagev1 "k8s.io/api/storage/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StorageClassHandler struct {
	backendClient *runaiBackendClient.BackendClient
	cachedReader  client.Reader
}

func NewStorageClassHandler(cachedReader client.Reader, runaiBackendClient *runaiBackendClient.BackendClient) *StorageClassHandler {
	return &StorageClassHandler{
		cachedReader:  cachedReader,
		backendClient: runaiBackendClient,
	}
}

func (storageClassHandler *StorageClassHandler) HandleUpdateItem(key client.ObjectKey) error {
	storageClass := &storagev1.StorageClass{}
	err := storageClassHandler.cachedReader.Get(context.Background(), key, storageClass)
	if err != nil {
		return err
	}

	return storageClassHandler.HandleStorageClassUpdate(storageClass)
}

func (storageClassHandler *StorageClassHandler) HandleStorageClassAdd(storageClass *storagev1.StorageClass) error {
	zap.S().Infof("Adding StorageClass < Name: %v, IsDefault: %v >", storageClass.Name, parsers.CheckIfStorageClassIsDefault(storageClass))
	parsedStorageClass := parsers.ParseStorageClassToServerScheme(storageClass)

	_, res, err := storageClassHandler.backendClient.K8sObjectTrackerClient.CreateStorageClass( // StorageClassesApi from  *swagger.APIClient
		storageClassHandler.backendClient.Context(), parsedStorageClass)

	return storageClassHandler.backendClient.HandleHttpResponse(res, err, fmt.Sprintf("Added StorageClasses %s", parsedStorageClass.StorageClassName))
}

func (storageClassHandler *StorageClassHandler) HandleStorageClassDelete(storageClass *storagev1.StorageClass) error {
	zap.S().Debugf("Deleting StorageClass < Name: %v,	IsDefault: %v >", storageClass.Name, parsers.CheckIfStorageClassIsDefault(storageClass))

	res, err := storageClassHandler.backendClient.K8sObjectTrackerClient.DeleteStorageClass(
		storageClassHandler.backendClient.Context(), config.Config().ClusterUuid, storageClass.Name)

	return storageClassHandler.backendClient.HandleHttpResponse(res, err, fmt.Sprintf("Deleted StorageClass %s", storageClass.Name))
}

func (storageClassHandler *StorageClassHandler) HandleStorageClassUpdate(storageClass *storagev1.StorageClass) error {
	zap.S().Debugf("Updating StorageClass < Name: %v,	IsDefault: %v >", storageClass.Name, parsers.CheckIfStorageClassIsDefault(storageClass))
	parsedStorageClass := parsers.ParseStorageClassToServerScheme(storageClass)
	_, res, err := storageClassHandler.backendClient.K8sObjectTrackerClient.UpdateStorageClass(
		storageClassHandler.backendClient.Context(), parsedStorageClass)

	return storageClassHandler.backendClient.HandleHttpResponse(res, err, fmt.Sprintf("Updated StorageClasses %s", parsedStorageClass.StorageClassName))
}

func (storageClassHandler *StorageClassHandler) HandleSyncEvent(waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	zap.S().Infof("Start syncing StorageClasses")

	// Don't convert to var since this sends an empty array when no data is in cluster vs null (which will result in bad request (400))
	storageClassesList := storagev1.StorageClassList{}
	err := storageClassHandler.cachedReader.List(context.Background(), &storageClassesList)
	if err != nil {
		zap.S().Errorf("Failed listing storage classes, error: %s", err.Error())
	}

	storageClasses := trackerclient.StorageClasses{}
	// loop over storageClassesList.Items
	for _, storageClass := range storageClassesList.Items {
		storageClasses.Items = append(storageClasses.Items, parsers.ParseStorageClassToServerScheme(&storageClass))
	}

	_, res, httpErr := storageClassHandler.backendClient.K8sObjectTrackerClient.SyncStorageClass(
		storageClassHandler.backendClient.Context(), config.Config().ClusterUuid, storageClasses.Items)

	finalError := storageClassHandler.backendClient.HandleHttpResponse(res, httpErr, "Sync storageClass data")

	if finalError != nil {
		zap.S().Errorf("Had an error while syncing StorageClasses: %+v", finalError)
	}

}
