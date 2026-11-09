/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gpuinventory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func testNode() *corev1.Node {
	n := &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "gpu-node", UID: "uid", ResourceVersion: "1",
		Labels: map[string]string{CountLabel: "1", ProductLabel: "NVIDIA-A10G", MemoryLabel: "23028"}}}
	n.Status.NodeInfo.BootID = "boot"
	return n
}
func testDevices() []Device {
	return []Device{{UUID: "GPU-004f2fa1-daf8-62ca-ad5b-6cbdcf805ce0", BusID: "0000:00:1e.0", PCIDeviceID: "2237:10de", Product: "NVIDIA-A10G"}}
}
func annotate(t *testing.T, n *corev1.Node, o Observation) {
	t.Helper()
	b, e := json.Marshal(o)
	if e != nil {
		t.Fatal(e)
	}
	n.Annotations = map[string]string{Annotation: string(b)}
}
func TestReportedIdentityRequiresFreshAssociatedHardware(t *testing.T) {
	now := time.Date(2026, 11, 8, 10, 0, 0, 0, time.UTC)
	cases := []struct {
		name   string
		change func(*corev1.Node, *Observation)
		want   string
	}{
		{"valid", func(_ *corev1.Node, _ *Observation) {}, "2237:10de"},
		{"expired", func(_ *corev1.Node, o *Observation) { o.ObservedAt = now.Add(-MaxAge) }, ""},
		{"future", func(_ *corev1.Node, o *Observation) { o.ObservedAt = now.Add(ClockSkew + time.Second) }, ""},
		{"replaced node", func(n *corev1.Node, _ *Observation) { n.UID = "replacement" }, ""},
		{"reboot", func(n *corev1.Node, _ *Observation) { n.Status.NodeInfo.BootID = "reboot" }, ""},
		{"changed memory", func(n *corev1.Node, _ *Observation) { n.Labels[MemoryLabel] = "40960" }, ""},
		{"partial count", func(n *corev1.Node, o *Observation) { n.Labels[CountLabel] = "2"; o.GPUCount = "2" }, ""},
		{"unknown version", func(_ *corev1.Node, o *Observation) { o.Version = 1 }, ""},
		{"zero PCI", func(_ *corev1.Node, o *Observation) { o.Devices[0].PCIDeviceID = "0000:10de" }, ""},
		{"unassociated product", func(_ *corev1.Node, o *Observation) { o.Devices[0].Product = "Tesla-T4" }, ""},
		{"noncanonical bus", func(_ *corev1.Node, o *Observation) { o.Devices[0].BusID = "00000000:00:1E.0" }, ""},
		{"invalid memory", func(n *corev1.Node, o *Observation) { n.Labels[MemoryLabel] = "0"; o.GPUMemory = "0" }, ""},
		{"missing boot", func(n *corev1.Node, o *Observation) { n.Status.NodeInfo.BootID = ""; o.BootID = "" }, ""},
		{"mixed", func(n *corev1.Node, o *Observation) {
			n.Labels[CountLabel] = "2"
			o.GPUCount = "2"
			o.Devices = append(o.Devices, Device{UUID: "GPU-other", BusID: "0000:00:1f.0", PCIDeviceID: "1eb8:10de", Product: "NVIDIA-A10G"})
		}, ""},
		{"duplicate UUID", func(n *corev1.Node, o *Observation) {
			n.Labels[CountLabel] = "2"
			o.GPUCount = "2"
			d := o.Devices[0]
			d.BusID = "0000:00:1f.0"
			o.Devices = append(o.Devices, d)
		}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			n := testNode()
			o := NewObservation(n, testDevices(), now)
			tc.change(n, &o)
			annotate(t, n, o)
			if got := ReportedIdentity(n, now); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
	n := testNode()
	if got := ReportedIdentity(n, now); got != "" {
		t.Fatal(got)
	}
	n.Annotations = map[string]string{Annotation: "{bad"}
	if got := ReportedIdentity(n, now); got != "" {
		t.Fatal(got)
	}
}
func hardwareFixture(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	sys := filepath.Join(root, "sys")
	proc := filepath.Join(root, "proc")
	files := map[string]string{
		filepath.Join(sys, "bus/pci/devices/0000:00:1e.0/vendor"): "0x10de\n",
		filepath.Join(sys, "bus/pci/devices/0000:00:1e.0/device"): "0x2237\n",
		filepath.Join(sys, "bus/pci/devices/0000:00:1e.0/class"):  "0x030200\n",
		filepath.Join(proc, "0000:00:1e.0/information"):           "Model: NVIDIA A10G\nGPU UUID: GPU-004f2fa1-daf8-62ca-ad5b-6cbdcf805ce0\nBus Location: 0000:00:1e.0\nGPU Excluded: No\n",
	}
	for p, s := range files {
		if e := os.MkdirAll(filepath.Dir(p), 0755); e != nil {
			t.Fatal(e)
		}
		if e := os.WriteFile(p, []byte(s), 0600); e != nil {
			t.Fatal(e)
		}
	}
	return sys, proc
}
func TestReadDevicesUsesPCIHardwareAndDriverAssociation(t *testing.T) {
	sys, proc := hardwareFixture(t)
	ds, e := ReadDevices(sys, proc)
	if e != nil {
		t.Fatal(e)
	}
	if len(ds) != 1 || ds[0] != testDevices()[0] {
		t.Fatalf("%+v", ds)
	}
	if e = os.WriteFile(filepath.Join(sys, "bus/pci/devices/0000:00:1e.0/device"), []byte("0x1eb8"), 0600); e != nil {
		t.Fatal(e)
	}
	ds, e = ReadDevices(sys, proc)
	if e != nil || ds[0].PCIDeviceID != "1eb8:10de" {
		t.Fatalf("name must not determine identity: %+v %v", ds, e)
	}
}
func TestReadDevicesRejectsMissingOrInconsistentEvidence(t *testing.T) {
	for _, change := range []string{"missing driver", "bad bus", "bad vendor", "excluded", "extra GPU"} {
		t.Run(change, func(t *testing.T) {
			sys, proc := hardwareFixture(t)
			switch change {
			case "missing driver":
				if e := os.RemoveAll(proc); e != nil {
					t.Fatal(e)
				}
			case "bad bus":
				if e := os.WriteFile(filepath.Join(proc, "0000:00:1e.0/information"), []byte("Model: NVIDIA A10G\nGPU UUID: GPU-test\nBus Location: 0000:00:1f.0\n"), 0600); e != nil {
					t.Fatal(e)
				}
			case "bad vendor":
				if e := os.WriteFile(filepath.Join(sys, "bus/pci/devices/0000:00:1e.0/vendor"), []byte("garbage"), 0600); e != nil {
					t.Fatal(e)
				}
			case "excluded":
				if e := os.WriteFile(filepath.Join(proc, "0000:00:1e.0/information"), []byte("Model: NVIDIA A10G\nGPU UUID: GPU-test\nBus Location: 0000:00:1e.0\nGPU Excluded: Yes\n"), 0600); e != nil {
					t.Fatal(e)
				}
			case "extra GPU":
				p := filepath.Join(sys, "bus/pci/devices/0000:00:1f.0")
				if e := os.MkdirAll(p, 0755); e != nil {
					t.Fatal(e)
				}
				for k, v := range map[string]string{"vendor": "0x10de", "class": "0x030200", "device": "0x2237"} {
					if e := os.WriteFile(filepath.Join(p, k), []byte(v), 0600); e != nil {
						t.Fatal(e)
					}
				}
			}
			if ds, e := ReadDevices(sys, proc); e == nil {
				t.Fatalf("accepted incomplete hardware: %+v", ds)
			}
		})
	}
}
