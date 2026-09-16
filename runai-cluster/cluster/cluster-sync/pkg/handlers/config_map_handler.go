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

	"go.uber.org/zap"

	auth_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	k8sobjectstracker "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	runaiBackendClient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/util"

	v1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ConfigMapHandler handles ConfigMap events
type ConfigMapHandler struct {
	backendClient *runaiBackendClient.BackendClient
	cachedReader  client.Reader
}

// NewConfigMapHandler returns a new ConfigMap handler
func NewConfigMapHandler(cachedReader client.Reader, runaiBackendClient *runaiBackendClient.BackendClient) *ConfigMapHandler {
	return &ConfigMapHandler{
		cachedReader:  cachedReader,
		backendClient: runaiBackendClient,
	}
}

func (h *ConfigMapHandler) HandleUpdateItem(key client.ObjectKey) error {
	configMap := &corev1.ConfigMap{}
	err := h.cachedReader.Get(context.Background(), key, configMap)
	if err != nil {
		return err
	}

	return h.HandleConfigMapUpdate(configMap)
}

// HandleConfigMapAdd handles a add ConfigMap event
func (h *ConfigMapHandler) HandleConfigMapAdd(configMap *corev1.ConfigMap) error {
	zap.S().Infof("Adding ConfigMap <Name: %v, Namespace: %v>", configMap.Name, configMap.Namespace)
	return h.handleConfigMapAddUpdate(configMap)
}

func (h *ConfigMapHandler) HandleConfigMapUpdate(configMap *corev1.ConfigMap) error {
	zap.S().Infof("Updating ConfigMap <Name: %v, Namespace: %v>", configMap.Name, configMap.Namespace)
	return h.handleConfigMapAddUpdate(configMap)
}

func (h *ConfigMapHandler) handleConfigMapAddUpdate(configMap *corev1.ConfigMap) error {
	parsedConfigMap, err := parsers.ParseConfigMapToServerScheme(configMap)
	if err != nil {
		zap.S().Warnf("Error parsing ConfigMap %s:%s, error: %v", configMap.Namespace, configMap.Name, err)
		return err
	}

	var response *http.Response

	clusterId := config.Config().ClusterUuid
	switch parsedConfigMap.Scope.Type {
	case auth_client.SCOPETYPE_CLUSTER:
		_, response, err = h.backendClient.K8sObjectTrackerClient.CreateClusterWideConfigmap(
			h.backendClient.Context(), clusterId, configMap.Name, parsedConfigMap.CreationRequest.Issues)
		if response != nil {
			_ = response.Body.Close()
		}

	case auth_client.SCOPETYPE_DEPARTMENT:
		_, response, err = h.backendClient.K8sObjectTrackerClient.CreateDepartmentsWideConfigmap(
			h.backendClient.Context(), clusterId, parsedConfigMap.Scope.Id, configMap.Name, parsedConfigMap.CreationRequest.Issues)
		if response != nil {
			_ = response.Body.Close()
		}

	case auth_client.SCOPETYPE_PROJECT:
		_, response, err = h.backendClient.K8sObjectTrackerClient.CreateProjectsConfigmap(
			h.backendClient.Context(), clusterId, parsedConfigMap.Scope.Id, configMap.Name, parsedConfigMap.CreationRequest.Issues)
		if response != nil {
			_ = response.Body.Close()
		}
	}

	return h.backendClient.HandleHttpResponse(response, err, fmt.Sprintf("Added ConfigMap %s:%s", configMap.Namespace, configMap.Name))
}

// HandleConfigMapDelete handles a delete ConfigMap event
func (h *ConfigMapHandler) HandleConfigMapDelete(configMap *corev1.ConfigMap) error {
	zap.S().Infof("Deleting ConfigMap <Name: %v, Namespace: %v>", configMap.Name, configMap.Namespace)

	parsedConfigMap, err := parsers.ParseConfigMapToServerScheme(configMap)
	if err != nil {
		zap.S().Warnf("Error parsing ConfigMap %s:%s, error: %v", configMap.Namespace, configMap.Name, err)
		return err
	}

	var response *http.Response

	switch parsedConfigMap.Scope.Type {
	case auth_client.SCOPETYPE_CLUSTER:
		response, err = h.backendClient.K8sObjectTrackerClient.DeleteClusterWideConfigmap(
			h.backendClient.Context(), config.Config().ClusterUuid, configMap.Name)
		if response != nil {
			_ = response.Body.Close()
		}

	case auth_client.SCOPETYPE_DEPARTMENT:
		response, err = h.backendClient.K8sObjectTrackerClient.DeleteDepartmentsWideConfigmap(
			h.backendClient.Context(), config.Config().ClusterUuid, parsedConfigMap.Scope.Id, configMap.Name)
		if response != nil {
			_ = response.Body.Close()
		}

	case auth_client.SCOPETYPE_PROJECT:
		response, err = h.backendClient.K8sObjectTrackerClient.DeleteProjectsConfigmap(
			h.backendClient.Context(), config.Config().ClusterUuid, parsedConfigMap.Scope.Id, configMap.Name)
		if response != nil {
			_ = response.Body.Close()
		}
	}

	return h.backendClient.HandleHttpResponse(response, err, fmt.Sprintf("Deleted ConfigMap %s:%s", configMap.Namespace, configMap.Name))
}

