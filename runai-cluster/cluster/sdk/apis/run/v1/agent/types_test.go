/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package agent

import (
	"encoding/json"

	runv1internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Agent", func() {
	It("retains the deprecated workload runner flag for rollback compatibility", func() {
		agent := Agent{}
		Expect(json.Unmarshal([]byte(`{"enableWorkloadRunner":false}`), &agent)).To(Succeed())

		internalSpec := &runv1internal.Internal{}
		internalSpec.SetDefaultsWhereNeeded()
		agent.SetDefaultsWhereNeeded(internalSpec, nil)

		data, err := json.Marshal(agent)

		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(ContainSubstring(`"enableWorkloadRunner":false`))
	})

	It("does not default the deprecated workload runner flag", func() {
		internalSpec := &runv1internal.Internal{}
		internalSpec.SetDefaultsWhereNeeded()
		agent := Agent{}
		agent.SetDefaultsWhereNeeded(internalSpec, nil)

		data, err := json.Marshal(agent)

		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).NotTo(ContainSubstring("enableWorkloadRunner"))
	})
})
