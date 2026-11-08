// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package dao

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	authorization "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGPUIdentitySchemaAndLegacyPayload(t *testing.T) {
	for _, tc := range []struct {
		id    string
		valid bool
	}{
		{"", true}, {"1eb8:10de", true}, {"0001:10de", true}, {"0010:10de", true}, {"0100:10de", true}, {"1000:10de", true}, {"0000:10de", false}, {"1EB8:10DE", false}, {"2330:1002", false}, {"1eb8", false}, {" 1eb8:10de", false},
	} {
		t.Run(tc.id, func(t *testing.T) {
			if err := api.AssertGpuInfoConstraints(api.GpuInfo{PciDeviceId: tc.id}); (err == nil) != tc.valid {
				t.Fatalf("GPU constraint valid=%v: %v", tc.valid, err)
			}
			if err := api.AssertNodepoolGpuTypeDetailsConstraints(api.NodepoolGpuTypeDetails{PciDeviceId: tc.id}); (err == nil) != tc.valid {
				t.Fatalf("pool constraint valid=%v: %v", tc.valid, err)
			}
		})
	}
	var info api.NodeInfo
	if err := json.Unmarshal([]byte(`{"status":"Ready","gpuInfo":{"gpuType":"Tesla-T4","gpuCount":1}}`), &info); err != nil {
		t.Fatal(err)
	}
	node := NodeInfoToNodeDao(1, "cluster", "node", info)
	b, err := json.Marshal(node.ToApiModel())
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatal(err)
	}
	if _, exists := decoded["gpuInfo"].(map[string]interface{})["pciDeviceId"]; exists {
		t.Fatalf("legacy node gained identity: %s", b)
	}
	info.GpuInfo.PciDeviceId = "1eb8:10de"
	observedAt := time.Now().Add(-time.Minute).Truncate(time.Microsecond)
	info.GpuInfo.PciDeviceIdObservedAt = &observedAt
	got := NodeInfoToNodeDao(1, "cluster", "node", info).ToApiModel()
	if got.GpuInfo.PciDeviceId != "1eb8:10de" {
		t.Fatalf("PCI storage conversion lost identity: %+v", got.GpuInfo)
	}
}

var _ = Describe("Node PCI identity persistence", func() {
	It("round trips, separates aggregation and clears identity through bulk and single upserts", func() {
		ctx := context.Background()
		tenant := int32(49631)
		cluster, err := CreateCluster(ctx, &models.Cluster{Name: "pci-identity", Domain: "pci.test", Version: "3.0.0", TenantId: tenant})
		Expect(err).To(BeNil())
		clusterID := cluster.ID.String()
		DeferCleanup(func() { Expect(DeleteClusterNodes(ctx, tenant, clusterID)).To(Succeed()) })
		node := func(name, id string, count int32) models.Node {
			return *NodeInfoToNodeDao(tenant, clusterID, name, api.NodeInfo{Status: "Ready", CreatedAt: time.Now(), NodePool: "pci-pool", GpuInfo: &api.GpuInfo{GpuType: "Same-Name", GpuCount: count, GpuMemoryBytes: 16106127360, PciDeviceId: id, PciDeviceIdObservedAt: func() *time.Time {
				if id == "" {
					return nil
				}
				observedAt := time.Now().Add(-time.Minute).Truncate(time.Microsecond)
				return &observedAt
			}()}})
		}
		nodes := []models.Node{node("pci-a", "1eb8:10de", 1), node("pci-b", "2330:10de", 2)}
		_, err = UpsertNodes(ctx, &nodes, clusterID)
		Expect(err).To(BeNil())
		single := node("pci-c", "1eb8:10de", 3)
		_, err = UpsertNode(ctx, &single)
		Expect(err).To(BeNil())
		stored, err := GetNodeInternal(ctx, tenant, clusterID, "pci-a", nil)
		Expect(err).To(BeNil())
		Expect(stored.ToApiModel().GpuInfo.PciDeviceId).To(Equal("1eb8:10de"))
		Expect(stored.GpuPciDeviceIdObservedAt).NotTo(BeNil())
		Expect(stored.GpuPciDeviceIdObservedAt.Equal(*nodes[0].GpuPciDeviceIdObservedAt)).To(BeTrue())
		scopes := authorization.PermittedScopes{Tenants: []string{"49631"}}
		key := NewNodepoolKey(clusterID, "pci-pool")
		groups, err := GetNodepoolsGpuTypes(ctx, tenant, scopes, []NodepoolKey{key})
		Expect(err).To(BeNil())
		Expect(groups[key]).To(Equal([]api.NodepoolGpuTypeDetails{
			{Type: "Same-Name", GpuMemoryBytes: 16106127360, PciDeviceId: "1eb8:10de", TotalDevices: 4},
			{Type: "Same-Name", GpuMemoryBytes: 16106127360, PciDeviceId: "2330:10de", TotalDevices: 2},
		}))
		nodes = []models.Node{node("pci-a", "", 1), node("pci-b", "", 2)}
		_, err = UpsertNodes(ctx, &nodes, clusterID)
		Expect(err).To(BeNil())
		for _, name := range []string{"pci-a", "pci-b"} {
			stored, err := GetNodeInternal(ctx, tenant, clusterID, name, nil)
			Expect(err).To(BeNil())
			Expect(stored.ToApiModel().GpuInfo.PciDeviceId).To(BeEmpty())
			Expect(stored.GpuPciDeviceId).To(BeEmpty())
			Expect(stored.GpuPciDeviceIdObservedAt).To(BeNil())
		}
		replacement := node("pci-c", "", 3)
		_, err = UpsertNode(ctx, &replacement)
		Expect(err).To(BeNil())
		stored, err = GetNodeInternal(ctx, tenant, clusterID, "pci-c", nil)
		Expect(err).To(BeNil())
		Expect(stored.GpuPciDeviceId).To(BeEmpty())
		Expect(stored.GpuPciDeviceIdObservedAt).To(BeNil())
		groups, err = GetNodepoolsGpuTypes(ctx, tenant, scopes, []NodepoolKey{key})
		Expect(err).To(BeNil())
		Expect(groups[key]).To(Equal([]api.NodepoolGpuTypeDetails{{Type: "Same-Name", GpuMemoryBytes: 16106127360, TotalDevices: 6}}))
	})
})

