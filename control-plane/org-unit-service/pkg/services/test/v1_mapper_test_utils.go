/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	. "github.com/onsi/gomega" //nolint:staticcheck // dot-import is the idiomatic gomega usage in test helpers
	"github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"
)

func AssertV1ApiResources(resources []api.NodePoolAssignedResourcesV1Response, oqFlag bool, cpuFlag bool, overrideDeserved bool) {
	Expect(len(resources)).To(Equal(2))
	defaultResources := resources[0]
	validateV1ApiNodepoolResources(defaultResources, oqFlag, cpuFlag, float32(deserved), float32(limit), float32(oqWeight), overrideDeserved)
	poolResources := resources[1]
	validateV1ApiNodepoolResources(poolResources, oqFlag, cpuFlag, float32(deserved2), float32(limit2), float32(oqWeight2), overrideDeserved)
}

func validateV1ApiNodepoolResources(resources api.NodePoolAssignedResourcesV1Response, oqFlag bool, cpuFlag bool, deserved float32, limit float32, oqWeight float32, overrideDeserved bool) {
	if overrideDeserved {
		projectsDeserved := float32(0)
		Expect(*resources.Gpu.Deserved).To(Equal(projectsDeserved))
		Expect(*resources.Gpu.MaxAllowed).To(Equal(limit))
		if oqFlag {
			Expect(*resources.Gpu.OverQuotaWeight).To(Equal(oqWeight))
		} else {
			Expect(*resources.Gpu.OverQuotaWeight).To(Equal(projectsDeserved))
		}
	} else {
		Expect(*resources.Gpu.Deserved).To(Equal(deserved))
		Expect(*resources.Gpu.MaxAllowed).To(Equal(limit))
		if oqFlag {
			Expect(*resources.Gpu.OverQuotaWeight).To(Equal(oqWeight))
		} else {
			Expect(*resources.Gpu.OverQuotaWeight).To(Equal(deserved))
		}
	}
	if cpuFlag {
		Expect(*resources.Cpu.Deserved).To(Equal(deserved))
		Expect(*resources.Cpu.MaxAllowed).To(Equal(limit))
		Expect(*resources.Memory.Deserved).To(Equal(deserved))
		Expect(*resources.Memory.MaxAllowed).To(Equal(limit))
		if oqFlag {
			Expect(*resources.Cpu.OverQuotaWeight).To(Equal(oqWeight))
			Expect(*resources.Memory.OverQuotaWeight).To(Equal(oqWeight))
		} else {
			Expect(*resources.Cpu.OverQuotaWeight).To(Equal(deserved))
			Expect(*resources.Memory.OverQuotaWeight).To(Equal(deserved))
		}
	} else {
		Expect(resources.Cpu.Deserved).To(BeNil())
		Expect(resources.Cpu.MaxAllowed).To(BeNil())
		Expect(resources.Cpu.OverQuotaWeight).To(BeNil())
		Expect(resources.Memory.Deserved).To(BeNil())
		Expect(resources.Memory.MaxAllowed).To(BeNil())
		Expect(resources.Memory.OverQuotaWeight).To(BeNil())
	}
}

func AssertV1ApiResourcesProj(resources []api.NodePoolAssignedResourcesV1Response, oqFlag bool, cpuFlag bool, overrideDeserved bool) {
	Expect(len(resources)).To(Equal(2))
	defaultResources := resources[0]
	validateV1ApiNodepoolResourcesProj(defaultResources, oqFlag, cpuFlag, float32(deserved), float32(limit), float32(oqWeight), overrideDeserved)
	poolResources := resources[1]
	validateV1ApiNodepoolResourcesProj(poolResources, oqFlag, cpuFlag, float32(deserved2), float32(limit2), float32(oqWeight2), overrideDeserved)
}

