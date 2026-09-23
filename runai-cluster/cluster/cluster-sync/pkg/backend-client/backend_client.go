/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package backendclient

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"time"

	authclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	authgenerated "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	clustersvcswagger "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	orgUnit "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client"
	orgunitgenerated "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	asssets "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client"
	assetsgenerated "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	datavolumesclient "github.com/dsx-ai-factory/run-ai/runai-clients/datavolumes-service-client"
	datavolumesgenerated "github.com/dsx-ai-factory/run-ai/runai-clients/datavolumes-service-client/generated"
	k8sobjectstracker "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client"
	k8strackergenerated "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	policies "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client"
	policiesgenerated "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
	workloadswagger "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/metrics"
	"go.uber.org/zap"
)

type BackendClient struct {
	Client                 *assetsgenerated.APIClient
	AccessRulesClient      authclient.IAuthorizationExternalClient
	DatavolumesClient      datavolumesclient.IDatavolumesClient
	ClusterServiceClient   *clustersvcswagger.APIClient
	K8sObjectTrackerClient k8sobjectstracker.IK8SObjectTrackerExternalClient
	WorkloadsClient        *workloadswagger.APIClient
	PolicyClient           policies.IPolicyExternalClient
	AssetsClient           asssets.IAssetsExternalClient
	OrgUnitClient          orgUnit.IOrgUnitExternalClient
	basePath               string
	runaiBackendAvailable  bool
	lastSuccessfulLiveness time.Duration
	syncHandlers           []SyncHandler
	livenessLock           sync.RWMutex
	syncLock               sync.Mutex
	wasClusterInfoSynced   bool
	metricsExporter        *metrics.MetricsExporter
}

func (backendClient *BackendClient) Context() context.Context {
	return context.Background()
}
func (backendClient *BackendClient) IsRunaiBackendAvailable() bool {
	backendClient.livenessLock.RLock()
	defer backendClient.livenessLock.RUnlock()
	return backendClient.runaiBackendAvailable
}

func NewBackendClient(basePath string, httpClient *http.Client) *BackendClient {
	if httpClient == nil {
		httpClient = &http.Client{Transport: http.DefaultTransport}
	}
	assetsConfig := &assetsgenerated.Configuration{
		Servers:    assetsgenerated.ServerConfigurations{{URL: basePath}},
		HTTPClient: httpClient,
	}
	runaiBackendClient := &BackendClient{
		Client: assetsgenerated.NewAPIClient(assetsConfig),
		AccessRulesClient: authclient.NewAccessRulesClientWithConfig(&authgenerated.Configuration{
			Servers:    authgenerated.ServerConfigurations{{URL: basePath}},
			HTTPClient: httpClient,
		}),
		DatavolumesClient: datavolumesclient.NewDatavolumesClientWithConfig(&datavolumesgenerated.Configuration{
			Servers:    datavolumesgenerated.ServerConfigurations{{URL: basePath}},
			HTTPClient: httpClient,
		}),
		ClusterServiceClient: clustersvcswagger.NewAPIClient(&clustersvcswagger.Configuration{
			Servers: []clustersvcswagger.ServerConfiguration{
				{
					URL: basePath,
				},
			},
			HTTPClient: httpClient,
		}),
		K8sObjectTrackerClient: k8sobjectstracker.NewK8SObjectTrackerExternalClientWithConfig(&k8strackergenerated.Configuration{
			Servers:    k8strackergenerated.ServerConfigurations{{URL: basePath}},
			HTTPClient: httpClient,
		}),
		WorkloadsClient: workloadswagger.NewAPIClient(&workloadswagger.Configuration{
			Servers: []workloadswagger.ServerConfiguration{
				{
					URL: basePath,
				},
			},
			HTTPClient: httpClient,
		}),
		PolicyClient: policies.NewPoliciesExternalClientWithConfig(&policiesgenerated.Configuration{
			Servers:    policiesgenerated.ServerConfigurations{{URL: basePath}},
			HTTPClient: httpClient,
		}),
		AssetsClient: asssets.NewExternalClientWithConfig(assetsConfig),
		OrgUnitClient: orgUnit.NewOrgUnitExternalClientWithConfig(&orgunitgenerated.Configuration{
			Servers:    orgunitgenerated.ServerConfigurations{{URL: basePath}},
			HTTPClient: httpClient,
		}),
		runaiBackendAvailable: false,
		basePath:              basePath,
		wasClusterInfoSynced:  false,
		metricsExporter:       metrics.GetMetricsExporter(),
	}

	//This disables the authentication of certificates in all http requests. This is insecure - do this for local-env only.
	if !config.Config().SSLCertificateVerification {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	runaiBackendClient.metricsExporter.Init()

	return runaiBackendClient
}

func (backendClient *BackendClient) RegisterSyncHandler(syncHandler SyncHandler) {
	if syncHandler == nil {
		return
	}
	if backendClient.syncHandlers == nil {
		backendClient.syncHandlers = make([]SyncHandler, 0, 2)
	}

	backendClient.syncHandlers = append(backendClient.syncHandlers, syncHandler)

}

// Sync is the operation that the service does to sync backend after it becomes available.
// For each resource there is a handler that is responsible to reflect the state of the resource in k8s and send it to backend.
func (backendClient *BackendClient) handleSync() {
	// Don't start new sync before old one ends
	backendClient.syncLock.Lock()
	defer backendClient.syncLock.Unlock()
	zap.S().Info("Start Sync requests to the cluster")
	wg := sync.WaitGroup{}
	for _, handler := range backendClient.syncHandlers {
		wg.Add(1)
		go handler.HandleSyncEvent(&wg)
	}
	wg.Wait()
	zap.S().Info("Finished all sync requests to the cluster")
}

// The function does the following things:
// 1) set if backend is available.
// 2) If backend availability changes to true, sync backend
// 3) if backend is down for too long make the service panic.
func (backendClient *BackendClient) setBackendAvailability(isAvailable bool, timeTag time.Duration) {
	backendClient.livenessLock.Lock()
	defer backendClient.livenessLock.Unlock()
	// Ignore liveness requests from the past
	if timeTag < backendClient.lastSuccessfulLiveness {
		return
	}
	// Handle state change
	if backendClient.runaiBackendAvailable != isAvailable {
		zap.S().Infof("Status of availability of runai had changed to: %t", isAvailable)
		if isAvailable {
			go backendClient.handleSync() // sync backend with cluster data
		}
	}
	backendClient.runaiBackendAvailable = isAvailable
	// Update lastSuccessfulLiveness
	if isAvailable {
		backendClient.lastSuccessfulLiveness = timeTag
	} else {
		backendDownDuration := time.Duration(time.Now().UnixMilli()) - backendClient.lastSuccessfulLiveness
		backendNotRespondedForTooLong := backendDownDuration*time.Millisecond > time.Second*time.Duration(config.MaxAllowedWaitTimeInSec())
		if backendClient.lastSuccessfulLiveness > 0 &&
			backendNotRespondedForTooLong {
			zap.S().Fatalf("Backend service was unavailable for too long. service terminating...")
		}
	}
}

func (backendClient *BackendClient) SyncPeriodically(minuteInterval time.Duration, stop <-chan struct{}) {
	ticker := time.NewTicker(minuteInterval * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				go backendClient.handleSync()
			case <-stop:
				ticker.Stop()
				return
			}
		}
	}()
}

