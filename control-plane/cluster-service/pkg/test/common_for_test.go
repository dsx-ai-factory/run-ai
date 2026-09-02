/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"encoding/json"
	"time"

	pointer "k8s.io/utils/ptr"

	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
)

const (
	testClusterUuid0            = "7daba6ee-2aa9-4ab3-b735-a3d4bcdff72f"
	testClusterUuid1            = "5fcba6ee-2aa9-4ab3-8ab3-a3d4bcdff72f"
	testClusterUuid2            = "2b9ba6ee-2aa9-4ab3-7a12-a3d4bcdff72f"
	notPermittedTestClusterUuid = "1daba6ee-2aa9-4ab3-b735-a3d4bcdff72f"

	testNodeName0 = "node-worker0"
	testNodeName1 = "node-worker1"
	testNodeName2 = "node-worker2"
	testNodeName3 = "node-worker3"
	testNodeName4 = "node-worker4"
	testNodeName5 = "node-worker5"

	testNodePoolName0 = "nodepool-a"
	testNodePoolName1 = "default"

	testGpuType1 = "Tesla-V100"
	testGpuType2 = "Navi"

	testGpuMemoryBytes1 = int64(34359738368)
	testGpuMemoryBytes2 = int64(17179869184)
)

var (
	testClusterUuids = []string{testClusterUuid0, testClusterUuid1, testClusterUuid2}
	testNodeNames    = []string{
		testNodeName0, testNodeName1, testNodeName2,
		testNodeName3, testNodeName4, testNodeName5,
	}

	testTopologyLabels0 = &[]api.NetworkTopologyLabel{
		{LabelKey: "zone", LabelValue: "zone-a"},
		{LabelKey: "region", LabelValue: "region-1"},
		{LabelKey: "hostname", LabelValue: "hostname-0"},
	}
	testTopologyLabels1 = &[]api.NetworkTopologyLabel{
		{LabelKey: "zone", LabelValue: "zone-b"},
		{LabelKey: "region", LabelValue: "region-1"},
		{LabelKey: "hostname", LabelValue: "hostname-1"},
	}
	testTopologyLabels2 = &[]api.NetworkTopologyLabel{
		{LabelKey: "zone", LabelValue: "zone-a"},
		{LabelKey: "region", LabelValue: "region-2"},
		{LabelKey: "hostname", LabelValue: "hostname-2"},
	}

	TestNodeInfo0 = &api.NodeInfo{
		Status: "NotReady",
		Conditions: []api.NodeStatusConditionDetails{
			{
				Type:   "DiskPressure",
				Reason: "NoSpaceAvailable",
			},
			{
				Type:    "UnschedulableCordoned",
				Reason:  "NodeMarkedUnschedulable",
				Message: "node marked unschedulable",
			},
		},
		Taints: []api.NodeTaint{
			{
				Key:    "some-key",
				Value:  "some-value",
				Effect: "NoSchedule",
			},
			{
				Key:    "some-key2",
				Effect: "NoExecute",
			},
		},
		NodePool:  testNodePoolName0,
		CreatedAt: time.Now(),
		GpuInfo: &api.GpuInfo{
			GpuType:        testGpuType1,
			GpuCount:       1,
			GpuMemoryBytes: testGpuMemoryBytes1,
		},
		NvLinkDomainUid:       pointer.To("DOMAIN-X"),
		NvLinkCliqueId:        pointer.To("CLIQUE-Y"),
		NetworkTopologyLabels: testTopologyLabels0,
		Numa: &api.NumaInfo{
			TopologyManagerPolicy: pointer.To(api.TOPOLOGYMANAGERPOLICY_SINGLE_NUMA_NODE),
		},
	}

	TestNodeInfo1 = &api.NodeInfo{
		Status: "Ready",
		Taints: []api.NodeTaint{
			{
				Key:    "some-key2",
				Effect: "NoExecute",
			},
		},
		NodePool:  testNodePoolName1,
		CreatedAt: time.Now(),
		GpuInfo: &api.GpuInfo{
			GpuType:        testGpuType2,
			GpuCount:       2,
			GpuMemoryBytes: testGpuMemoryBytes2,
		},
		NetworkTopologyLabels: testTopologyLabels1,
	}

	TestNodeInfo2 = &api.NodeInfo{
		Status:                "Unknown",
		NodePool:              testNodePoolName1,
		CreatedAt:             time.Now(),
		NetworkTopologyLabels: testTopologyLabels2,
	}

	TestNodeInfo3 = *TestNodeInfo2
	TestNodeInfo4 = *TestNodeInfo0
	TestNodeInfo5 = *TestNodeInfo1

	testNodeInfos = []*api.NodeInfo{
		TestNodeInfo0, TestNodeInfo1, TestNodeInfo2,
		&TestNodeInfo3, &TestNodeInfo4, &TestNodeInfo5,
	}
)

func httpResponseBodyToNodes(httpResponse api.ImplResponse) *[]api.Node {
	bytes, err := json.Marshal(httpResponse.Body)
	Expect(err).To(BeNil())

	nodes := &api.Nodes{}
	err = json.Unmarshal(bytes, nodes)
	Expect(err).To(BeNil())
	return &nodes.Nodes
}

func httpResponseBodyToNode(httpResponse api.ImplResponse) *api.Node {
	bytes, err := json.Marshal(httpResponse.Body)
	Expect(err).To(BeNil())

	node := &api.Node{}
	err = json.Unmarshal(bytes, node)
	Expect(err).To(BeNil())
	return node
}
