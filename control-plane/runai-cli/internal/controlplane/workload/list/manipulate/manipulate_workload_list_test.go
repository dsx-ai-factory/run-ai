/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	workloadservice "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

var _ = Describe("getPodsValue", func() {
	It("should handle nil RequestedPods", func() {
		workload := &workloadservice.Workload{
			RunningPods:   5,
			RequestedPods: nil,
		}
		expected := &Pods{
			RunningPods: ptr.To(int32(5)),
		}

		result := getPodsValue(workload)
		Expect(result).To(Equal(expected))
	})

	It("should handle RequestedPods.Number", func() {
		workload := &workloadservice.Workload{
			RunningPods: 3,
			RequestedPods: &workloadservice.RequestedPods{
				Number: *workloadservice.NewNullableInt32(ptr.To(int32(8))),
			},
		}
		expected := &Pods{
			RunningPods:   ptr.To(int32(3)),
			RequestedPods: ptr.To(int32(8)),
		}

		result := getPodsValue(workload)
		Expect(result).To(Equal(expected))
	})

	It("should handle RequestedPods.Min and Max", func() {
		workload := &workloadservice.Workload{
			RunningPods: 2,
			RequestedPods: &workloadservice.RequestedPods{
				Min: *workloadservice.NewNullableInt32(ptr.To(int32(2))),
				Max: *workloadservice.NewNullableInt32(ptr.To(int32(10))),
			},
		}
		expected := &Pods{
			RunningPods:      ptr.To(int32(2)),
			RequestedPodsMin: ptr.To(int32(2)),
			RequestedPodsMax: ptr.To(int32(10)),
		}

		result := getPodsValue(workload)
		Expect(result).To(Equal(expected))
	})

	It("should handle RequestedPods.Parallelism", func() {
		workload := &workloadservice.Workload{
			RunningPods: 4,
			RequestedPods: &workloadservice.RequestedPods{
				Parallelism: *workloadservice.NewNullableInt32(ptr.To(int32(6))),
			},
		}
		expected := &Pods{
			RunningPods: ptr.To(int32(4)),
			Parallelism: ptr.To(int32(6)),
		}

		result := getPodsValue(workload)
		Expect(result).To(Equal(expected))
	})

	It("should handle zero running pods", func() {
		workload := &workloadservice.Workload{
			RunningPods: 0,
			RequestedPods: &workloadservice.RequestedPods{
				Number: *workloadservice.NewNullableInt32(ptr.To(int32(3))),
			},
		}
		expected := &Pods{
			RunningPods:   ptr.To(int32(0)),
			RequestedPods: ptr.To(int32(3)),
		}

		result := getPodsValue(workload)
		Expect(result).To(Equal(expected))
	})

	It("should prioritize Number over other fields", func() {
		workload := &workloadservice.Workload{
			RunningPods: 3,
			RequestedPods: &workloadservice.RequestedPods{
				Number:      *workloadservice.NewNullableInt32(ptr.To(int32(5))),
				Parallelism: *workloadservice.NewNullableInt32(ptr.To(int32(7))),
			},
		}
		expected := &Pods{
			RunningPods:   ptr.To(int32(3)),
			RequestedPods: ptr.To(int32(5)),
		}

		result := getPodsValue(workload)
		Expect(result).To(Equal(expected))
		Expect(result.Parallelism).To(BeNil()) // Should not be set due to priority
	})
})

var _ = Describe("formatPodsValue", func() {
	It("should format fixed number pods", func() {
		workloadRow := WorkloadsAlignedRow{
			Pods: &Pods{
				RunningPods:   ptr.To(int32(3)),
				RequestedPods: ptr.To(int32(5)),
			},
		}

		result := formatPodsValue(ptr.To(workloadRow))
		Expect(result).To(Equal("3/5"))
	})

	It("should format range pods", func() {
		workloadRow := WorkloadsAlignedRow{
			Pods: &Pods{
				RunningPods:      ptr.To(int32(2)),
				RequestedPodsMin: ptr.To(int32(1)),
				RequestedPodsMax: ptr.To(int32(8)),
			},
		}

		result := formatPodsValue(ptr.To(workloadRow))
		Expect(result).To(Equal("2/1-8"))
	})

	It("should format parallelism pods", func() {
		workloadRow := WorkloadsAlignedRow{
			Pods: &Pods{
				RunningPods: ptr.To(int32(4)),
				Parallelism: ptr.To(int32(6)),
			},
		}

		result := formatPodsValue(ptr.To(workloadRow))
		Expect(result).To(Equal("4/6"))
	})

	It("should handle nil running pods", func() {
		workloadRow := WorkloadsAlignedRow{
			Pods: &Pods{
				RunningPods:   nil,
				RequestedPods: ptr.To(int32(3)),
			},
		}

		result := formatPodsValue(ptr.To(workloadRow))
		Expect(result).To(Equal("0/3"))
	})

	It("should return running pods only when no requested pods", func() {
		workloadRow := WorkloadsAlignedRow{
			Pods: &Pods{
				RunningPods: ptr.To(int32(2)),
			},
		}
		result := formatPodsValue(ptr.To(workloadRow))
		Expect(result).To(Equal("2/-"))
	})
})

var _ = Describe("formatTypeValue", func() {
	It("should transform Distributed to Training", func() {
		workloadRow := WorkloadsAlignedRow{
			Type: "Distributed",
		}

		result := formatTypeValue(ptr.To(workloadRow))
		Expect(result).To(Equal("Training"))
	})

	It("should keep other types unchanged", func() {
		workloadRow := WorkloadsAlignedRow{
			Type: "Inference",
		}

		result := formatTypeValue(ptr.To(workloadRow))
		Expect(result).To(Equal("Inference"))
	})
})

var _ = Describe("formatPreemptibleValue", func() {
	It("should return Yes for true", func() {
		workloadRow := WorkloadsAlignedRow{
			Preemptible: ptr.To(true),
		}

		result := formatPreemptibleValue(ptr.To(workloadRow))
		Expect(result).To(Equal("Yes"))
	})

	It("should return No for false", func() {
		workloadRow := WorkloadsAlignedRow{
			Preemptible: ptr.To(false),
		}

		result := formatPreemptibleValue(ptr.To(workloadRow))
		Expect(result).To(Equal("No"))
	})

	It("should return dash for nil", func() {
		workloadRow := WorkloadsAlignedRow{
			Preemptible: nil,
		}

		result := formatPreemptibleValue(ptr.To(workloadRow))
		Expect(result).To(Equal("-"))
	})
})
