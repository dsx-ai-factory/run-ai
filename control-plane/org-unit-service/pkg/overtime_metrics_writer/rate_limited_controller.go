/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package overtime_metrics_writer

import (
	"context"
	"time"

	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/dao"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"golang.org/x/time/rate"
	"k8s.io/client-go/util/workqueue"
)

const (
	rateLimiterBaseDelay = time.Millisecond * 500
	rateLimiterMaxDelay  = time.Minute * 120
)

type RateLimitingController struct {
	queue workqueue.TypedRateLimitingInterface[RateLimitedItem]
}

type RateLimitedItem struct {
	overtimeMetricRange dao.OvertimeMetricRange
	tenantId            string
}

// NewRateLimitingController - creates a new RateLimitingController.
// limit and burst are the params for rate.Limiter.
func NewRateLimitingController(limit, burst int) *RateLimitingController {
	rateLimiter := workqueue.NewTypedMaxOfRateLimiter[RateLimitedItem](
		// per-item rate limiting
		workqueue.NewTypedItemExponentialFailureRateLimiter[RateLimitedItem](rateLimiterBaseDelay, rateLimiterMaxDelay),
		// overall rate limiting (not per item)
		&workqueue.TypedBucketRateLimiter[RateLimitedItem]{Limiter: rate.NewLimiter(rate.Limit(limit), burst)},
	)

	queue := workqueue.NewTypedRateLimitingQueue[RateLimitedItem](rateLimiter)

	logging.Infof("Rate limiting controller created with limit %d and burst %d", limit, burst)

	return &RateLimitingController{
		queue: queue,
	}
}

func (c *RateLimitingController) AddRateLimited(item RateLimitedItem) {
	c.queue.AddRateLimited(item)
}

// Run - runs the RateLimitingController. This should be called as a goroutine.
func (c *RateLimitingController) Run(ctx context.Context) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Setting up a go routine of the rate-limiting controller for overtime metrics writer")

	for c.processNextItem(ctx) {
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Shutting down rate limiting controller")
}

func (c *RateLimitingController) ShutDown(ctx context.Context) {
	// Let the workers stop when we are done
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Stopping rate limiting controller")
	c.queue.ShutDown()
}

func (c *RateLimitingController) processNextItem(ctx context.Context) bool {
	// Wait until there is a new item in the working queue
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Waiting for the next item in the rate limiting controller queue")
	item, shutdown := c.queue.Get()
	if shutdown {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Shutting down rate limiting controller - queue.Get returned shutdown true")
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers.
	// This allows safe parallel processing because two pods with the same key are never processed in
	// parallel.
	defer c.queue.Done(item)

	ctx = context.WithValue(ctx, runaictx.CfTenantId, item.tenantId)
	err := QueryOvertimeMetrics(ctx, item.tenantId, item.overtimeMetricRange)
	if err != nil {
		// Put the item back to the queue to handle it later
		// the ItemExponentialFailureRateLimiter will take care of the rate limiting
		c.queue.AddRateLimited(item)
	} else {
		// Forget the item so the ItemExponentialFailureRateLimiter will not take care of it anymore
		c.queue.Forget(item)
	}
	return true
}
