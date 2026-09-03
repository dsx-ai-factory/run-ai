/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pkg

import (
	"context"
	"sync"
	"time"

	configmaps "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/config-maps"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/config"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	multierror "github.com/hashicorp/go-multierror"

	"github.com/rs/zerolog/log"

	runaiBackendClient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/handlers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/metrics"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewInteractivePoliciesHandler and NewTrainingPoliciesHandler are added in preInitHandler
// if PolicySyncEnabled is true
var handlerInitFuncs = []handlers.ResourceHandlerInitFunc{
	handlers.NewDepartmentsHandler,
	handlers.NewProjectsHandler,
	handlers.NewAccessRulesHandler,
	handlers.NewDatavolumesHandler,
	handlers.NewPriorityClassesHandler,
	handlers.NewWorkloadsConfigHandler,
	handlers.NewInteractiveWorkloadsHandler,
	handlers.NewTrainingWorkloadsHandler,
	handlers.NewDistributedWorkloadsHandler,
	handlers.NewInferenceWorkloadsHandler,
	handlers.NewDistributedInferenceWorkloadsHandler,
	handlers.NewCredentialAssetsHandler,
	handlers.NewDatasourceAssetsHandler,
	handlers.NewNetworkTopologiesHandler,
	handlers.NewWorkloadsV2Handler,
	handlers.NewApplicationsHandler,
	handlers.NewKaiNodePoolsHandler,
}

type RunaiAgent struct {
	runaiClient   client.Client
	cachedClient  client.Reader
	backendClient *runaiBackendClient.BackendClient

	metricsExporter *metrics.MetricsExporter

	resourceHandlers []handlers.ResourceHandler

	syncIntervalMillis int

	syncLock sync.Mutex
}

func NewRunaiAgent(
	runaiBackendClient *runaiBackendClient.BackendClient,
	runaiClient client.Client,
	cachedClient client.Reader,
	syncIntervalMillis int) *RunaiAgent {
	return &RunaiAgent{
		runaiClient:        runaiClient,
		cachedClient:       cachedClient,
		backendClient:      runaiBackendClient,
		metricsExporter:    metrics.GetMetricsExporter(),
		resourceHandlers:   make([]handlers.ResourceHandler, 0, len(handlerInitFuncs)),
		syncIntervalMillis: syncIntervalMillis,
	}
}

func (agent *RunaiAgent) Start(ctx context.Context) {
	agent.preInitHandler()

	agent.initResourceHandlers()

	agent.metricsExporter.Init()

	log.Info().Msgf("Agent will run every <%d> seconds", agent.syncIntervalMillis/1000)
	agent.syncPeriodically(time.Duration(agent.syncIntervalMillis), ctx.Done())
}

func (agent *RunaiAgent) preInitHandler() {
	_, err := configmaps.CreateAgentConfigIfNotExists(agent.runaiClient, agent.cachedClient)
	if err != nil {
		log.Fatal().Msgf("Failed to create agent config, error: %s", err.Error())
	}

	if config.Config().PolicySyncEnabled {
		handlerInitFuncs = append(handlerInitFuncs, handlers.NewInteractivePoliciesHandler)
		handlerInitFuncs = append(handlerInitFuncs, handlers.NewTrainingPoliciesHandler)
		handlerInitFuncs = append(handlerInitFuncs, handlers.NewDistributedPoliciesHandler)
		handlerInitFuncs = append(handlerInitFuncs, handlers.NewInferencePoliciesHandler)
	}
}

func (agent *RunaiAgent) initResourceHandlers() {
	for _, handlerInitFunc := range handlerInitFuncs {
		resourceHandler, err := handlerInitFunc(agent.backendClient, agent.runaiClient, agent.cachedClient)
		if err != nil {
			log.Error().Msgf("Failed to init resource handler, error: %s", err.Error())
		} else {
			agent.resourceHandlers = append(agent.resourceHandlers, resourceHandler)
		}
	}

	if len(agent.resourceHandlers) == 0 {
		log.Fatal().Msgf("Failed to init all resource handlers, aborting...")
	}
}

func (agent *RunaiAgent) syncPeriodically(millisInterval time.Duration, stopCh <-chan struct{}) {
	go agent.handleSync()

	ticker := time.NewTicker(millisInterval * time.Millisecond)

	go func() {
		for {
			select {
			case <-ticker.C:
				go agent.handleSync()
			case <-stopCh:
				ticker.Stop()
				return
			}
		}
	}()
}

func (agent *RunaiAgent) handleSync() {
	if !agent.syncLock.TryLock() {
		log.Info().Msgf("Agent sync is still in progress - not starting new sync")
		return
	}
	defer agent.syncLock.Unlock()

	log.Debug().Msgf("Agent syncing")

	err := agent.handleResources()
	if err != nil {
		return
	}

	agent.metricsExporter.IncBackendSyncCounter(metrics.Pull)
}

func (agent *RunaiAgent) handleResources() error {
	var err error

	if agent.isHelmBeingDeleted() {
		log.Info().Msgf("Helm is being deleted, skipping sync")
		return nil
	}

	for _, resourceHandler := range agent.resourceHandlers {
		// Measure and record the duration of the resource sync operation
		timer := agent.metricsExporter.StartSyncTimer(resourceHandler.MetricLabel())
		innerErr := resourceHandler.Handle()
		if innerErr != nil {
			err = multierror.Append(err, innerErr)

			agent.metricsExporter.IncSyncFailure(resourceHandler.MetricLabel())
		}

		timer.ObserveDuration()
	}

	if err != nil {
		log.Error().Msgf("Failed to handle resources, error: %s", err.Error())
		return err
	}

	log.Debug().Msgf("Successfully handled all resources")
	return nil
}

func (agent *RunaiAgent) isHelmBeingDeleted() bool {
	clusterResource, err := agent.getClusterResource()
	if err != nil {
		return false
	}

	return !clusterResource.DeletionTimestamp.IsZero()
}

func (agent *RunaiAgent) getClusterResource() (*runaiv1.Cluster, error) {
	cluster := &runaiv1.Cluster{}
	objectKey := types.NamespacedName{Namespace: runaiv1.DefaultNamespace, Name: runaiv1.ClusterName}
	err := agent.runaiClient.Get(context.Background(), objectKey, cluster)
	if err != nil {
		log.Error().Msgf("Failed getting cluster resource <%v>, error: %v", runaiv1.ClusterName, err.Error())
		return nil, err
	}
	return cluster, nil
}
