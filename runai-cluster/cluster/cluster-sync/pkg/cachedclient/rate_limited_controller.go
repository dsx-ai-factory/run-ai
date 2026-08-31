/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cachedclient

import (
	"time"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/metrics"
	"golang.org/x/time/rate"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/config"
	"go.uber.org/zap"
	"k8s.io/client-go/util/workqueue"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	rateLimiterBaseDelay = time.Millisecond
	rateLimit            = 30
	rateBursts           = 90
)

type Handler interface {
	HandleUpdateItem(key client.ObjectKey) error
}

type RateLimitedController struct {
	minWaitRateLimiter *MinWaitRateLimiter[client.ObjectKey]
	queue              workqueue.TypedRateLimitingInterface[client.ObjectKey]
	resourceName       RateLimiterName
	handler            Handler
	workers            int
}

type rateLimiterConfigurationSettings struct {
	// it's either regular scale or "workloads" scale which is bigger -
	// for example, workload/pods/etc.. require large scale,
	// projects/nodepools/etc.. require default scale
	workloadsScale bool
	// if it's a deprecated resource, we set the num of workers to 1
	deprecated bool
}

type RateLimiterName string

var (
	ClusterInfoRL                   RateLimiterName = "ClusterInfo"
	ConfigMapRL                     RateLimiterName = "ConfigMap"
	DatavolumeRL                    RateLimiterName = "Datavolume"
	AccessRuleRL                    RateLimiterName = "AccessRule"
	ExternalWorkloadRL              RateLimiterName = "ExternalWorkload"
	TrainingWorkloadRL              RateLimiterName = "TrainingWorkload"
	TrainingWorkloadDeprecatedRL    RateLimiterName = "TrainingWorkloadDeprecated"
	InteractiveWorkloadRL           RateLimiterName = "InteractiveWorkload"
	InteractiveWorkloadDeprecatedRL RateLimiterName = "InteractiveWorkloadDeprecated"
	RevisionRL                      RateLimiterName = "Revision"
	InferenceWorkloadRL             RateLimiterName = "InferenceWorkload"
	InferenceWorkloadDeprecatedRL   RateLimiterName = "InferenceWorkloadDeprecated"
	DistributedWorkloadRL           RateLimiterName = "DistributedWorkload"
	DistributedWorkloadDeprecatedRL RateLimiterName = "DistributedWorkloadDeprecated"
	DistributedInferenceWorkloadRL  RateLimiterName = "DistributedInferenceWorkload"
	NodeRL                          RateLimiterName = "Node"
	NumaNodeRL                      RateLimiterName = "NumaNode"
	NodepoolRL                      RateLimiterName = "Nodepool"
	PodRL                           RateLimiterName = "Pod"
	PodDeprecatedRL                 RateLimiterName = "PodDeprecated"
	PodGroupRL                      RateLimiterName = "PodGroup"
	TrainingPolicyRL                RateLimiterName = "TrainingPolicy"
	InteractivePolicyRL             RateLimiterName = "InteractivePolicy"
	InferencePolicyRL               RateLimiterName = "InferencePolicy"
	DistributedPolicyRL             RateLimiterName = "DistributedPolicy"
	ProjectRL                       RateLimiterName = "Project"
	DepartmentRL                    RateLimiterName = "Department"
	PVCRL                           RateLimiterName = "PVC"
	SecretRL                        RateLimiterName = "Secret"
	StorageClassRL                  RateLimiterName = "StorageClass"
	PriorityClassRL                 RateLimiterName = "PriorityClass"
	ApplicationRL                   RateLimiterName = "Application"
	KartaRL                         RateLimiterName = "Karta"
	WorkloadRunnerRL                RateLimiterName = "WorkloadRunner"

	rateLimitersConfigurations = map[RateLimiterName]rateLimiterConfigurationSettings{
		ClusterInfoRL:                   {},
		ConfigMapRL:                     {},
		DatavolumeRL:                    {},
		AccessRuleRL:                    {},
		ExternalWorkloadRL:              {workloadsScale: true},
		TrainingWorkloadRL:              {workloadsScale: true},
		TrainingWorkloadDeprecatedRL:    {workloadsScale: true, deprecated: true},
		InteractiveWorkloadRL:           {workloadsScale: true},
		InteractiveWorkloadDeprecatedRL: {workloadsScale: true, deprecated: true},
		InferenceWorkloadRL:             {workloadsScale: true},
		InferenceWorkloadDeprecatedRL:   {workloadsScale: true, deprecated: true},
		DistributedWorkloadRL:           {workloadsScale: true},
		DistributedWorkloadDeprecatedRL: {workloadsScale: true, deprecated: true},
		DistributedInferenceWorkloadRL:  {workloadsScale: true},
		NodeRL:                          {},
		NumaNodeRL:                      {},
		NodepoolRL:                      {},
		PodRL:                           {workloadsScale: true},
		RevisionRL:                      {workloadsScale: true},
		PodDeprecatedRL:                 {workloadsScale: true, deprecated: true},
		PodGroupRL:                      {workloadsScale: true, deprecated: true},
		TrainingPolicyRL:                {},
		InteractivePolicyRL:             {},
		DistributedPolicyRL:             {},
		InferencePolicyRL:               {},
		ProjectRL:                       {},
		DepartmentRL:                    {},
		PVCRL:                           {},
		SecretRL:                        {},
		StorageClassRL:                  {},
		PriorityClassRL:                 {},
		ApplicationRL:                   {},
		KartaRL:                         {},
		WorkloadRunnerRL:                {workloadsScale: true},
	}
)

