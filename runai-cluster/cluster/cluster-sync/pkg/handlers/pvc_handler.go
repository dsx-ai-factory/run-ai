/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	auth_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	assetsswagger "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	clustercommonconsts "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"go.uber.org/zap"

	k8sobjectstracker "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	runaiBackendClient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/util"

	v1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PVCHandler struct {
	backendClient *runaiBackendClient.BackendClient
	cachedReader  client.Reader
}

func NewPVCHandler(cachedReader client.Reader, runaiBackendClient *runaiBackendClient.BackendClient) *PVCHandler {
	return &PVCHandler{
		cachedReader:  cachedReader,
		backendClient: runaiBackendClient,
	}
}

func (pvcHandler *PVCHandler) HandleUpdateItem(key client.ObjectKey) error {
	pvc := &corev1.PersistentVolumeClaim{}
	err := pvcHandler.cachedReader.Get(context.Background(), key, pvc)
	if err != nil {
		return err
	}

	return pvcHandler.HandlePVCUpdate(pvc)
}

func (pvcHandler *PVCHandler) HandlePVCAdd(pvc *corev1.PersistentVolumeClaim) error {
	zap.S().Infof("Adding PVC <Name: %v, Namespace: %v>", pvc.Name, pvc.Namespace)
	return pvcHandler.handlePVCAddUpdate(pvc)
}

func (pvcHandler *PVCHandler) HandlePVCUpdate(pvc *corev1.PersistentVolumeClaim) error {
	zap.S().Infof("Updating PVC <Name: %v, Namespace: %v>", pvc.Name, pvc.Namespace)
	return pvcHandler.handlePVCAddUpdate(pvc)
}

func (pvcHandler *PVCHandler) handlePVCAddUpdate(pvc *corev1.PersistentVolumeClaim) error {
	parsedPvc, err := parsers.ParsePvcToServerScheme(pvc)
	if err != nil {
		zap.S().Warnf("Error parsing PVC %s:%s, error: %v", pvc.Namespace, pvc.Name, err)
		return err
	}

	var response *http.Response

	clusterId := config.Config().ClusterUuid

	switch parsedPvc.Scope.Type {
	case auth_client.SCOPETYPE_CLUSTER:
		_, response, err = pvcHandler.backendClient.K8sObjectTrackerClient.CreateClusterWidePvc(
			pvcHandler.backendClient.Context(), clusterId, pvc.Name, parsedPvc.CreationRequest)

	case auth_client.SCOPETYPE_DEPARTMENT:
		_, response, err = pvcHandler.backendClient.K8sObjectTrackerClient.CreateDepartmentWidePvc(
			pvcHandler.backendClient.Context(), clusterId, parsedPvc.Scope.Id, pvc.Name, parsedPvc.CreationRequest)

	case auth_client.SCOPETYPE_PROJECT:
		_, response, err = pvcHandler.backendClient.K8sObjectTrackerClient.CreateProjectPvc(
			pvcHandler.backendClient.Context(), clusterId, parsedPvc.Scope.Id, pvc.Name, parsedPvc.CreationRequest)
	}

	return pvcHandler.backendClient.HandleHttpResponse(response, err, fmt.Sprintf("Added PVC %s:%s", pvc.Namespace, pvc.Name))
}

func (pvcHandler *PVCHandler) HandlePVCDelete(pvc *corev1.PersistentVolumeClaim) error {
	zap.S().Infof("Deleting PVC <Name: %v, Namespace: %v>", pvc.Name, pvc.Namespace)

	parsedPvc, err := parsers.ParsePvcToServerScheme(pvc)
	if err != nil {
		zap.S().Warnf("Error parsing PVC %s:%s, error: %v", pvc.Namespace, pvc.Name, err)
		return err
	}

	var response *http.Response

	switch parsedPvc.Scope.Type {
	case auth_client.SCOPETYPE_CLUSTER:
		response, err = pvcHandler.backendClient.K8sObjectTrackerClient.DeleteClusterWidePvc(
			pvcHandler.backendClient.Context(), config.Config().ClusterUuid, pvc.Name)

	case auth_client.SCOPETYPE_DEPARTMENT:
		response, err = pvcHandler.backendClient.K8sObjectTrackerClient.DeleteDepartmentWidePvc(
			pvcHandler.backendClient.Context(), config.Config().ClusterUuid, parsedPvc.Scope.Id, pvc.Name)

	case auth_client.SCOPETYPE_PROJECT:
		response, err = pvcHandler.backendClient.K8sObjectTrackerClient.DeleteProjectPvc(
			pvcHandler.backendClient.Context(), config.Config().ClusterUuid, parsedPvc.Scope.Id, pvc.Name)
	}

	return pvcHandler.backendClient.HandleHttpResponse(response, err, fmt.Sprintf("Deleted PVC %s:%s", pvc.Namespace, pvc.Name))
}

