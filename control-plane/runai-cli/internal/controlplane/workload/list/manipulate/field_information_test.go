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

var _ = Describe("ParseWorkloadColumns function", func() {
	Context("when columns slice is empty", func() {
		It("should return nil", func() {
			result, err := ParseWorkloadColumns([]string{})
			Expect(err).To(BeNil())
			Expect(result).To(BeNil())
		})

		It("should return nil for nil input", func() {
			result, err := ParseWorkloadColumns(nil)
			Expect(err).To(BeNil())
			Expect(result).To(BeNil())
		})
	})

	Context("when columns contain valid values", func() {
		It("should parse single valid column", func() {
			result, err := ParseWorkloadColumns([]string{"name"})
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))
			Expect(result[0]).To(Equal(ColumnKeyName))
		})

		It("should parse multiple valid columns", func() {
			result, err := ParseWorkloadColumns([]string{"name", "type", "status"})
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyType, ColumnKeyStatus))
		})

		It("should handle case insensitive input", func() {
			result, err := ParseWorkloadColumns([]string{"NAME", "Type", "STATUS"})
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyType, ColumnKeyStatus))
		})

		It("should trim whitespace", func() {
			result, err := ParseWorkloadColumns([]string{"  name  ", " type ", "status"})
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyType, ColumnKeyStatus))
		})

		It("should parse all available columns", func() {
			allColumnNames := []string{
				"name", "type", "framework", "status", "project", "category",
				"priority", "department", "submittedby", "createdat",
				"totalpendingtimeseconds", "allocatedgpumemory", "allocatedcpucompute",
				"allocatedcpumemory", "completedat", "urls", "preemptible",
				"pods", "allocatedgpu",
			}
			result, err := ParseWorkloadColumns(allColumnNames)
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(len(allColumnNames)))
		})
	})

	Context("when columns contain invalid values", func() {
		It("should return error for single invalid column", func() {
			result, err := ParseWorkloadColumns([]string{"invalid"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid column names: \"invalid\""))
			Expect(result).To(BeNil())
		})

		It("should return error for multiple invalid columns", func() {
			result, err := ParseWorkloadColumns([]string{"invalid1", "", "invalid2"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid column names: \"invalid1\", \"\", \"invalid2\""))
			Expect(result).To(BeNil())
		})

		It("should return error when mixing valid and invalid columns", func() {
			result, err := ParseWorkloadColumns([]string{"name", "invalid", "type"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid column names: \"invalid\""))
			Expect(result).To(BeNil())
		})

		It("should return error for partially matching column names", func() {
			result, err := ParseWorkloadColumns([]string{"nam", "typ"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid column names: \"nam\", \"typ\""))
			Expect(result).To(BeNil())
		})
	})

	Context("when columns contain duplicates", func() {
		It("should include duplicates in result", func() {
			result, err := ParseWorkloadColumns([]string{"name", "name", "type"})
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))
			Expect(result[0]).To(Equal(ColumnKeyName))
			Expect(result[1]).To(Equal(ColumnKeyName))
			Expect(result[2]).To(Equal(ColumnKeyType))
		})
	})
})

