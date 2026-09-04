/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"context"
	"log"

	"github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/resource-rec/projects"
	"github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/resource-rec/topology"
	"github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/sampler_actions"
	"github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/scale"
	"github.com/kelseyhightower/envconfig"
)

type resourceRecConfig struct {
	ProjectCount  int    `envconfig:"PROJECT_COUNT" default:"100"`
	ProjectPrefix string `envconfig:"PROJECT_PREFIX" default:"load-project"`
}

func main() {
	ctx := context.Background()

	clusterConfig := scale.ConstructConfig()

	var rrCfg resourceRecConfig
	if err := envconfig.Process("", &rrCfg); err != nil {
		log.Fatalf("failed to load resource-rec config: %v", err)
	}

	kubeConfig := sampler_actions.InitKubeConfig(clusterConfig)
	clients, err := sampler_actions.InitClients(ctx, clusterConfig, kubeConfig)
	if err != nil {
		log.Fatalf("failed to initialize clients: %v", err)
	}

	// Resolve cluster ID once up front; both topology and project provisioners need it.
	if err := sampler_actions.SetClusterId(ctx, clients, clusterConfig); err != nil {
		log.Fatalf("failed to resolve cluster ID: %v", err)
	}
	log.Printf("cluster ID: %s", clusterConfig.ClusterID)

	topoProvisioner := topology.NewProvisioner(clusterConfig, clients)
	if err := topoProvisioner.Run(ctx); err != nil {
		log.Fatalf("topology provisioning failed: %v", err)
	}

	projectsProvisioner := projects.NewProvisioner(clusterConfig, rrCfg.ProjectCount, rrCfg.ProjectPrefix, clients)
	if err := projectsProvisioner.Run(ctx); err != nil {
		log.Fatalf("project provisioning failed: %v", err)
	}
}
