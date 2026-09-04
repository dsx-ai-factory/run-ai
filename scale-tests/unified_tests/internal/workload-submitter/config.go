/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload_submitter

import (
	"fmt"
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type SubmissionMode string

const (
	SubmissionModeRate  SubmissionMode = "rate"
	SubmissionModeBurst SubmissionMode = "burst"
)

type Config struct {
	// Cluster connectivity
	ControlPlaneURL      string `envconfig:"CONTROL_PLANE_URL" required:"true"`
	ControlPlaneUsername string `envconfig:"CONTROL_PLANE_USERNAME" required:"true"`
	ControlPlanePassword string `envconfig:"CONTROL_PLANE_PASSWORD" required:"true"`
	ClusterID            string `envconfig:"CLUSTER_ID"` // auto-resolved if empty

	// Cluster topology (used to derive W_cap)
	NodepoolCount int `envconfig:"NODEPOOL_COUNT" default:"1"`
	NodesPerPool  int `envconfig:"NODES_PER_POOL" default:"313"`
	GPUsPerNode   int `envconfig:"GPUS_PER_NODE" default:"8"`

	// Submitter behaviour
	SubmissionMode  SubmissionMode `envconfig:"SUBMISSION_MODE" default:"rate"`
	PodsPerWorkload int            `envconfig:"PODS_PER_WORKLOAD" default:"1"`
	ProjectPrefix   string         `envconfig:"PROJECT_PREFIX" default:"load-project"`
	ProjectCount    int            `envconfig:"PROJECT_COUNT" default:"1"`

	// NodepoolPrefix selects the seeded nodepools, excluding the cluster's "default" one.
	NodepoolPrefix string `envconfig:"NODEPOOL_PREFIX" default:"nodepool-"`

	// Seconds after completion before a workload is deleted; 0 = never.
	AutoDeletionSeconds int `envconfig:"AUTO_DELETION_SECONDS" default:"300"`

	// Where retention.json is written for the orchestrator to stamp onto results.
	OutputDir string `envconfig:"OUTPUT_DIR" default:"output"`

	// Rate mode
	// WorkloadDuration must be empty in rate mode: D is derived as W_cap/R (the cycle duration
	// that keeps queue depth self-maintaining) and cannot be set independently of R.
	WPending             int     `envconfig:"W_PENDING" default:"0"`
	SubmissionRatePerMin float64 `envconfig:"SUBMISSION_RATE" default:"100"`
	SteadyStateDuration  string  `envconfig:"STEADY_STATE_DURATION" default:"120m"`
	MaxConcurrent        int     `envconfig:"MAX_CONCURRENT" default:"10"`

	// Burst mode
	WorkloadDuration string `envconfig:"WORKLOAD_DURATION" default:""`
	// BurstTarget is the number of workloads to submit. Defaults to 50×W_cap when 0.
	BurstTarget      int    `envconfig:"BURST_TARGET" default:"0"`
	BurstWorkers     int    `envconfig:"BURST_WORKERS" default:"100"`
	// ObservationWindow is how long to wait after all workloads are submitted before
	// exiting. Defaults to (ceil(BurstTarget/W_cap)+2)×WorkloadDuration when empty.
	ObservationWindow string `envconfig:"OBSERVATION_WINDOW" default:""`
}

func ConstructConfig() *Config {
	var c Config
	if err := envconfig.Process("", &c); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	if err := c.validate(); err != nil {
		log.Fatalf("invalid config: %v", err)
	}
	return &c
}

func (c *Config) validate() error {
	if c.PodsPerWorkload <= 0 {
		return fmt.Errorf("PODS_PER_WORKLOAD must be > 0, got %d", c.PodsPerWorkload)
	}
	if c.GPUsPerNode <= 0 {
		return fmt.Errorf("GPUS_PER_NODE must be > 0, got %d", c.GPUsPerNode)
	}
	if c.NodepoolCount < 1 {
		return fmt.Errorf("NODEPOOL_COUNT must be >= 1, got %d", c.NodepoolCount)
	}
	if c.ProjectCount < 1 {
		return fmt.Errorf("PROJECT_COUNT must be >= 1, got %d", c.ProjectCount)
	}

	switch c.SubmissionMode {
	case SubmissionModeRate:
		if c.WorkloadDuration != "" {
			return fmt.Errorf("WORKLOAD_DURATION must be empty in rate mode: D is derived as W_cap/R and cannot be set independently")
		}
		if c.SubmissionRatePerMin <= 0 {
			return fmt.Errorf("SUBMISSION_RATE must be > 0, got %f", c.SubmissionRatePerMin)
		}
		if c.MaxConcurrent <= 0 {
			return fmt.Errorf("MAX_CONCURRENT must be > 0, got %d", c.MaxConcurrent)
		}
		if _, err := time.ParseDuration(c.SteadyStateDuration); err != nil {
			return fmt.Errorf("invalid STEADY_STATE_DURATION %q: %w", c.SteadyStateDuration, err)
		}
	case SubmissionModeBurst:
		if c.WorkloadDuration == "" {
			return fmt.Errorf("WORKLOAD_DURATION is required in burst mode")
		}
		if _, err := time.ParseDuration(c.WorkloadDuration); err != nil {
			return fmt.Errorf("invalid WORKLOAD_DURATION %q: %w", c.WorkloadDuration, err)
		}
		if c.BurstTarget < 0 {
			return fmt.Errorf("BURST_TARGET must be >= 0, got %d", c.BurstTarget)
		}
		if c.BurstWorkers <= 0 {
			return fmt.Errorf("BURST_WORKERS must be > 0, got %d", c.BurstWorkers)
		}
		if c.ObservationWindow != "" {
			if _, err := time.ParseDuration(c.ObservationWindow); err != nil {
				return fmt.Errorf("invalid OBSERVATION_WINDOW %q: %w", c.ObservationWindow, err)
			}
		}
	default:
		return fmt.Errorf("SUBMISSION_MODE must be %q or %q, got %q", SubmissionModeRate, SubmissionModeBurst, c.SubmissionMode)
	}
	return nil
}