func TestGPUIdentityFreshnessOnIngestAndRead(t *testing.T) {
	now := time.Now()
	fresh, old, future := now.Add(-time.Minute), now.Add(-30*time.Minute), now.Add(10*time.Minute)
	nearFuture := now.Add(30 * time.Second)
	for _, tc := range []struct {
		name, id   string
		observedAt *time.Time
		wantID     string
	}{
		{"fresh collector observation", "1eb8:10de", &fresh, "1eb8:10de"},
		{"permitted collector clock skew", "1eb8:10de", &nearFuture, "1eb8:10de"},
		{"malformed identity", "1EB8:10DE", &fresh, ""},
		{"missing observation", "1eb8:10de", nil, ""},
		{"expired observation", "1eb8:10de", &old, ""},
		{"future observation", "1eb8:10de", &future, ""},
		{"orphan timestamp", "", &fresh, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			node := NodeInfoToNodeDao(1, "cluster", "node", api.NodeInfo{GpuInfo: &api.GpuInfo{GpuType: "Tesla-T4", GpuCount: 1, PciDeviceId: tc.id, PciDeviceIdObservedAt: tc.observedAt}})
			got := node.ToApiModel().GpuInfo
			if got.PciDeviceId != tc.wantID {
				t.Fatalf("ingested/read PCI=%q want=%q", got.PciDeviceId, tc.wantID)
			}
			if tc.wantID == "" && got.PciDeviceIdObservedAt != nil {
				t.Fatalf("unknown ID retained timestamp: %v", got.PciDeviceIdObservedAt)
			}
			if tc.wantID != "" && (got.PciDeviceIdObservedAt == nil || !got.PciDeviceIdObservedAt.Equal(*tc.observedAt)) {
				t.Fatalf("collector time lost or replaced: %v want %v", got.PciDeviceIdObservedAt, tc.observedAt)
			}
			// Simulate a previously stored row without invoking the ingest normalizer.
			raw, err := json.Marshal(map[string]interface{}{"GpuType": "Tesla-T4", "GpuCount": 1, "GpuPciDeviceId": tc.id, "GpuPciDeviceIdObservedAt": tc.observedAt})
			if err != nil {
				t.Fatal(err)
			}
			var stored models.Node
			if err := json.Unmarshal(raw, &stored); err != nil {
				t.Fatal(err)
			}
			read := stored.ToApiModel().GpuInfo
			if read.PciDeviceId != tc.wantID || (tc.wantID == "" && read.PciDeviceIdObservedAt != nil) {
				t.Fatalf("stored row freshness bypassed: %+v", read)
			}
		})
	}
}

