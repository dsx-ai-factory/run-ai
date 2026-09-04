/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"k8s.io/utils/ptr"
)

var _ = Describe("BriefListFromResponse", func() {
	var apps []openapi.AIApplication

	BeforeEach(func() {
		statusStr := "Running"
		apps = []openapi.AIApplication{
			{
				Id:          ptr.To("app-1"),
				Name:        ptr.To("my-app"),
				Type:        ptr.To("inference"),
				ProjectName: ptr.To("team-alpha"),
			},
		}
		apps[0].ActionStatus.Set(&statusStr)
	})

	It("should convert app list to brief list", func() {
		result := BriefListFromResponse(apps)

		Expect(result).To(HaveLen(1))
		Expect(result[0].Name).To(Equal("my-app"))
		Expect(result[0].Type).To(Equal("inference"))
		Expect(result[0].ActionStatus).To(Equal("Running"))
		Expect(result[0].ProjectName).To(Equal("team-alpha"))
	})

	It("should handle empty app list", func() {
		result := BriefListFromResponse([]openapi.AIApplication{})

		Expect(result).To(HaveLen(0))
	})

	It("should handle app with missing optional fields", func() {
		apps = []openapi.AIApplication{
			{Id: ptr.To("app-2"), Name: ptr.To("bare-app")},
		}

		result := BriefListFromResponse(apps)

		Expect(result).To(HaveLen(1))
		Expect(result[0].Name).To(Equal("bare-app"))
		Expect(result[0].Type).To(Equal(""))
		Expect(result[0].ActionStatus).To(Equal(""))
		Expect(result[0].ProjectName).To(Equal(""))
	})

	It("should convert multiple apps", func() {
		apps = []openapi.AIApplication{
			{Id: ptr.To("app-1"), Name: ptr.To("app-one")},
			{Id: ptr.To("app-2"), Name: ptr.To("app-two")},
			{Id: ptr.To("app-3"), Name: ptr.To("app-three")},
		}

		result := BriefListFromResponse(apps)

		Expect(result).To(HaveLen(3))
		Expect(result[0].Name).To(Equal("app-one"))
		Expect(result[1].Name).To(Equal("app-two"))
		Expect(result[2].Name).To(Equal("app-three"))
	})
})