func validateV1ApiNodepoolResourcesProj(resources api.NodePoolAssignedResourcesV1Response, oqFlag bool, cpuFlag bool, deserved float32, limit float32, oqWeight float32, overrideDeserved bool) {
	if overrideDeserved {
		projectsDeserved := float32(0)
		Expect(resources.Gpu.Deserved).To(Equal(projectsDeserved))
		Expect(*resources.Gpu.MaxAllowed).To(Equal(limit))
		if oqFlag {
			Expect(*resources.Gpu.OverQuotaWeight).To(Equal(oqWeight))
		} else {
			Expect(*resources.Gpu.OverQuotaWeight).To(Equal(projectsDeserved))
		}
	} else {
		Expect(*resources.Gpu.Deserved).To(Equal(deserved))
		Expect(*resources.Gpu.MaxAllowed).To(Equal(limit))
		if oqFlag {
			Expect(*resources.Gpu.OverQuotaWeight).To(Equal(oqWeight))
		} else {
			Expect(*resources.Gpu.OverQuotaWeight).To(Equal(deserved))
		}
	}
	if cpuFlag {
		Expect(*resources.Cpu.Deserved).To(Equal(deserved))
		Expect(*resources.Cpu.MaxAllowed).To(Equal(limit))
		Expect(*resources.Memory.Deserved).To(Equal(deserved))
		Expect(*resources.Memory.MaxAllowed).To(Equal(limit))
		if oqFlag {
			Expect(*resources.Cpu.OverQuotaWeight).To(Equal(oqWeight))
			Expect(*resources.Memory.OverQuotaWeight).To(Equal(oqWeight))
		} else {
			Expect(*resources.Cpu.OverQuotaWeight).To(Equal(deserved))
			Expect(*resources.Memory.OverQuotaWeight).To(Equal(deserved))
		}
	} else {
		Expect(resources.Cpu.Deserved).To(BeNil())
		Expect(resources.Cpu.MaxAllowed).To(BeNil())
		Expect(resources.Cpu.OverQuotaWeight).To(BeNil())
		Expect(resources.Memory.Deserved).To(BeNil())
		Expect(resources.Memory.MaxAllowed).To(BeNil())
		Expect(resources.Memory.OverQuotaWeight).To(BeNil())
	}
}

func validateV1ApiNodepoolResourcesUnlimitedProj(resources api.NodePoolAssignedResourcesV1Response, oqFlag bool, cpuFlag bool, deserved float32, limit float32, oqWeight float32, overrideDeserved bool) {
	if overrideDeserved {
		projectsDeserved := float32(0)
		Expect(resources.Gpu.Deserved).To(Equal(projectsDeserved))
		Expect(*resources.Gpu.MaxAllowed).To(Equal(limit))
		if oqFlag {
			Expect(*resources.Gpu.OverQuotaWeight).To(Equal(oqWeight))
		} else {
			Expect(*resources.Gpu.OverQuotaWeight).To(Equal(projectsDeserved))
		}
	} else {
		Expect(*resources.Gpu.Deserved).To(Equal(deserved))
		Expect(*resources.Gpu.MaxAllowed).To(Equal(limit))
		if oqFlag {
			Expect(*resources.Gpu.OverQuotaWeight).To(Equal(oqWeight))
		} else {
			Expect(*resources.Gpu.OverQuotaWeight).To(Equal(deserved))
		}
	}
	Expect(*resources.Cpu.Deserved).To(Equal(deserved))
	Expect(*resources.Cpu.MaxAllowed).To(Equal(float32(-1)))
	Expect(*resources.Memory.Deserved).To(Equal(deserved))
	Expect(*resources.Memory.MaxAllowed).To(Equal(float32(-1)))
	if oqFlag {
		Expect(*resources.Cpu.OverQuotaWeight).To(Equal(oqWeight))
		Expect(*resources.Memory.OverQuotaWeight).To(Equal(oqWeight))
	} else {
		Expect(*resources.Cpu.OverQuotaWeight).To(Equal(deserved))
		Expect(*resources.Memory.OverQuotaWeight).To(Equal(deserved))
	}
}
