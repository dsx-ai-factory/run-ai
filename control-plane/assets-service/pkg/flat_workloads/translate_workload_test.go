/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package flat_workloads

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	policyconstants "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("uid/gid derivation tests", func() {
	Context("translate preemptible to priority class", func() {
		It("should translate correctly", func() {
			payloadMap := make(map[string]interface{})

			handleOverQuota(nil, auth.WorkloadTypeWorkspace, payloadMap)
			Expect(payloadMap["priorityClass"]).To(BeNil())

			handleOverQuota(convutils.Ref(false), auth.WorkloadTypeWorkspace, payloadMap)
			Expect(payloadMap["priorityClass"]).To(Equal(policyconstants.PriorityClassBuild))

			handleOverQuota(convutils.Ref(true), auth.WorkloadTypeWorkspace, payloadMap)
			Expect(payloadMap["priorityClass"]).To(Equal(policyconstants.PriorityClassInteractivePreemptible))
		})
	})
})
