/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"sync"

	trackerclient "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	runaiBackendClient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	"go.uber.org/zap"
	resourcev1 "k8s.io/api/resource/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DeviceClassHandler syncs cluster DeviceClasses to the k8s-objects-tracker.
// The tracker only supports a full-set sync (there are no incremental create/update/delete
// endpoints), so every change triggers a re-sync of the complete set.
type DeviceClassHandler struct {
	backendClient *runaiBackendClient.BackendClient
	cachedReader  client.Reader
}

func NewDeviceClassHandler(cachedReader client.Reader, runaiBackendClient *runaiBackendClient.BackendClient) *DeviceClassHandler {
	return &DeviceClassHandler{
		cachedReader:  cachedReader,
		backendClient: runaiBackendClient,
	}
}

// HandleUpdateItem is invoked (via the rate limiter) on any DeviceClass change. The key is
// ignored on purpose: the tracker only accepts the full set, so we re-sync everything.
func (deviceClassHandler *DeviceClassHandler) HandleUpdateItem(_ client.ObjectKey) error {
	return deviceClassHandler.sync()
}

func (deviceClassHandler *DeviceClassHandler) HandleSyncEvent(waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	zap.S().Infof("Start syncing DeviceClasses")

	if err := deviceClassHandler.sync(); err != nil {
		zap.S().Errorf("Had an error while syncing DeviceClasses: %+v", err)
	}
}

func (deviceClassHandler *DeviceClassHandler) sync() error {
	deviceClassesList := resourcev1.DeviceClassList{}
	if err := deviceClassHandler.cachedReader.List(context.Background(), &deviceClassesList); err != nil {
		zap.S().Errorf("Failed listing device classes, error: %s", err.Error())
		return err
	}

	// Initialize with an empty (non-nil) slice so an empty cluster sends [] rather than null.
	deviceClasses := trackerclient.DeviceClasses{Items: []trackerclient.DeviceClass{}}
	for _, deviceClass := range deviceClassesList.Items {
		deviceClasses.Items = append(deviceClasses.Items, parsers.ParseDeviceClassToServerScheme(&deviceClass))
	}

	_, res, httpErr := deviceClassHandler.backendClient.K8sObjectTrackerClient.SyncDeviceClass(
		deviceClassHandler.backendClient.Context(), config.Config().ClusterUuid, deviceClasses.Items)
	if res != nil {
		_ = res.Body.Close()
	}

	return deviceClassHandler.backendClient.HandleHttpResponse(res, httpErr, "Sync deviceClass data")
}
