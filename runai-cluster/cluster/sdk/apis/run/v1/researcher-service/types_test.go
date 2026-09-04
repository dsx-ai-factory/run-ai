/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package researcher_service

import (
	. "github.com/onsi/ginkgo/v2"
	"k8s.io/utils/ptr"

	. "github.com/onsi/gomega"
)

var _ = Describe("Researcher Service", func() {
	var (
		researcherService *ResearcherService
	)

	BeforeEach(func() {
		researcherService = &ResearcherService{
			Enabled: ptr.To[bool](true),
		}
	})

	It("global.replicaCount is set to default Replicas should be 1", func() {
		researcherService.SetDefaultsWhereNeeded("", ptr.To[int32](1))
		Expect(*researcherService.Replicas).To(Equal(int32(1)))
	})

	It("global.replicaCount is nil Replicas should be 1", func() {
		researcherService.SetDefaultsWhereNeeded("", nil)
		Expect(*researcherService.Replicas).To(Equal(int32(1)))
	})

	It("global.replicaCount is set, cluster.replicas is not set be equal to global.replicaCount", func() {
		researcherService.SetDefaultsWhereNeeded("", ptr.To[int32](2))
		Expect(*researcherService.Replicas).To(Equal(int32(2)))
	})

	It("global.replicaCount is set, cluster.replicas is set should bot change", func() {
		researcherService.Replicas = ptr.To[int32](2)
		researcherService.SetDefaultsWhereNeeded("", ptr.To[int32](3))
		Expect(*researcherService.Replicas).To(Equal(int32(2)))
	})

})
