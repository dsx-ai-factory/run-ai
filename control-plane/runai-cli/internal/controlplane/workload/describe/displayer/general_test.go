/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package displayer_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/describe/displayer"
)

var _ = Describe("GeneralBuilder", func() {
	It("should build a General correctly", func() {
		createdAt := time.Now()
		updatedAt := createdAt.Add(time.Hour)

		general := displayer.NewGeneralBuilder().
			Name("test-workload").
			ID("123").
			Type("Job").
			Phase(displayer.Running).
			PhaseMessage("Running smoothly").
			Project("project1").
			Department("dept1").
			CreatedAt(createdAt).
			UpdatedAt(updatedAt).
			SubmittedBy("user1").
			Priority("high").
			Category("training").
			Topology("mnnvl").
			DistributedFramework("pytorch").
			EnvironmentVariables(map[string]string{"ENV1": "value1"}).
			Arguments("arg1 arg2").
			Command("echo hello").
			RunaiCliCommand("runai submit --image=ubuntu").
			Build()

		Expect(general.Name).To(Equal("test-workload"))
		Expect(general.ID).To(Equal("123"))
		Expect(general.Type).To(Equal("Job"))
		Expect(general.Phase).To(Equal(displayer.Running))
		Expect(*general.PhaseMessage).To(Equal("Running smoothly"))
		Expect(general.Project).To(Equal("project1"))
		Expect(general.Department).To(Equal("dept1"))
		Expect(general.CreatedAt).To(Equal(createdAt))
		Expect(general.UpdatedAt).To(Equal(updatedAt))
		Expect(*general.SubmittedBy).To(Equal("user1"))
		Expect(*general.Priority).To(Equal("high"))
		Expect(general.Category).To(Equal("training"))
		Expect(*general.Topology).To(Equal("mnnvl"))
		Expect(*general.DistributedFramework).To(Equal("pytorch"))
		Expect(*general.EnvironmentVariables).To(Equal(map[string]string{"ENV1": "value1"}))
		Expect(*general.Arguments).To(Equal("arg1 arg2"))
		Expect(*general.Command).To(Equal("echo hello"))
		Expect(*general.RunaiCliCommand).To(Equal("runai submit --image=ubuntu"))
	})
})
