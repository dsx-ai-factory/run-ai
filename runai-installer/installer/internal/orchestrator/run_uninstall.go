/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orchestrator

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	"github.com/run-ai/runai/runai-installer/installer/internal/flow"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	"github.com/run-ai/runai/runai-installer/installer/internal/uninstall"
	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
)

func RunUninstall(cfg *config.UninstallConfig, deps uninstall.Deps) error {
	if err := logging.InitFromDir(cfg.LogsDir); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to initialize log file in %s: %v\n", cfg.LogsDir, err)
	}
	logging.Infof("Starting CLI uninstall...")
	step.SetDefaultStepTimeout(cfg.StepTimeout)

	externalEvents := make(chan *flow.FlowStatusReportingMsg, 100)
	flowGroups := buildUninstallFlowGroups(cfg, deps, externalEvents)

	executionTimeout := cfg.ExecutionTimeout
	if executionTimeout <= 0 {
		executionTimeout = 50 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), executionTimeout)
	defer cancel()

	go func() {
		defer close(externalEvents)
		for _, fg := range flowGroups {
			internalEvents := make(chan interface{}, 100)
			fg.Manager.ExecuteFlow(ctx, nil, internalEvents)
			fg.Manager.Wait()
			close(internalEvents)
		}
	}()

	var firstFailure string
	for event := range externalEvents {
		switch event.FlowStatus {
		case model.Installing:
			logging.Infof("Uninstalling %s...", event.FlowName)
		case model.Installed:
			logging.Infof("Uninstalled %s", event.FlowName)
		case model.Failed, model.Timeout:
			logging.Errorf("Failed to uninstall %s", event.FlowName)
			if firstFailure == "" {
				firstFailure = event.FlowName
			}
		}
	}

	if firstFailure != "" {
		return fmt.Errorf("uninstall failed: flow %q did not succeed", firstFailure)
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("uninstall aborted: %w", err)
	}
	return nil
}