var _ = Describe("Node PCI identity expiry without sync", func() {
	It("groups expired, missing and future observations as unknown while retaining distinct fresh IDs", func() {
		ctx := context.Background()
		tenant := int32(49632)
		cluster, err := CreateCluster(ctx, &models.Cluster{Name: "pci-freshness", Domain: "pci.test", Version: "3.0.0", TenantId: tenant})
		Expect(err).To(BeNil())
		clusterID := cluster.ID.String()
		DeferCleanup(func() { Expect(DeleteClusterNodes(ctx, tenant, clusterID)).To(Succeed()) })
		fresh := time.Now().Add(-time.Minute).Truncate(time.Microsecond)
		old, future := fresh.Add(-30*time.Minute), fresh.Add(10*time.Minute)
		nodes := []models.Node{}
		for i, record := range []struct{ name, id string }{{"fresh-a", "1eb8:10de"}, {"fresh-b", "2330:10de"}, {"old-a", "1eb8:10de"}, {"old-b", "2330:10de"}, {"missing", "2237:10de"}, {"future", "2941:10de"}} {
			nodes = append(nodes, *NodeInfoToNodeDao(tenant, clusterID, record.name, api.NodeInfo{Status: "Ready", CreatedAt: time.Now(), NodePool: "pci-pool", GpuInfo: &api.GpuInfo{GpuType: "Same-Name", GpuCount: int32(i + 1), GpuMemoryBytes: 16106127360, PciDeviceId: record.id, PciDeviceIdObservedAt: &fresh}}))
		}
		_, err = UpsertNodes(ctx, &nodes, clusterID)
		Expect(err).To(BeNil())
		// Simulate passage of time and legacy rows directly, without another sync request.
		for _, record := range []struct {
			name       string
			observedAt *time.Time
		}{{"old-a", &old}, {"old-b", &old}, {"missing", nil}, {"future", &future}} {
			_, dbErr := runaidb.DB().WithContext(ctx).Exec("UPDATE node SET gpu_pci_device_id_observed_at = ? WHERE cluster_uuid = ? AND name = ?", record.observedAt, clusterID, record.name)
			Expect(dbErr).NotTo(HaveOccurred())
			stored, getErr := GetNodeInternal(ctx, tenant, clusterID, record.name, nil)
			Expect(getErr).To(BeNil())
			Expect(stored.GpuPciDeviceId).NotTo(BeEmpty())
			Expect(stored.ToApiModel().GpuInfo.PciDeviceId).To(BeEmpty())
			Expect(stored.ToApiModel().GpuInfo.PciDeviceIdObservedAt).To(BeNil())
		}
		key := NewNodepoolKey(clusterID, "pci-pool")
		groups, err := GetNodepoolsGpuTypes(ctx, tenant, authorization.PermittedScopes{Tenants: []string{"49632"}}, []NodepoolKey{key})
		Expect(err).To(BeNil())
		Expect(groups[key]).To(Equal([]api.NodepoolGpuTypeDetails{
			{Type: "Same-Name", GpuMemoryBytes: 16106127360, TotalDevices: 18},
			{Type: "Same-Name", GpuMemoryBytes: 16106127360, PciDeviceId: "1eb8:10de", TotalDevices: 1},
			{Type: "Same-Name", GpuMemoryBytes: 16106127360, PciDeviceId: "2330:10de", TotalDevices: 2},
		}))
		// Raw upserts must not bypass the same pair normalization used by API conversion.
		for _, observedAt := range []*time.Time{nil, &old, &future} {
			raw := nodes[0]
			raw.GpuPciDeviceId = "1eb8:10de"
			raw.GpuPciDeviceIdObservedAt = observedAt
			_, err = UpsertNode(ctx, &raw)
			Expect(err).To(BeNil())
			stored, getErr := GetNodeInternal(ctx, tenant, clusterID, raw.Name, nil)
			Expect(getErr).To(BeNil())
			Expect(stored.GpuPciDeviceId).To(BeEmpty())
			Expect(stored.GpuPciDeviceIdObservedAt).To(BeNil())
		}
	})
})
