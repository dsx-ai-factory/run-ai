/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/application-controller/pkg/deployers/helm_deployer"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	helmWorkerCount = 5
	helmJobBufSize  = 100
)

type HelmJob struct {
	AppKey types.NamespacedName
	Task   helm_deployer.Task
}

type HelmWorkers struct {
	// channel to queue jobs for execution
	jobsChannel chan HelmJob

	// wait group to synchronize the workers
	waitGroup sync.WaitGroup

	// keep track on the running tasks.
	// key: types.NamespacedName,
	// value: task action.
	runningTasks sync.Map
}

// IsHelmTaskAlive returns whether a Helm task for this app is currently queued or running in the worker pool.
func (hw *HelmWorkers) IsHelmTaskAlive(appKey types.NamespacedName) (bool, helm_deployer.TaskActionType) {
	if v, ok := hw.runningTasks.Load(appKey); ok {
		return true, v.(helm_deployer.TaskActionType)
	}
	return false, ""
}

func (c *ApplicationController) helmWorker(ctx context.Context, jobs <-chan HelmJob) {
	defer c.helmWorkers.waitGroup.Done()

	for job := range jobs {
		appKey := job.AppKey

		app := &v1alpha1.Application{}

		err := c.client.Get(ctx, appKey, app)
		if err != nil {
			c.helmWorkers.runningTasks.Delete(appKey)
			if ctx.Err() != nil {
				log.Warn().Msgf("context cancelled while fetching %s, stopping worker", appKey)
				break
			}
			log.Warn().Err(err).Msgf("failed to get application %s, skipping job", appKey)
			continue
		}

		status := helm_deployer.NewAppStatus().TaskStarted(job.Task)
		c.updateHelmStatus(ctx, appKey, status)

		status = job.Task.Execute(ctx, func(progress *helm_deployer.AppStatus) {
			c.updateHelmStatus(ctx, appKey, progress)
		})

		c.updateHelmStatus(ctx, appKey, status)
		c.helmWorkers.runningTasks.Delete(appKey)
	}
}

// reserve validates that a task can be submitted and registers it in runningTasks.
// On success the caller must call dispatch to put the job in the worker channel.
func (hw *HelmWorkers) reserve(appKey types.NamespacedName, task helm_deployer.Task) error {
	if hw.jobsChannel == nil {
		err := fmt.Errorf("helm job channel is inactive, cannot submit %s task for %s", task.Action(), appKey.String())
		log.Error().Err(err).Str("appKey", appKey.String()).Msg("failed to submit helm task")
		return err
	}

	if _, loaded := hw.runningTasks.LoadOrStore(appKey, task.Action()); loaded {
		return fmt.Errorf("helm task already queued or running for %s", appKey.String())
	}

	return nil
}

// dispatch sends a reserved job to the worker channel.
// reserve must succeed before calling dispatch.
func (hw *HelmWorkers) dispatch(appKey types.NamespacedName, task helm_deployer.Task) error {
	select {
	case hw.jobsChannel <- HelmJob{AppKey: appKey, Task: task}:
		log.Info().Msgf("queued %s task of %s", task.Action(), appKey.String())
		return nil
	default:
		hw.runningTasks.Delete(appKey)
		return fmt.Errorf("helm job queue full for %s, will retry", appKey.String())
	}
}
