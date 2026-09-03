/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe_test

import (
	"time"

	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/describe"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/describe/displayer"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("WorkloadBuilder", func() {
	It("should build a Workload correctly", func() {
		createdAt := time.Now()
		submittedBy := "user1"
		framework := "pytorch"

		general := displayer.NewGeneralBuilder().
			Name("test-workload").
			Type("Job").
			Phase(displayer.Running).
			PhaseMessage("Running smoothly").
			Project("project1").
			Department("dept1").
			CreatedAt(createdAt).
			SubmittedBy(submittedBy).
			DistributedFramework(framework).
			RunaiCliCommand("runai training submit --image=ubuntu").
			Command("echo 'Hello, World!'").
			Arguments("arg1 arg2").
			EnvironmentVariables(map[string]string{"ENV1": "value1", "ENV2": "value2"}).
			Build()

		runtime := displayer.NewRuntimeBuilder().
			Preemptible(true).
			Images([]string{"image1", "image2"}).
			Build()

		workload := describe.NewWorkloadBuilder().
			SetGeneral(general).
			SetRuntime(runtime).
			GPUNetworkAcceleration(ptr.To(true)).
			Build()

		Expect(workload.General.Name).To(Equal("test-workload"))
		Expect(workload.General.Type).To(Equal("Job"))
		Expect(workload.General.Phase).To(Equal(displayer.Running))
		Expect(*workload.General.PhaseMessage).To(Equal("Running smoothly"))
		Expect(workload.General.Project).To(Equal("project1"))
		Expect(workload.General.Department).To(Equal("dept1"))
		Expect(workload.General.CreatedAt).To(Equal(createdAt))
		Expect(*workload.General.SubmittedBy).To(Equal(submittedBy))
		Expect(*workload.General.DistributedFramework).To(Equal(framework))
		Expect(*workload.Runtime.Preemptible).To(BeTrue())
		Expect(workload.Runtime.Images).To(Equal([]string{"image1", "image2"}))
		Expect(*workload.General.RunaiCliCommand).To(Equal("runai training submit --image=ubuntu"))
		Expect(*workload.General.Command).To(Equal("echo 'Hello, World!'"))
		Expect(*workload.General.Arguments).To(Equal("arg1 arg2"))
		Expect(*workload.General.EnvironmentVariables).To(Equal(map[string]string{"ENV1": "value1", "ENV2": "value2"}))
		Expect(*workload.GPUAcceleration).To(Equal(true))
	})

	It("should set NetworkTopology correctly", func() {
		nt := &displayer.NetworkTopology{
			Name:               ptr.To("my-topology"),
			RequiredPlacement:  ptr.To("rack"),
			PreferredPlacement: ptr.To("block"),
			ActualPlacement:    ptr.To("rack (rack-1)"),
		}

		workload := describe.NewWorkloadBuilder().
			SetNetworkTopology(nt).
			Build()

		Expect(workload.NetworkTopology).NotTo(BeNil())
		Expect(*workload.NetworkTopology.Name).To(Equal("my-topology"))
		Expect(*workload.NetworkTopology.RequiredPlacement).To(Equal("rack"))
		Expect(*workload.NetworkTopology.PreferredPlacement).To(Equal("block"))
		Expect(*workload.NetworkTopology.ActualPlacement).To(Equal("rack (rack-1)"))
	})

	It("should leave NetworkTopology nil when not set", func() {
		workload := describe.NewWorkloadBuilder().
			Build()

		Expect(workload.NetworkTopology).To(BeNil())
	})
})