var _ = Describe("Extract Functions", func() {
	var (
		workload    *workloadservice.Workload
		workloadRow *WorkloadsAlignedRow
	)

	BeforeEach(func() {
		workload = &workloadservice.Workload{}
		workloadRow = &WorkloadsAlignedRow{}
	})

	Context("extractFrameworkData", func() {
		It("should set default framework when DistributedFramework is nil", func() {
			workload.DistributedFramework = nil

			extractFrameworkData(workload, workloadRow)

			Expect(workloadRow.DistributedFramework).ToNot(BeNil())
			Expect(*workloadRow.DistributedFramework).To(Equal(RunaiStandardFramework))
		})

		It("should use workload's DistributedFramework when set", func() {
			framework := "PyTorch"
			workload.DistributedFramework = &framework

			extractFrameworkData(workload, workloadRow)

			Expect(workloadRow.DistributedFramework).ToNot(BeNil())
			Expect(*workloadRow.DistributedFramework).To(Equal("PyTorch"))
		})
	})

	Context("extractStatusData", func() {
		It("should extract valid phase", func() {
			workload.Phase = workloadservice.PHASE_RUNNING

			extractStatusData(workload, workloadRow)

			Expect(workloadRow.Status).ToNot(BeNil())
			Expect(*workloadRow.Status).To(Equal(workloadservice.PHASE_RUNNING))
		})

		It("should set UNKNOWN phase for empty phase", func() {
			workload.Phase = ""

			extractStatusData(workload, workloadRow)

			Expect(workloadRow.Status).ToNot(BeNil())
			Expect(*workloadRow.Status).To(Equal(workloadservice.PHASE_UNKNOWN))
		})

		It("should set UNKNOWN phase for invalid phase", func() {
			// Create an invalid phase by casting
			workload.Phase = workloadservice.Phase("InvalidPhase")

			extractStatusData(workload, workloadRow)

			Expect(workloadRow.Status).ToNot(BeNil())
			Expect(*workloadRow.Status).To(Equal(workloadservice.PHASE_UNKNOWN))
		})
	})

	Context("extractPreemptibleData", func() {
		It("should extract preemptible value when set to true", func() {
			workload.Preemptible = *workloadservice.NewNullableBool(ptr.To(true))

			extractPreemptibleData(workload, workloadRow)

			Expect(workloadRow.Preemptible).ToNot(BeNil())
			Expect(*workloadRow.Preemptible).To(BeTrue())
		})

		It("should extract preemptible value when set to false", func() {
			workload.Preemptible = *workloadservice.NewNullableBool(ptr.To(false))

			extractPreemptibleData(workload, workloadRow)

			Expect(workloadRow.Preemptible).ToNot(BeNil())
			Expect(*workloadRow.Preemptible).To(BeFalse())
		})

		It("should leave preemptible nil when not set", func() {
			workload.Preemptible = workloadservice.NullableBool{}

			extractPreemptibleData(workload, workloadRow)

			Expect(workloadRow.Preemptible).To(BeNil())
		})
	})
})

var _ = Describe("Format Functions", func() {
	var workloadRow *WorkloadsAlignedRow

	BeforeEach(func() {
		workloadRow = &WorkloadsAlignedRow{}
	})

	Context("formatTypeValue", func() {
		It("should return type as-is for non-Distributed types", func() {
			workloadRow.Type = "Workspace"

			result := formatTypeValue(workloadRow)

			Expect(result).To(Equal("Workspace"))
		})

		It("should convert Distributed to Training", func() {
			workloadRow.Type = "Distributed"

			result := formatTypeValue(workloadRow)

			Expect(result).To(Equal("Training"))
		})

		It("should handle empty type", func() {
			workloadRow.Type = ""

			result := formatTypeValue(workloadRow)

			Expect(result).To(Equal(""))
		})
	})

	Context("formatStatusValue", func() {
		It("should return status string when status is set", func() {
			status := workloadservice.PHASE_RUNNING
			workloadRow.Status = &status

			result := formatStatusValue(workloadRow)

			Expect(result).To(Equal("Running"))
		})

		It("should return empty string when status is nil", func() {
			workloadRow.Status = nil

			result := formatStatusValue(workloadRow)

			Expect(result).To(Equal(""))
		})

		It("should handle different status values", func() {
			statuses := []workloadservice.Phase{
				workloadservice.PHASE_PENDING,
				workloadservice.PHASE_RUNNING,
				workloadservice.PHASE_FAILED,
				workloadservice.PHASE_COMPLETED,
			}

			for _, status := range statuses {
				workloadRow.Status = &status
				result := formatStatusValue(workloadRow)
				Expect(result).To(Equal(string(status)))
			}
		})
	})

	Context("formatPreemptibleValue", func() {
		It("should return 'Yes' when preemptible is true", func() {
			workloadRow.Preemptible = ptr.To(true)

			result := formatPreemptibleValue(workloadRow)

			Expect(result).To(Equal("Yes"))
		})

		It("should return 'No' when preemptible is false", func() {
			workloadRow.Preemptible = ptr.To(false)

			result := formatPreemptibleValue(workloadRow)

			Expect(result).To(Equal("No"))
		})

		It("should return '-' when preemptible is nil", func() {
			workloadRow.Preemptible = nil

			result := formatPreemptibleValue(workloadRow)

			Expect(result).To(Equal("-"))
		})
	})

	Context("formatPodsValue", func() {
		BeforeEach(func() {
			workloadRow.Pods = &Pods{}
		})

		It("should format fixed number pods correctly", func() {
			workloadRow.Pods.RunningPods = ptr.To(int32(3))
			workloadRow.Pods.RequestedPods = ptr.To(int32(5))

			result := formatPodsValue(workloadRow)

			Expect(result).To(Equal("3/5"))
		})

		It("should format range pods correctly", func() {
			workloadRow.Pods.RunningPods = ptr.To(int32(3))
			workloadRow.Pods.RequestedPodsMin = ptr.To(int32(2))
			workloadRow.Pods.RequestedPodsMax = ptr.To(int32(8))

			result := formatPodsValue(workloadRow)

			Expect(result).To(Equal("3/2-8"))
		})

		It("should format parallelism pods correctly", func() {
			workloadRow.Pods.RunningPods = ptr.To(int32(4))
			workloadRow.Pods.Parallelism = ptr.To(int32(6))

			result := formatPodsValue(workloadRow)

			Expect(result).To(Equal("4/6"))
		})

		It("should format completions pods correctly", func() {
			workloadRow.Pods.RunningPods = ptr.To(int32(2))
			workloadRow.Pods.Completions = ptr.To(int32(10))

			result := formatPodsValue(workloadRow)

			Expect(result).To(Equal("2/-"))
		})

		It("should handle nil running pods", func() {
			workloadRow.Pods.RunningPods = nil
			workloadRow.Pods.RequestedPods = ptr.To(int32(5))

			result := formatPodsValue(workloadRow)

			Expect(result).To(Equal("0/5"))
		})

		It("should return default format when no fields are set", func() {
			workloadRow.Pods.RunningPods = ptr.To(int32(1))

			result := formatPodsValue(workloadRow)

			Expect(result).To(Equal("1/-"))
		})

		It("should prioritize RequestedPods over range", func() {
			workloadRow.Pods.RunningPods = ptr.To(int32(3))
			workloadRow.Pods.RequestedPods = ptr.To(int32(5))
			workloadRow.Pods.RequestedPodsMin = ptr.To(int32(2))
			workloadRow.Pods.RequestedPodsMax = ptr.To(int32(8))

			result := formatPodsValue(workloadRow)

			Expect(result).To(Equal("3/5"))
		})
	})
})

