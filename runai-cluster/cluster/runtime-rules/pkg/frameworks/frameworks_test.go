/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package frameworks

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/extensions"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/extensions/distributed"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GetDistributedJobFrameworksStatus", func() {
	var (
		extensionsGroup extensions.Group
	)

	BeforeEach(func() {
		extensionsGroup = extensions.Group{
			distributed.MPI: extensions.Extension{
				Type: distributed.MPI,
				Status: extensions.Status{
					Enabled: true,
					Capabilities: map[extensions.CapabilityType]bool{
						distributed.Suspension: true,
					},
				},
			},
			distributed.Jax: extensions.Extension{
				Type: distributed.Jax,
				Status: extensions.Status{
					Enabled:      true,
					Capabilities: map[extensions.CapabilityType]bool{},
				},
			},
			distributed.XGBoost: extensions.Extension{
				Type: distributed.XGBoost,
				Status: extensions.Status{
					Enabled: false,
				},
			},
		}
	})

	Context("When all extensions are enabled", func() {
		It("should return the correct DistributedJobFrameworksStatus map", func() {
			statusMap, err := GetDistributedJobFrameworksStatus(extensionsGroup)
			Expect(err).To(BeNil())
			Expect(statusMap).To(HaveKey(constants.KindMpiJob))
			Expect(statusMap).To(HaveKey(constants.KindJAXJob))
			Expect(statusMap).NotTo(HaveKey(constants.KindXGBoostJob))

			// Verify properties of returned statuses
			Expect(statusMap[constants.KindMpiJob].SuspendEnabled).To(BeTrue())
			Expect(statusMap[constants.KindJAXJob].SuspendEnabled).To(BeFalse())
		})
	})

	Context("When an extension does not have a matching job kind", func() {
		BeforeEach(func() {
			// Add an unsupported extension
			extensionsGroup["Unsupported"] = extensions.Extension{
				Type: "Unsupported",
				Status: extensions.Status{
					Enabled: true,
				},
			}
		})

		It("should return an error", func() {
			_, err := GetDistributedJobFrameworksStatus(extensionsGroup)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("the job kind for the extension"))
		})
	})
})
