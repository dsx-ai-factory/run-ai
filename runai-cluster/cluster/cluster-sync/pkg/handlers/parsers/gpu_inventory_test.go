/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package parsers

import (
	"encoding/json"
	"testing"
	"time"

	syncconfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/gpuinventory"
	rcpconfig "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestHardwareObservationReachesNodeAndFullSync(t *testing.T) {
	if err := rcpconfig.Register[*syncconfig.ClusterSyncConfig](); err != nil {
		t.Fatal(err)
	}
	defer rcpconfig.FlushAll()
	s := runtime.NewScheme()
	if err := corev1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	parser := NewNodeParser(fake.NewClientBuilder().WithScheme(s).Build())
	node := &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "gpu", UID: "node", Labels: map[string]string{gpuinventory.CountLabel: "1", gpuinventory.ProductLabel: "NVIDIA-A10G", gpuinventory.MemoryLabel: "23028"}}}
	node.Status.NodeInfo.BootID = "boot"
	now := time.Now().UTC().Truncate(time.Second)
	obs := gpuinventory.NewObservation(node, []gpuinventory.Device{{UUID: "GPU-test", BusID: "0000:00:1e.0", PCIDeviceID: "2237:10de", Product: "NVIDIA-A10G"}}, now)
	for _, tc := range []struct {
		name    string
		age     time.Duration
		present bool
		want    string
	}{
		{"fresh", 0, true, "2237:10de"}, {"expired", -gpuinventory.MaxAge - time.Second, true, ""}, {"removed", 0, false, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			obs.ObservedAt = now.Add(tc.age)
			b, err := json.Marshal(obs)
			if err != nil {
				t.Fatal(err)
			}
			node.Annotations = map[string]string{}
			if tc.present {
				node.Annotations[gpuinventory.Annotation] = string(b)
			}
			parsed := parser.ParseNodeToServerScheme(node)
			got := parsed.GpuInfo.Get()
			if got == nil || got.GetPciDeviceId() != tc.want || got.GetGpuMemoryBytes() != 24146608128 || got.GpuCount != 1 {
				t.Fatalf("unexpected inventory: %+v", got)
			}
			if tc.want != "" && !got.GetPciDeviceIdObservedAt().Equal(obs.ObservedAt) {
				t.Fatal("observation time was not forwarded")
			}
			if tc.want == "" && got.PciDeviceIdObservedAt.Get() != nil {
				t.Fatal("unknown identity kept an observation time")
			}
			_, full := parser.ParseNodesToServerScheme([]corev1.Node{*node})
			if len(full) != 1 || full[0].NodeInfo.GpuInfo.Get().GetPciDeviceId() != tc.want {
				t.Fatalf("full sync: %+v", full)
			}
		})
	}
	cpu := &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "cpu"}}
	if got := parser.ParseNodeToServerScheme(cpu); got.GpuInfo.Get() != nil {
		t.Fatal("CPU node acquired GPU inventory")
	}
}
