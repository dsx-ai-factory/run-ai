/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"context"
	"log"
	"math"
	"time"

	ws "github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/workload-submitter"
)

func main() {
	cfg := ws.ConstructConfig()

	ctx := context.Background()

	clients, err := ws.InitClients(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to init clients: %v", err)
	}

	// Resolve cluster ID
	if cfg.ClusterID == "" {
		cfg.ClusterID, err = ws.GetFirstClusterID(ctx, clients)
		if err != nil {
			log.Fatalf("failed to resolve cluster ID: %v", err)
		}
	}
	log.Printf("cluster ID: %s", cfg.ClusterID)

	// Resolve actual nodepool names from the cluster
	nodepoolNames, err := ws.ResolveNodepoolNames(ctx, clients, cfg.NodepoolPrefix)
	if err != nil {
		log.Fatalf("failed to resolve nodepool names: %v", err)
	}
	// Exact, and no truncation. Taking the first N would submit into a topology that
	// silently differs from the one the run is supposed to measure.
	if len(nodepoolNames) != cfg.NodepoolCount {
		log.Fatalf("expected exactly %d nodepools with prefix %q, found %d: %v",
			cfg.NodepoolCount, cfg.NodepoolPrefix, len(nodepoolNames), nodepoolNames)
	}
	log.Printf("nodepools (%d, prefix %q): %v", len(nodepoolNames), cfg.NodepoolPrefix, nodepoolNames)

	// Resolve project IDs by prefix
	projectIDs, err := ws.ResolveProjectIDs(ctx, clients, cfg.ClusterID, cfg.ProjectPrefix)
	if err != nil {
		log.Fatalf("failed to resolve project IDs: %v", err)
	}
	if len(projectIDs) != cfg.ProjectCount {
		log.Fatalf("expected exactly %d projects with prefix %q, found %d",
			cfg.ProjectCount, cfg.ProjectPrefix, len(projectIDs))
	}
	log.Printf("projects: %d matching prefix %q", len(projectIDs), cfg.ProjectPrefix)

	// W_cap is computed per nodepool first (integer division) to ensure that round-robin
	// assignment never overflows a single nodepool's GPU capacity.
	// Example: 3 nodepools × 313 nodes × 8 GPUs / 200 pods per workload
	//   wrong:  floor(7512/200) = 37  →  nodepool-1 gets 13 workloads × 200 = 2600 > 2504
	//   correct: 3 × floor(2504/200) = 3 × 12 = 36  →  each pool gets exactly 12
	wCapPerPool := (cfg.NodesPerPool * cfg.GPUsPerNode) / cfg.PodsPerWorkload
	wCap := cfg.NodepoolCount * wCapPerPool
	if wCap == 0 {
		log.Fatalf("W_cap is 0: NodesPerPool×GPUsPerNode (%d) < PodsPerWorkload (%d)",
			cfg.NodesPerPool*cfg.GPUsPerNode, cfg.PodsPerWorkload)
	}
	log.Printf("  W_cap_per_pool = %d  (NodesPerPool=%d × GPUsPerNode=%d / PodsPerWorkload=%d)",
		wCapPerPool, cfg.NodesPerPool, cfg.GPUsPerNode, cfg.PodsPerWorkload)
	log.Printf("  W_cap          = %d  (%d nodepools × %d)", wCap, cfg.NodepoolCount, wCapPerPool)

	var params ws.Params
	params.WCap = wCap

	switch cfg.SubmissionMode {
	case ws.SubmissionModeBurst:
		dBurst, _ := time.ParseDuration(cfg.WorkloadDuration) // validated in ConstructConfig

		burstTarget := cfg.BurstTarget
		if burstTarget == 0 {
			burstTarget = 50 * wCap
		}

		var observationWindow time.Duration
		if cfg.ObservationWindow != "" {
			observationWindow, _ = time.ParseDuration(cfg.ObservationWindow) // validated
		} else {
			waves := int(math.Ceil(float64(burstTarget) / float64(wCap)))
			observationWindow = time.Duration(waves+2) * dBurst
		}

		log.Printf("  mode           = burst")
		log.Printf("  burst_target   = %d  (%d×W_cap)", burstTarget, burstTarget/wCap)
		log.Printf("  D_burst        = %v", dBurst)
		log.Printf("  burst_workers  = %d", cfg.BurstWorkers)
		log.Printf("  obs_window     = %v", observationWindow)

		params.DSecs = int(dBurst.Seconds())
		params.BurstTarget = burstTarget
		params.ObservationWindow = observationWindow

		log.Printf("applying KWOK stage (duration %ds)...", params.DSecs)
		if err := ws.ApplyKwokStage(ctx, clients, params.DSecs); err != nil {
			log.Fatalf("failed to apply KWOK stage: %v", err)
		}

		s := ws.NewSubmitter(cfg, clients, params, projectIDs, nodepoolNames)
		s.RunBurst(ctx)
		s.PrintSummary()
		if err := s.WriteRetention(cfg.OutputDir); err != nil {
			log.Printf("WARNING: could not write retention.json: %v", err)
		}

	case ws.SubmissionModeRate:
		wTotal := wCap + cfg.WPending
		rampRate := cfg.SubmissionRatePerMin * float64(wTotal) / float64(wCap)
		dSecs := int(float64(wCap) / cfg.SubmissionRatePerMin * 60.0)
		steadyDur, _ := time.ParseDuration(cfg.SteadyStateDuration) // validated in ConstructConfig

		log.Printf("  mode           = rate")
		log.Printf("  W_pending      = %d", cfg.WPending)
		log.Printf("  W_total        = %d", wTotal)
		log.Printf("  R (steady)     = %.2f/min", cfg.SubmissionRatePerMin)
		log.Printf("  ramp_rate      = %.2f/min  (R × W_total/W_cap)", rampRate)
		log.Printf("  D              = %ds  (W_cap/R = cycle duration)", dSecs)
		log.Printf("  steady_dur     = %v", steadyDur)

		params.DSecs = dSecs
		params.WTotal = wTotal
		params.RampRate = rampRate
		params.SteadyDuration = steadyDur

		log.Printf("applying KWOK stage (duration %ds)...", dSecs)
		if err := ws.ApplyKwokStage(ctx, clients, dSecs); err != nil {
			log.Fatalf("failed to apply KWOK stage: %v", err)
		}

		s := ws.NewSubmitter(cfg, clients, params, projectIDs, nodepoolNames)
		s.Run(ctx)
		s.PrintSummary()
		if err := s.WriteRetention(cfg.OutputDir); err != nil {
			log.Printf("WARNING: could not write retention.json: %v", err)
		}
	}
}
