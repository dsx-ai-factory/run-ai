/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	cluster_seeder "github.com/run-ai/runai/scale-tests/unified_tests/internal/cluster-seeder"
	"github.com/run-ai/runai/scale-tests/unified_tests/internal/scale"
	seed_cluster "github.com/run-ai/runai/scale-tests/unified_tests/internal/workload-scale/seed-cluster"
)

func main() {
	clusterConfig := scale.ConstructConfig()
	workloadScaleSampleRunner := seed_cluster.NewWorkloadSamplesRunner(clusterConfig)
	cluster_seeder.DeleteCluster(workloadScaleSampleRunner, clusterConfig)
}