var _ = Describe("getPodsValue Helper Function", func() {
	var workload *workloadservice.Workload

	BeforeEach(func() {
		workload = &workloadservice.Workload{
			RunningPods: 3,
		}
	})

	It("should handle nil RequestedPods", func() {
		workload.RequestedPods = nil

		result := getPodsValue(workload)

		Expect(result.RunningPods).To(Equal(ptr.To(int32(3))))
		Expect(result.RequestedPods).To(BeNil())
	})

	It("should extract RequestedPods.Number when set", func() {
		workload.RequestedPods = &workloadservice.RequestedPods{
			Number: *workloadservice.NewNullableInt32(ptr.To(int32(5))),
		}

		result := getPodsValue(workload)

		Expect(result.RunningPods).To(Equal(ptr.To(int32(3))))
		Expect(result.RequestedPods).To(Equal(ptr.To(int32(5))))
	})

	It("should extract min/max when both are set", func() {
		workload.RequestedPods = &workloadservice.RequestedPods{
			Min: *workloadservice.NewNullableInt32(ptr.To(int32(2))),
			Max: *workloadservice.NewNullableInt32(ptr.To(int32(8))),
		}

		result := getPodsValue(workload)

		Expect(result.RunningPods).To(Equal(ptr.To(int32(3))))
		Expect(result.RequestedPodsMin).To(Equal(ptr.To(int32(2))))
		Expect(result.RequestedPodsMax).To(Equal(ptr.To(int32(8))))
	})

	It("should extract parallelism when set", func() {
		workload.RequestedPods = &workloadservice.RequestedPods{
			Parallelism: *workloadservice.NewNullableInt32(ptr.To(int32(4))),
		}

		result := getPodsValue(workload)

		Expect(result.RunningPods).To(Equal(ptr.To(int32(3))))
		Expect(result.Parallelism).To(Equal(ptr.To(int32(4))))
	})

	It("should extract completions when set", func() {
		workload.RequestedPods = &workloadservice.RequestedPods{
			Completions: *workloadservice.NewNullableInt32(ptr.To(int32(10))),
		}

		result := getPodsValue(workload)

		Expect(result.RunningPods).To(Equal(ptr.To(int32(3))))
		Expect(result.Completions).To(Equal(ptr.To(int32(10))))
	})

	It("should return default when RequestedPods is empty", func() {
		workload.RequestedPods = &workloadservice.RequestedPods{}

		result := getPodsValue(workload)

		Expect(result.RunningPods).To(Equal(ptr.To(int32(3))))
		Expect(result.RequestedPods).To(BeNil())
		Expect(result.RequestedPodsMin).To(BeNil())
		Expect(result.RequestedPodsMax).To(BeNil())
		Expect(result.Parallelism).To(BeNil())
		Expect(result.Completions).To(BeNil())
	})
})
