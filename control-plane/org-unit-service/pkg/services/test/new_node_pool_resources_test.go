/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	clusteropenapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

var _ = Describe("Department resources for a new node pool", func() {
	quotaMode := &clusteropenapi.SchedulingCapabilities{Quota: true, Ranks: true}
	weightsMode := &clusteropenapi.SchedulingCapabilities{AdjustableWeights: true}
	ranksMode := &clusteropenapi.SchedulingCapabilities{Ranks: true}

	DescribeTable("GPU quota and weight",
		func(capabilities *clusteropenapi.SchedulingCapabilities, unlimited bool, deserved float64, weight float64) {
			resources := services.GetDepartmentResourcesToAdd(models.OrgUnitData{}, unlimited, "1", "pool", capabilities)
			Expect(resources.GpuDeserved).To(Equal(ptr.To(deserved)))
			Expect(resources.GpuOverQuotaWeight).To(Equal(ptr.To(weight)))
		},
		Entry("unknown capabilities keep the weight derived from quota", nil, false, 0.0, 0.0),
		Entry("unknown capabilities keep the unlimited default department", nil, true, -1.0, -1.0),
		Entry("quota mode keeps the weight derived from quota", quotaMode, false, 0.0, 0.0),
		Entry("quota mode keeps the unlimited default department", quotaMode, true, -1.0, -1.0),
		Entry("weights mode gives a department the default weight", weightsMode, false, 0.0, models.DefaultOverQuotaWeight),
		Entry("weights mode gives the default department no quota", weightsMode, true, 0.0, models.DefaultOverQuotaWeight),
		Entry("ranks mode keeps the weight derived from quota", ranksMode, false, 0.0, 0.0),
		Entry("ranks mode gives the default department no quota", ranksMode, true, 0.0, 0.0),
	)
})
