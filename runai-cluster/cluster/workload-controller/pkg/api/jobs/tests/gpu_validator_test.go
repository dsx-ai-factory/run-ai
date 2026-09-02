/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	fields "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	workloadUtils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource/workload"
	"k8s.io/utils/ptr"
)

const (
	NoError = ""

	FractionalNotSupported = "fractional GPU is not supported in this cluster, please use the 'gpuDevicesRequest' field"
)

type workloadGPUTestEntry struct {
	description string
	workload    *jobs.JobWorkload
	NvidiaError string
}

type workloadGPUSpecFields struct {
	gpuDevices     *int
	gpu            *float64
	gpuLimit       *float64
	gpuMemory      string
	gpuMemoryLimit string
}

const (
	moreThanOneInteger    = 4
	smallFractionalNumber = 0.1
	someMemorySize        = "1G"
)

func createGinkgoTestEntryFromWorkloadTestEntries(tests []workloadGPUTestEntry) []TableEntry {
	var testEntries []TableEntry
	for _, entry := range tests {
		testEntries = append(testEntries, Entry(entry.description, entry.workload, entry.NvidiaError))
	}
	return testEntries
}

func createDistributedWorkloadGPUTestEntries() []workloadGPUTestEntry {
	tests := []workloadGPUTestEntry{
		{
			description: "distributed workload + gpu: master -> integer, worker -> nil",
			workload: CreateGPUWorkloadWithMaster(
				// worker
				workloadGPUSpecFields{},
				// master
				&workloadGPUSpecFields{
					gpu: ptr.To[float64](moreThanOneInteger),
				}),
			NvidiaError: NoError,
		},
		{
			description: "distributed workload + gpu: master -> fractional, worker -> nil",
			workload: CreateGPUWorkloadWithMaster(
				// worker
				workloadGPUSpecFields{},
				// master
				&workloadGPUSpecFields{
					gpu: ptr.To[float64](smallFractionalNumber),
				}),
			NvidiaError: NoError,
		},
		{
			description: "distributed workload + gpu: master -> nil, worker -> integer",
			workload: CreateGPUWorkloadWithMaster(
				// worker
				workloadGPUSpecFields{
					gpu: ptr.To[float64](moreThanOneInteger),
				},
				// master
				&workloadGPUSpecFields{}),
			NvidiaError: NoError,
		},
		{
			description: "distributed workload + gpu: master -> nil, worker -> fractional",
			workload: CreateGPUWorkloadWithMaster(
				// worker
				workloadGPUSpecFields{
					gpu: ptr.To[float64](smallFractionalNumber),
				},
				// master
				&workloadGPUSpecFields{}),
			NvidiaError: NoError,
		},
		{
			description: "distributed workload + gpu: master -> integer, worker -> integer",
			workload: CreateGPUWorkloadWithMaster(
				// worker
				workloadGPUSpecFields{
					gpu: ptr.To[float64](moreThanOneInteger),
				},
				// master
				&workloadGPUSpecFields{
					gpu: ptr.To[float64](moreThanOneInteger),
				}),
			NvidiaError: NoError,
		},
		{
			description: "distributed workload + gpu: master -> fractional, worker -> fractional",
			workload: CreateGPUWorkloadWithMaster(
				// worker
				workloadGPUSpecFields{
					gpu: ptr.To[float64](smallFractionalNumber),
				},
				// master
				&workloadGPUSpecFields{
					gpu: ptr.To[float64](smallFractionalNumber),
				}),
			NvidiaError: NoError,
		},
		{
			description: "distributed workload + gpu: master -> integer, worker -> fractional",
			workload: CreateGPUWorkloadWithMaster(
				// worker
				workloadGPUSpecFields{
					gpu: ptr.To[float64](smallFractionalNumber),
				},
				// master
				&workloadGPUSpecFields{
					gpu: ptr.To[float64](moreThanOneInteger),
				}),
			NvidiaError: NoError,
		},
	}
	return tests
}

var _ = Describe("Workload GPU Validator - Distributed Workloads", func() {

	var (
		validator                          jobs.WorkloadGpuValidator
		ctx                                context.Context
		distributedWorkloadsGpuTestEntries []workloadGPUTestEntry
	)
	distributedWorkloadsGpuTestEntries = createDistributedWorkloadGPUTestEntries()

	Context("nvidia cluster", func() {
		var nvidiaTestWorkloadEntries []TableEntry

		nvidiaTestWorkloadEntries = createGinkgoTestEntryFromWorkloadTestEntries(distributedWorkloadsGpuTestEntries)

		BeforeEach(func() {
			validator = jobs.NewWorkloadGpuValidator()
		})

		DescribeTable("Validate", func(workload *jobs.JobWorkload, expectedError string) {
			checkValidatorWorkload(ctx, validator, workload, expectedError)
		}, nvidiaTestWorkloadEntries)
	})
})

func CreateGPUWorkloadWithMaster(masterResourceFields workloadGPUSpecFields, workerResourceFields *workloadGPUSpecFields) *jobs.JobWorkload {
	jobWorkload := workloadUtils.NewJobWorkloadBuilder().WithName(workloadName).WithMasterSpec(&jobs.JobSpec{}).Build()
	if workerResourceFields != nil {
		setSpecWorkloadGpuProfile(&jobWorkload.Spec, *workerResourceFields)
	}
	setSpecWorkloadGpuProfile(jobWorkload.MasterSpec, masterResourceFields)
	return jobWorkload
}

func setSpecWorkloadGpuProfile(jobSpec *jobs.JobSpec, resourceFields workloadGPUSpecFields) {
	jobSpec.ResourceAllocationFields = field_groups.ResourceAllocationFields{
		GpuDevices:     fields.NewOptionalIntField(resourceFields.gpuDevices),
		Gpu:            fields.NewOptionalNumberField(resourceFields.gpu),
		GpuLimit:       fields.NewOptionalNumberField(resourceFields.gpuLimit),
		GpuMemory:      fields.NewOptionalQuantityField(resourceFields.gpuMemory, ""),
		GpuMemoryLimit: fields.NewOptionalQuantityField(resourceFields.gpuMemoryLimit, ""),
	}
}

func checkValidatorWorkload(ctx context.Context, validator jobs.WorkloadGpuValidator, workload *jobs.JobWorkload, expectedError string) {
	err := validator.Validate(ctx, workload)
	if expectedError == NoError {
		Expect(err).To(Succeed())
	} else {
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(expectedError))
	}
}