func (pvcHandler *PVCHandler) HandleSyncEvent(waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	zap.S().Infof("Start syncing PVCs")

	// get a list of all the PVCs in the cluster. Don't convert to var since this sends an empty array when
	// no data is in cluster vs null (which will result in bad request (400))
	pvcs := &corev1.PersistentVolumeClaimList{}
	err := pvcHandler.cachedReader.List(context.TODO(), pvcs)
	if err != nil {
		zap.S().Errorf("Failed listing PVCs, error: %v", err)
		return
	}

	// get a map of all the project namespaces
	namespacesMap := util.GetAllRunAiManagedNamespacesMap(pvcHandler.cachedReader)

	// initialize a map that arranges the PVCs by their scope
	pvcByScope := make(map[parsers.ScopeTypeAndId]*k8sobjectstracker.PVCs)
	pvcsOfAssets := make([]assetsswagger.AssetResourcesIdsItem, 0)

	for _, pvc := range pvcs.Items {
		if !util.IsUserExposedPvcForRunai(&pvc) && !util.IsAssetRelatedResource(&pvc.ObjectMeta) {
			continue
		}

		// if the PVC is an asset related resource, add it to the list of assets to sync
		if util.IsAssetRelatedResource(&pvc.ObjectMeta) {
			if assetId, ok := pvc.Labels[clustercommonconsts.RunaiAssetIdLabel]; ok {
				pvcsOfAssets = append(pvcsOfAssets, assetsswagger.AssetResourcesIdsItem{
					AssetId:    assetId,
					ResourceId: string(pvc.UID),
				})
			}
		}

		parsedPvc, err := parsers.ParsePvcToServerScheme(&pvc)
		if err != nil {
			zap.S().Warnf("Error parsing PVC %s, error; %v", pvc.Name, err)
			continue
		}

		if parsedPvc.Scope.Type == auth_client.SCOPETYPE_PROJECT {
			_, namespaceFound := namespacesMap[pvc.Namespace]
			if !namespaceFound {
				continue // project PVCs only in project related namespace (all others are replicates)
			}
		} else {
			if pvc.Namespace != v1.DefaultNamespace {
				continue // non-project PVCs only in runai namespace
			}
		}

		pvcToAdd := k8sobjectstracker.PVC{
			Name:              pvc.Name,
			Origin:            parsedPvc.CreationRequest.Origin,
			Issues:            parsedPvc.CreationRequest.Issues,
			StorageClassName:  parsedPvc.CreationRequest.StorageClassName,
			Phase:             parsedPvc.CreationRequest.Phase,
			ActualCapacity:    parsedPvc.CreationRequest.ActualCapacity,
			RequestedCapacity: parsedPvc.CreationRequest.RequestedCapacity,
		}

		if pvcsOfThisType, ok := pvcByScope[parsedPvc.Scope]; ok {
			pvcsOfThisType.Pvcs = append(pvcsOfThisType.Pvcs, pvcToAdd)
		} else {
			pvcByScope[parsedPvc.Scope] = &k8sobjectstracker.PVCs{
				Pvcs: []k8sobjectstracker.PVC{pvcToAdd},
			}
		}
	}

	if len(pvcsOfAssets) > 0 {
		_, err := pvcHandler.backendClient.AssetsClient.SyncAssetsResourcesIds(
			pvcHandler.backendClient.Context(), assetsswagger.AssetResourcesIdsSyncRequest{
				Items: pvcsOfAssets,
			})
		if err != nil {
			zap.S().Warnf("Error syncing assets resources ids, error: %v", err)
		}
	}

	for scope, pvcList := range pvcByScope {
		zap.S().Debugf("apply %d PVCs of %s %s", len(pvcList.Pvcs), scope.Type, scope.Id)

		var execError error
		var response *http.Response

		switch scope.Type {
		case auth_client.SCOPETYPE_CLUSTER:
			_, response, execError = pvcHandler.backendClient.K8sObjectTrackerClient.SyncClusterWidePvcs(
				pvcHandler.backendClient.Context(), config.Config().ClusterUuid, pvcList.Pvcs)

		case auth_client.SCOPETYPE_DEPARTMENT:
			_, response, execError = pvcHandler.backendClient.K8sObjectTrackerClient.SyncDepartmentWidePvcs(
				pvcHandler.backendClient.Context(), config.Config().ClusterUuid, scope.Id, pvcList.Pvcs)

		case auth_client.SCOPETYPE_PROJECT:
			_, response, execError = pvcHandler.backendClient.K8sObjectTrackerClient.SyncProjectPvcs(
				pvcHandler.backendClient.Context(), config.Config().ClusterUuid, scope.Id, pvcList.Pvcs)
		}

		if execError != nil {
			zap.S().Warnf("Error while syncing PVCs of %s %s, error: %v", scope.Type, scope.Id, execError)
		} else {
			zap.S().Debugf("HandlePVCSync of %s %s res %v ", scope.Type, scope.Id, response)
		}
	}
}

func (pvcHandler *PVCHandler) HandlePvcAssetAdd(pvc *corev1.PersistentVolumeClaim) error {
	assetId, ok := pvc.Labels[clustercommonconsts.RunaiAssetIdLabel]
	if !ok {
		zap.S().Infof("Failed to find assetId label on PVC: '%s', namespace: '%s'", pvc.Name, pvc.Namespace)
		return nil
	}

	_, err := pvcHandler.backendClient.AssetsClient.SyncAssetsResourcesIds(pvcHandler.backendClient.Context(), assetsswagger.AssetResourcesIdsSyncRequest{
		Items: []assetsswagger.AssetResourcesIdsItem{{
			AssetId:    assetId,
			ResourceId: string(pvc.UID),
		}},
	})
	if err != nil {
		zap.S().Warnf("Error syncing asset resources ids for PVC %s:%s, error: %v", pvc.Namespace, pvc.Name, err)
	}

	return nil
}
