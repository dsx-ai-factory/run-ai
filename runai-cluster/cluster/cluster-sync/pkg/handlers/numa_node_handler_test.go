/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	kaiv1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1"
	kainodepool "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

var _ = Describe("IsNumaEnabled", func() {
	DescribeTable("reports whether the NUMA plugin is enabled",
		func(nodePool *kainodepool.NodePool, expected bool) {
			Expect(IsNumaEnabled(nodePool)).To(Equal(expected))
		},
		Entry("no scheduling shard config", &kainodepool.NodePool{}, false),
		Entry("no plugins", &kainodepool.NodePool{
			Spec: kainodepool.NodePoolSpec{
				SchedulingShardConfig: &kainodepool.SchedulingShardConfig{},
			},
		}, false),
		// The agent only writes the plugin when NUMA is on, so an absent key means off.
		Entry("another plugin only", nodePoolWithPlugins(map[string]kaiv1.PluginConfig{
			"gpusharing": {Enabled: ptr.To(true)},
		}), false),
		Entry("numa plugin without an enabled flag", nodePoolWithPlugins(map[string]kaiv1.PluginConfig{
			numaPluginName: {},
		}), false),
		Entry("numa plugin disabled", nodePoolWithPlugins(map[string]kaiv1.PluginConfig{
			numaPluginName: {Enabled: ptr.To(false)},
		}), false),
		Entry("numa plugin enabled", nodePoolWithPlugins(map[string]kaiv1.PluginConfig{
			numaPluginName: {Enabled: ptr.To(true)},
		}), true),
	)
})

func nodePoolWithPlugins(plugins map[string]kaiv1.PluginConfig) *kainodepool.NodePool {
	return &kainodepool.NodePool{
		Spec: kainodepool.NodePoolSpec{
			SchedulingShardConfig: &kainodepool.SchedulingShardConfig{
				Plugins: plugins,
			},
		},
	}
}
