/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	"context"
	"fmt"
	dwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	iwv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
	"time"

	workloadconfig "github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	workloadMigratorConfigName         = "workload-controller"
	workloadMigratorConfigNamespace    = "runai"
	defaultGracePeriod                 = 30 * time.Second
	webserverGracePeriodTimeoutMessage = "webserver did not start during the grace period"
)

var log = logger.Log.WithName("migration")

type WorkloadMigration struct {
	k8sClient                client.Client
	migrationConfigMapName   string
	migrationConfigNamespace string
	migrationSteps           []Migrator
	webserverGraceDelay      time.Duration
}

type WebhookServer interface {
	Start(ctx context.Context) error
	StartedChecker() healthz.Checker
}

type Migrator interface {
	Migrate(ctx context.Context) error
	String() string
	RequiresWebhooks() bool
}

func (wm *WorkloadMigration) GetMigrationConfigMap(ctx context.Context) (corev1.ConfigMap, error) {
	configMap := corev1.ConfigMap{}
	err := wm.k8sClient.Get(ctx, client.ObjectKey{
		Namespace: wm.migrationConfigNamespace,
		Name:      wm.migrationConfigMapName,
	}, &configMap)
	return configMap, err
}

func (wm *WorkloadMigration) Run(ctx context.Context, webServer WebhookServer) error {
	log.Info("Running migration steps")
	timeNow := time.Now()
	defer func() {
		log.V(4).Info("migration took", "time", time.Since(timeNow))
	}()
	configMap, err := wm.GetMigrationConfigMap(ctx)
	if err != nil {
		log.Error(err, "failed to get migration config map, running migration steps without config map")
	}
	var create bool
	if errors.IsNotFound(err) {
		create = true
	}
	defer func() {
		if err = wm.updateMigrationConfigMap(ctx, configMap, create); err != nil {
			log.Error(err, "failed to update migration config map")
		}
	}()
	if configMap.Data == nil {
		configMap.Data = make(map[string]string)
	}

	startedWebserver := false
	if wm.webserverGraceDelay == 0 {
		wm.webserverGraceDelay = defaultGracePeriod
	}
	webserverCtx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		if startedWebserver {
			time.Sleep(wm.webserverGraceDelay)
		}
	}()

	for _, migrator := range wm.migrationSteps {
		migratorCp := migrator
		migratorName := migratorCp.String()
		if _, ok := configMap.Data[migratorName]; !ok {
			log.V(4).Info("Running", migratorName)

			if migratorCp.RequiresWebhooks() && !startedWebserver {
				err = wm.startWebServer(webServer, webserverCtx, webServer.StartedChecker())
				if err != nil {
					return err
				}
				startedWebserver = true
			}

			err = migratorCp.Migrate(ctx)
			if err != nil {
				log.Error(err, "failed to run one of the migration step", "step",
					migratorName)
				return err
			}
			configMap.Data[migratorName] = "true"

		} else {
			log.V(4).Info("Skip migration step, has already been executed.",
				"step", migratorName)
		}
	}
	log.Info("migration steps completed successfully")
	return nil
}

func (wm *WorkloadMigration) startWebServer(webServer WebhookServer, webserverCtx context.Context, webhookHealthFn healthz.Checker) error {
	startupErrChan := make(chan error)

	go func() {
		err := webServer.Start(webserverCtx)
		if err != nil {
			log.Error(err, "failed to start webserver")
			startupErrChan <- err
		}
	}()

	log.V(4).Info("waiting for the webserver to start")

	ctxTimeout, cancel := context.WithTimeout(context.Background(), wm.webserverGraceDelay)
	defer cancel()
	go func() {
		for {
			select {
			case <-ctxTimeout.Done():
				startupErrChan <- fmt.Errorf(webserverGracePeriodTimeoutMessage)
				return
			default:
				if err := webhookHealthFn(&http.Request{}); err == nil {
					log.V(4).Info("webserver started and reachable")
					startupErrChan <- nil
					return
				}
			}
		}
	}()

	return <-startupErrChan
}

func (wm *WorkloadMigration) updateMigrationConfigMap(ctx context.Context, configMap corev1.ConfigMap, create bool) error {
	if create {
		configMap.Name = wm.migrationConfigMapName
		configMap.Namespace = wm.migrationConfigNamespace
		return wm.k8sClient.Create(ctx, &configMap)
	}
	return wm.k8sClient.Update(ctx, &configMap)
}

func NewWorkloadMigration(k8sClient client.Client, clusterVersion string, migrationMaxWorkers int) (*WorkloadMigration, error) {
	suspend := NewSuspendWorkloads(k8sClient, migrationMaxWorkers)
	externalCleaner := NewNetworkCleaner(k8sClient, workloadconfig.IsOpenshift())
	labels, err := NewLabelWorkloads(k8sClient, clusterVersion, []WorkloadList{&twv2.TrainingWorkloadList{}, &iwv2a1.InteractiveWorkloadList{}, &dwv2.DistributedWorkloadList{}}, nil, migrationMaxWorkers)

	if err != nil {
		return nil, err
	}

	return &WorkloadMigration{
		k8sClient:                k8sClient,
		migrationConfigMapName:   workloadMigratorConfigName,
		migrationConfigNamespace: workloadMigratorConfigNamespace,
		migrationSteps:           []Migrator{suspend, externalCleaner, labels},
		webserverGraceDelay:      defaultGracePeriod,
	}, nil
}
