/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package extensions_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/extensions"
)

var _ = Describe("Group.Equals", func() {
	var groupA, groupB extensions.Group

	BeforeEach(func() {
		// Reset groups before each test
		groupA = extensions.Group{}
		groupB = extensions.Group{}
	})

	genericExtension := extensions.Extension{
		Status: extensions.Status{
			Enabled:      true,
			Capabilities: map[extensions.CapabilityType]bool{"cap1": true},
		},
	}
	Describe("when groups are equal", func() {
		It("should return true if both groups are empty", func() {
			Expect(groupA.Equals(groupB)).To(BeTrue())
		})

		It("should return true if groups have identical content", func() {
			groupA["Type1"] = genericExtension
			groupB["Type1"] = genericExtension

			Expect(groupA.Equals(groupB)).To(BeTrue())
		})
		It("should return true if capabilities are structurally equal but map ordering differs", func() {
			groupA["Type1"] = extensions.Extension{
				Status: extensions.Status{
					Enabled:      true,
					Capabilities: map[extensions.CapabilityType]bool{"cap1": true, "cap2": false},
				},
			}
			groupB["Type1"] = extensions.Extension{
				Status: extensions.Status{
					Enabled:      true,
					Capabilities: map[extensions.CapabilityType]bool{"cap2": false, "cap1": true},
				},
			}

			Expect(groupA.Equals(groupB)).To(BeTrue())
		})

	})

	Describe("when groups are not equal", func() {
		It("should return false if one group is empty and the other is not", func() {
			groupA["Type1"] = genericExtension

			Expect(groupA.Equals(groupB)).To(BeFalse())
		})

		It("should return false if groups have different lengths", func() {
			groupA["Type1"] = genericExtension
			groupA["Type2"] = extensions.Extension{
				Status: extensions.Status{
					Enabled:      false,
					Capabilities: map[extensions.CapabilityType]bool{"cap2": false},
				},
			}
			groupB["Type1"] = genericExtension

			Expect(groupA.Equals(groupB)).To(BeFalse())
		})

		It("should return false if a type exists in one group but not in the other", func() {
			groupA["Type1"] = genericExtension
			groupB["Type2"] = genericExtension

			Expect(groupA.Equals(groupB)).To(BeFalse())
		})

		It("should return false if extension status differs between groups", func() {
			groupA["Type1"] = genericExtension
			groupB["Type1"] = extensions.Extension{
				Status: extensions.Status{
					Enabled:      false,
					Capabilities: map[extensions.CapabilityType]bool{"cap1": true},
				},
			}

			Expect(groupA.Equals(groupB)).To(BeFalse())
		})

		It("should return false if capabilities differ between extensions in the groups", func() {
			groupA["Type1"] = genericExtension
			groupB["Type1"] = extensions.Extension{
				Status: extensions.Status{
					Enabled:      true,
					Capabilities: map[extensions.CapabilityType]bool{"cap2": false},
				},
			}

			Expect(groupA.Equals(groupB)).To(BeFalse())
		})

		It("should return false if capabilities differ in keys but both are true", func() {
			groupA["Type1"] = genericExtension
			groupB["Type1"] = extensions.Extension{
				Status: extensions.Status{
					Enabled:      true,
					Capabilities: map[extensions.CapabilityType]bool{"cap2": true},
				},
			}
			Expect(groupA.Equals(groupB)).To(BeFalse())
		})
		It("should return false if extensions are enabled but different capbilites", func() {
			groupA["Type1"] = extensions.Extension{
				Status: extensions.Status{
					Enabled:      true,
					Capabilities: map[extensions.CapabilityType]bool{"cap1": true, "cap2": false},
				},
			}
			groupB["Type1"] = extensions.Extension{
				Status: extensions.Status{
					Enabled:      true,
					Capabilities: map[extensions.CapabilityType]bool{"cap2": false, "cap3": true},
				},
			}
			Expect(groupA.Equals(groupB)).To(BeFalse())
		})

	})
})
