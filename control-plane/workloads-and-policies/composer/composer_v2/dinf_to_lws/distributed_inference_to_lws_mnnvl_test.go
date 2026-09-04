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

// mnnvlGoldenBaseInput is a minimal valid distributed-inference spec reused
// by every MNNVL entry so the goldens differ only in the resolved MNNVL
// plugin mode.
const mnnvlGoldenBaseInput = "dinfw_serving_port_basic"

// mnnvlTestClusterConfig returns a cluster storage config that carries only
// the `dinfGPUNetworkAcceleration` flag. MNNVL entries use this instead of
// the default legacy typed cluster config so their goldens capture solely
// the MNNVL plugin resolution and not the ambient OAuth/proxy config.
func mnnvlTestClusterConfig(dinfEnabled bool) composer.ClusterStorageConfig {
	cluster := clusterApi.NewDisplayedCluster("cluster-a", 1, "cluster-a", time.Time{})
	cluster.SetDomain("https://test.run.ai")
	reported := clusterApi.ClusterReportedConfig{}
	reported.SetDinfGPUNetworkAcceleration(dinfEnabled)
	status := clusterApi.ClusterDisplayedStatus{}
	status.SetConfig(reported)
	cluster.Status.Set(&status)
	return composer.NewClusterStorageConfigFromCluster(cluster)
}

// mnnvlTests exercises MNNVL plugin resolution: the mode is taken from the
// leader's run.ai/MNNVL label when present, otherwise defaulted from the
// cluster's dinfGPUNetworkAcceleration flag. Every entry uses
// mnnvlGoldenBaseInput so the goldens differ only in the resolved mode.
// These entries flow through the same `ComposerTests` DescribeTable as the
// main golden cases.
var mnnvlTests = []TableEntry{
	// Cluster-driven defaults (dinfGPUNetworkAcceleration=true on the
	// cluster): absent label -> Required; explicit None overrides the
	// cluster default.
	Entry("mnnvl cluster acceleration on defaults to Required", ComposerTest{
		input:         mnnvlGoldenBaseInput,
		expected:      "mnnvl_cluster_on_default_required",
		clusterConfig: mnnvlTestClusterConfig(true),
	}),
	Entry("mnnvl explicit None label overrides cluster on", ComposerTest{
		input:         mnnvlGoldenBaseInput,
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