func (backendClient *BackendClient) HandleHttpResponse(res *http.Response, err error, baseMsg string) error {
	_, httpError := backendClient.HandleHttpResponseGetStatus(res, err, baseMsg)
	return httpError
}

func (backendClient *BackendClient) HandleHttpResponseGetStatus(res *http.Response, err error, baseMsg string) (int, error) {
	if err != nil {
		if res != nil {
			return res.StatusCode, fmt.Errorf("Had error while sending http request.\n URL: %s.\n Status: %s.\n Error: %v\n", res.Request.URL, res.Status, err) //nolint:staticcheck
		}
		return 0, err
	}
	if res != nil && res.StatusCode >= http.StatusBadRequest {
		return res.StatusCode, fmt.Errorf("Had error while sending http request.\n URL: %s.\n Status: %s\n", res.Request.URL, res.Status) //nolint:staticcheck
	}
	zap.S().Infof("%s successfully.", baseMsg)
	return 0, nil
}

// StartSendingLivenessPeriodically should be called after all handlers registered with RegisterSyncHandler.
func (backendClient *BackendClient) StartSendingLivenessPeriodically(stopCh <-chan struct{}) {
	ticker := time.NewTicker(time.Second * (time.Duration(config.Config().LivenessIntervalSeconds)))
	defer ticker.Stop()
	//This loop sends liveness every Liveness interval (from config), and enters on first iteration immediately
loop:
	for {
		backendClient.sendLiveness()
		select {
		case <-ticker.C:
			continue
		case <-stopCh:
			break loop
		}
	}
}

func (backendClient *BackendClient) sendLiveness() {
	var livenessErr error
	livenessErr = nil
	requestTimeTag := time.Now().UnixMilli()
	livenessRequest := clustersvcswagger.NewClusterLivenessSyncRequest(requestTimeTag)
	res, err := backendClient.ClusterServiceClient.ClustersAPI.UpdateClusterLiveness(
		backendClient.Context(), config.Config().ClusterUuid,
	).ClusterLivenessSyncRequest(*livenessRequest).Execute()
	if res != nil {
		_ = res.Body.Close()
	}
	if err != nil || res.StatusCode >= http.StatusBadRequest {
		livenessErr = backendClient.HandleHttpResponse(res, err, "Sent liveness")
		if livenessErr != nil {
			zap.S().Errorf("Had an error while sending liveness: %+v", livenessErr)
		}
	} else {
		zap.S().Debug("Sent Liveness Successfully")
	}
	backendAvailable := livenessErr == nil
	backendClient.setBackendAvailability(backendAvailable, time.Duration(requestTimeTag))

}

// StartSyncingClusterInfo should be removed once we find a better way to export a health metric.
func (backendClient *BackendClient) StartSyncingClusterInfo(stopCh <-chan struct{}) {
	ticker := time.NewTicker(time.Second * (time.Duration(config.Config().ClusterInfoSyncIntervalSeconds)))
	defer ticker.Stop()
loop:
	for {
		backendClient.metricsExporter.IncBackendSyncCounter(metrics.ClusterInfoPush)
		select {
		case <-ticker.C:
			continue
		case <-stopCh:
			break loop
		}
	}
}
