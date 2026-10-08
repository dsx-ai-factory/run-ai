/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"
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
// Mirrors the StorageClass pattern: informer events drive per-item POST/PUT/DELETE
// for low-latency updates, while HandleSyncEvent performs a periodic full-set
// reconciliation that also prunes server-side rows no longer present in the cluster.
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

// HandleUpdateItem is the rate-limited callback used for update events: it
// re-reads the object from the cache (so we send the latest observed state,
// not whichever version was queued) and forwards it as a per-item update.
func (deviceClassHandler *DeviceClassHandler) HandleUpdateItem(key client.ObjectKey) error {
	deviceClass := &resourcev1.DeviceClass{}
	if err := deviceClassHandler.cachedReader.Get(context.Background(), key, deviceClass); err != nil {
		return err
	}
	return deviceClassHandler.HandleDeviceClassUpdate(deviceClass)
}

func (deviceClassHandler *DeviceClassHandler) HandleDeviceClassAdd(deviceClass *resourcev1.DeviceClass) error {
	zap.S().Infof("Adding DeviceClass < Name: %v >", deviceClass.Name)
	parsed := parsers.ParseDeviceClassToServerScheme(deviceClass)

	_, res, err := deviceClassHandler.backendClient.K8sObjectTrackerClient.CreateDeviceClass(
		deviceClassHandler.backendClient.Context(), parsed)
	if res != nil {
		_ = res.Body.Close()
	}

	return deviceClassHandler.backendClient.HandleHttpResponse(res, err, fmt.Sprintf("Added DeviceClass %s", parsed.DeviceClassName))
}

func (deviceClassHandler *DeviceClassHandler) HandleDeviceClassUpdate(deviceClass *resourcev1.DeviceClass) error {
	zap.S().Debugf("Updating DeviceClass < Name: %v >", deviceClass.Name)
	parsed := parsers.ParseDeviceClassToServerScheme(deviceClass)

	_, res, err := deviceClassHandler.backendClient.K8sObjectTrackerClient.UpdateDeviceClass(
		deviceClassHandler.backendClient.Context(), parsed)
	if res != nil {
		_ = res.Body.Close()
	}

	return deviceClassHandler.backendClient.HandleHttpResponse(res, err, fmt.Sprintf("Updated DeviceClass %s", parsed.DeviceClassName))
}

func (deviceClassHandler *DeviceClassHandler) HandleDeviceClassDelete(deviceClass *resourcev1.DeviceClass) error {
	zap.S().Debugf("Deleting DeviceClass < Name: %v >", deviceClass.Name)

	res, err := deviceClassHandler.backendClient.K8sObjectTrackerClient.DeleteDeviceClass(
		deviceClassHandler.backendClient.Context(), config.Config().ClusterUuid, deviceClass.Name)
	if res != nil {
		_ = res.Body.Close()
	}

	return deviceClassHandler.backendClient.HandleHttpResponse(res, err, fmt.Sprintf("Deleted DeviceClass %s", deviceClass.Name))
}

// HandleSyncEvent performs a full-set reconciliation. It lists every DeviceClass
// in the cluster and sends them in one request; the server upserts all and prunes
// anything no longer in the list (drift correction for missed per-item events).
func (deviceClassHandler *DeviceClassHandler) HandleSyncEvent(waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	zap.S().Infof("Start syncing DeviceClasses")

	deviceClassesList := resourcev1.DeviceClassList{}
	if err := deviceClassHandler.cachedReader.List(context.Background(), &deviceClassesList); err != nil {
		zap.S().Errorf("Failed listing device classes, error: %s", err.Error())
		return
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

	if err := deviceClassHandler.backendClient.HandleHttpResponse(res, httpErr, "Sync deviceClass data"); err != nil {
		zap.S().Errorf("Had an error while syncing DeviceClasses: %+v", err)
	}
}