// HandleSyncEvent handles a sync ConfigMap event
func (h *ConfigMapHandler) HandleSyncEvent(waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	zap.S().Infof("Start syncing ConfigMaps")

	// get a list of all the ConfigMaps in the cluster. Don't convert to var since this sends an empty array when
	// no data is in cluster vs null (which will result in bad request (400))
	configMaps := &corev1.ConfigMapList{}
	err := h.cachedReader.List(context.TODO(), configMaps,
		client.HasLabels{config.Config().ConfigMapToSyncLabel})
	if err != nil {
		zap.S().Errorf("Failed listing ConfigMapss, error: %v", err)
		return
	}

	// get a map of all the project namespaces
	namespacesMap := util.GetAllRunAiManagedNamespacesMap(h.cachedReader)

	// initialize a map that arranges the ConfigMaps by their scope
	configMapByScope := make(map[parsers.ScopeTypeAndId]*k8sobjectstracker.ConfigMaps)

	for _, configMap := range configMaps.Items {
		parsedConfigMap, err := parsers.ParseConfigMapToServerScheme(&configMap)
		if err != nil {
			zap.S().Warnf("Error parsing ConfigMap %s:%s, error: %v", configMap.Namespace, configMap.Name, err)
			continue
		}

		if parsedConfigMap.Scope.Type == auth_client.SCOPETYPE_PROJECT {
			_, namespaceFound := namespacesMap[configMap.Namespace]
			if !namespaceFound {
				continue // project ConfigMaps only in project related namespace (all others are replicates)
			}
		} else {
			if configMap.Namespace != v1.DefaultNamespace {
				continue // non-project ConfigMaps only in runai namespace
			}
		}

		configMapToAdd := k8sobjectstracker.ConfigMap{
			Name:   configMap.Name,
			Issues: util.ExtractIssuesFromObject(configMap.GetAnnotations()),
		}

		if configMapsOfThisType, ok := configMapByScope[parsedConfigMap.Scope]; ok {
			configMapsOfThisType.Configmaps = append(configMapsOfThisType.Configmaps, configMapToAdd)
		} else {
			configMapByScope[parsedConfigMap.Scope] = &k8sobjectstracker.ConfigMaps{
				Configmaps: []k8sobjectstracker.ConfigMap{configMapToAdd},
			}
		}
	}

	for scope, configMapList := range configMapByScope {
		zap.S().Debugf("apply %d ConfigMaps of %s %s", len(configMapList.Configmaps), scope.Type, scope.Id)

		var execError error
		var response *http.Response

		switch scope.Type {
		case auth_client.SCOPETYPE_CLUSTER:
			_, response, execError = h.backendClient.K8sObjectTrackerClient.SyncClusterWideConfigmaps(
				h.backendClient.Context(), config.Config().ClusterUuid, configMapList.Configmaps)
			if response != nil {
				_ = response.Body.Close()
			}

		case auth_client.SCOPETYPE_DEPARTMENT:
			_, response, execError = h.backendClient.K8sObjectTrackerClient.SyncDepartmentsWideConfigmap(
				h.backendClient.Context(), config.Config().ClusterUuid, scope.Id, configMapList.Configmaps)
			if response != nil {
				_ = response.Body.Close()
			}

		case auth_client.SCOPETYPE_PROJECT:
			_, response, execError = h.backendClient.K8sObjectTrackerClient.SyncProjectsConfigmap(
				h.backendClient.Context(), config.Config().ClusterUuid, scope.Id, configMapList.Configmaps)
			if response != nil {
				_ = response.Body.Close()
			}
		}

		if execError != nil {
			zap.S().Warnf("Error while syncing ConfigMaps of %s %s, error: %v", scope.Type, scope.Id, execError)
		} else {
			zap.S().Infof("HandleSyncEvent of %s %s res %v ", scope.Type, scope.Id, response)
		}
	}
}
