/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cluster/apis"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"go.uber.org/zap"
)

var instance = config.GetInstance()

func getRunningPhaseForWorkload(workload *apis.Workload) string {
	switch {
	// Both NIM and workloads resolved through workloads-manager use the
	// external cluster API route, whose ready phase is Running.
	case workload.Framework == api.NimService || workload.Framework == api.UndefinedFramework:
		return WorkloadRunningStatus
	case workload.Type == api.Inference || workload.Type == api.DistributedInference:
		return InferenceWorkloadRunningStatus
	default:
		return WorkloadRunningStatus
	}
}

func WaitForWorkloadInClusterRunning(ctx context.Context, workload *apis.Workload, clusterUrlStr, token string, waitTime time.Duration) error {
	expectedPhase := getRunningPhaseForWorkload(workload)
	return WaitForWorkloadInClusterStatus(ctx, workload, clusterUrlStr, token, waitTime, expectedPhase)
}

func WaitForWorkloadInCluster(ctx context.Context, workload *apis.Workload, clusterUrlStr, token string, waitTime time.Duration) error {
	return WaitForWorkloadInClusterStatus(ctx, workload, clusterUrlStr, token, waitTime, "")
}

func WaitForWorkloadInClusterStatus(ctx context.Context, workload *apis.Workload, clusterUrlStr, token string, waitTime time.Duration, expectedStatus string) error {
	var (
		err        error
		firstCheck = true
	)
	logger := instance.GetLogger()
	end := time.Now().Add(waitTime)
	logger.Debug("waiting for cluster to get workload", zap.Time("now", time.Now()), zap.Time("end", end))
	timeInc := time.Second

	for !time.Now().After(end) || firstCheck {
		firstCheck = false
		time.Sleep(timeInc)
		if expectedStatus == "" {
			_, err = GetWorkloadStatus(ctx, workload, clusterUrlStr, token)
		} else {
			err = ValidateWorkloadStatus(ctx, workload, clusterUrlStr, token, expectedStatus)
		}
		if err == nil {
			return nil
		}
		logger.Debug("validation workload failed", zap.Error(err))
		timeInc = timeInc * 2
		if time.Now().Add(timeInc).After(end) {
			timeInc = time.Until(end)
			timeInc += time.Second
		}
		if timeInc > 0 {
			logger.Debug("increment time sleep", zap.Time("now", time.Now()), zap.Duration("duration", timeInc))
		}
	}
	return fmt.Errorf("workload's waiting time was exceeded")
}