func NewRateLimitedController(resourceName RateLimiterName, handler Handler) *RateLimitedController {
	workers, delay := getRateLimiterConf(resourceName)
	minWaitRateLimiter := NewMinWaitRateLimiter[client.ObjectKey](delay)
	queue := workqueue.NewTypedRateLimitingQueueWithConfig(
		workqueue.NewTypedMaxOfRateLimiter(
			workqueue.NewTypedItemExponentialFailureRateLimiter[client.ObjectKey](rateLimiterBaseDelay, delay*2),
			&workqueue.TypedBucketRateLimiter[client.ObjectKey]{Limiter: rate.NewLimiter(rate.Limit(rateLimit), rateBursts)},
			minWaitRateLimiter,
		), workqueue.TypedRateLimitingQueueConfig[client.ObjectKey]{Name: string(resourceName), MetricsProvider: metrics.PrometheusMetricsProviderInstance})

	//TODO: create metrics trigger in CP to get notified when the queue is full

	return &RateLimitedController{
		minWaitRateLimiter: minWaitRateLimiter,
		queue:              queue,
		resourceName:       resourceName,
		handler:            handler,
		workers:            workers,
	}
}

func (c *RateLimitedController) AddRateLimited(key client.ObjectKey) {
	c.queue.AddRateLimited(key)
}

func (c *RateLimitedController) MarkItemDeleted(key client.ObjectKey) {
	c.queue.Forget(key)
}

// Run - runs the rateLimitedController. This should NOT be called as a goroutine,
// since the func will start a goroutine for each worker.
func (c *RateLimitedController) Run() {
	zap.S().Infof("Starting %v controller with %d workers", c.resourceName, c.workers)

	for i := 0; i < c.workers; i++ {
		go c.runWorker()
	}
}

func (c *RateLimitedController) ShutDown() {
	// Let the workers stop when we are done
	zap.S().Infof("Stopping %v controller", c.resourceName)
	c.queue.ShutDown()
}

func (c *RateLimitedController) runWorker() {
	for c.processNextItem() {
	}
}

func (c *RateLimitedController) processNextItem() bool {
	// Wait until there is a new item in the working queue
	item, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers.
	// This allows safe parallel processing because two pods with the same key are never processed in
	// parallel.
	defer c.queue.Done(item)

	shouldContinue := c.minWaitRateLimiter.UpdateLastTimestampToNowIfNotForgotten(item)
	if !shouldContinue {
		return true
	}

	err := c.handler.HandleUpdateItem(item)
	c.handleErr(err, item)
	return true
}

func (c *RateLimitedController) handleErr(err error, key client.ObjectKey) {
	if err != nil {
		if errors.IsNotFound(err) {
			return
		}
		c.queue.AddRateLimited(key)
		zap.S().Errorf("Had an error while trying to update %v %v: %v \n",
			c.resourceName, key, err)
	}
}

func getRateLimiterConf(resourceName RateLimiterName) (int, time.Duration) {
	confSettings, ok := rateLimitersConfigurations[resourceName]
	if !ok {
		zap.S().Fatalf("Resource %v not found in rate limiters configurations; please add it to 'rateLimitersConfigurations' and re-run", resourceName)
	}

	var workers int
	var delay time.Duration

	if confSettings.workloadsScale {
		delay = time.Second * time.Duration(config.Config().RateLimiterWorkloadsDelaySeconds)
		if !confSettings.deprecated {
			workers = config.Config().RateLimiterWorkloadsNumOfWorkers
		} else {
			workers = 1
		}
	} else {
		delay = time.Second * time.Duration(config.Config().RateLimiterRegularDelaySeconds)
		workers = config.Config().RateLimiterRegularNumOfWorkers
	}

	var zeroDuration time.Duration = 0
	if delay == zeroDuration {
		delay = time.Millisecond * 10
	}

	return workers, delay
}
