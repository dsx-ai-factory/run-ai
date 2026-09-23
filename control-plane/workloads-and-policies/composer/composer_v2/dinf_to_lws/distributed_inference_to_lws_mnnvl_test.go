/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dinf_to_lws

import (
	"time"

	clusterApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	. "github.com/onsi/ginkgo/v2"
)

// mnnvlGoldenBaseInput is a minimal valid distributed-inference spec that
// requests no GPU. mnnvlGoldenGpuInput is the same spec with a leader GPU
// request. MNNVL entries pick the fixture matching the scenario so the
// GPU request lives in the spec rather than being injected by the harness.
const (
	mnnvlGoldenBaseInput = "dinfw_serving_port_basic"
	mnnvlGoldenGpuInput  = "dinfw_serving_port_gpu"
)

// mnnvlTestClusterConfig returns a cluster storage config that carries only
// the `dinfGPUNetworkAcceleration` flag. MNNVL entries use this instead of
// the default legacy typed cluster config so their goldens capture solely
// the MNNVL plugin resolution and not the ambient OAuth/proxy config.
func mnnvlTestClusterConfig(dinfEnabled bool) composer.ClusterStorageConfig {
	cluster := clusterApi.NewDisplayedCluster("cluster-a", 1, "cluster-a", time.Time{})
	cluster.SetDomain("https://mock-q7z3k.example.com")
	reported := clusterApi.ClusterReportedConfig{}
	reported.SetDinfGPUNetworkAcceleration(dinfEnabled)
	status := clusterApi.ClusterDisplayedStatus{}
	status.SetConfig(reported)
	cluster.Status.Set(&status)
	return composer.NewClusterStorageConfigFromCluster(cluster)
}

// mnnvlTests exercises MNNVL plugin resolution: the mode is taken from the
// leader's run.ai/MNNVL label when present, otherwise defaulted from the
// cluster's dinfGPUNetworkAcceleration flag -- but only for workloads that
// request GPU. Every entry uses mnnvlGoldenBaseInput so the goldens differ
// only in the resolved mode (and the injected leader GPU request, when set).
// These entries flow through the same `ComposerTests` DescribeTable as the
// main golden cases.
var mnnvlTests = []TableEntry{
	// Cluster-driven defaults (dinfGPUNetworkAcceleration=true on the
	// cluster): a GPU-requesting workload with no label -> Required; an
	// explicit None label overrides the cluster default.
	Entry("mnnvl cluster acceleration on with gpu defaults to Required", ComposerTest{
		input:         mnnvlGoldenGpuInput,
		expected:      "mnnvl_cluster_on_gpu_default_required",
		clusterConfig: mnnvlTestClusterConfig(true),
	}),
	// A workload that requests no GPU never inherits the cluster default,
	// even when dinfGPUNetworkAcceleration is on -> None.
	Entry("mnnvl cluster acceleration on without gpu defaults to None", ComposerTest{
		input:         mnnvlGoldenBaseInput,
		expected:      "mnnvl_cluster_on_no_gpu_none",
		clusterConfig: mnnvlTestClusterConfig(true),
	}),
	Entry("mnnvl explicit None label overrides cluster on", ComposerTest{
		input:         mnnvlGoldenGpuInput,
		expected:      "mnnvl_label_none_overrides_cluster_on",
		clusterConfig: mnnvlTestClusterConfig(true),
		mnnvlLabel:    "None",
	}),

	// Label-driven resolution when the cluster's MNNVL flag is off. The
	// WorkloadRunner plugin normalizes Preferred -> Required, so labels
	// map: absent -> None, Preferred -> Required, explicit None -> None.
	Entry("mnnvl no label defaults to None (cluster off)", ComposerTest{
		input:         mnnvlGoldenBaseInput,
		expected:      "mnnvl_no_label_defaults_none",
		clusterConfig: mnnvlTestClusterConfig(false),
	}),
	Entry("mnnvl Preferred label normalizes to Required (cluster off)", ComposerTest{
		input:         mnnvlGoldenBaseInput,
		expected:      "mnnvl_preferred_normalizes_required",
		clusterConfig: mnnvlTestClusterConfig(false),
		mnnvlLabel:    "Preferred",
	}),
	Entry("mnnvl explicit None label honored (cluster off)", ComposerTest{
		input:         mnnvlGoldenBaseInput,
		expected:      "mnnvl_explicit_none_honored",
		clusterConfig: mnnvlTestClusterConfig(false),
		mnnvlLabel:    "None",
	}),
}
