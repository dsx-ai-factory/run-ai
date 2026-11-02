/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"go.uber.org/zap"
	"sync"

	backendclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PodGroupHandler struct {
	cachedReader       client.Reader
	runaiBackendClient *backendclient.BackendClient

	// namespace -> UIDs of already synced pod groups
	syncedObjects      map[string]map[types.UID]bool
	syncedObjectsMutex sync.RWMutex
}

// NewPodGroupHandler creates a new PodGroupHandler.
// Not deleting completely as soon there will be a need to handle pod groups again (only for ids...).
func NewPodGroupHandler(cachedClient cachedclient.CachedClient, runaiBackendClient *backendclient.BackendClient) *PodGroupHandler {
	return &PodGroupHandler{
		cachedReader:       cachedClient,
		runaiBackendClient: runaiBackendClient,
		syncedObjects:      map[string]map[types.UID]bool{},
		syncedObjectsMutex: sync.RWMutex{},
	}
}

func (handler *PodGroupHandler) HandleUpdateItem(key client.ObjectKey) error {
	podGroup := &kaiv2alpha2.PodGroup{}
	err := handler.cachedReader.Get(context.Background(), key, podGroup)
	if err != nil {
		return err
	}

	return handler.HandleUpdate(podGroup)
}

func (handler *PodGroupHandler) HandleAdd(podGroup *kaiv2alpha2.PodGroup) error {
	if WasSynced(podGroup.UID, podGroup.Namespace, handler.syncedObjects, &handler.syncedObjectsMutex) {
		zap.S().Infof("PodGroup <%s/%s> was already synced - not adding", podGroup.Namespace, podGroup.Name)
		return nil
	}

	// currently doing nothing
	return nil
}

func (handler *PodGroupHandler) HandleDelete(podGroup *kaiv2alpha2.PodGroup) error {
	// Don't update with last podGroup data since currently (16/12/21) data is missing upon deletion (no image name + annotations)
	// currently doing nothing
	return nil
}

func (handler *PodGroupHandler) HandleUpdate(podGroup *kaiv2alpha2.PodGroup) error {
	// currently doing nothing
	return nil
}

func (handler *PodGroupHandler) HandleSyncEvent(waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	zap.S().Infof("Start PodGroups Sync")

	UpdateNamespaceHandlerWrapper(handler.handleUpdateInNamespace, handler.cachedReader, "pod groups")
}

func (handler *PodGroupHandler) handleUpdateInNamespace(namespace *corev1.Namespace) {
	// currently doing nothing
}
