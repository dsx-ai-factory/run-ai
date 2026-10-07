/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

func TestGenerateComputeSpecGpuDevicesRequest(t *testing.T) {
	for _, tc := range []struct {
		name    string
		devices *int32
		want    string
	}{
		{name: "omitted", want: `{"cpuCoreRequest":1}`},
		{name: "explicit zero", devices: ptr.To(int32(0)), want: `{"cpuCoreRequest":1,"gpuDevicesRequest":0}`},
		{name: "positive", devices: ptr.To(int32(2)), want: `{"cpuCoreRequest":1,"gpuDevicesRequest":2}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			compute := &Compute{}
			compute.CpuCoreRequest.Set(ptr.To(float64(1)))
			if tc.devices != nil {
				compute.GpuDeviceRequest.Set(tc.devices)
			}

			spec, err := compute.GenerateComputeSpec()
			g.Expect(err).NotTo(HaveOccurred())
			payload, err := json.Marshal(spec)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(payload).To(MatchJSON(tc.want))
		})
	}
}
