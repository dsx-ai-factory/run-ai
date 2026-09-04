/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
)

var _ = Describe("template related validations", func() {
	Context("get tool type -> external tool", func() {
		It("should return the type of the external tool", func() {
			connection := api.Connection{
				IsExternal: true,
				ExternalToolInfo: &api.ExternalToolInfo{
					ToolType: api.EXTERNALTOOLTYPE_WANDB,
				},
			}
			Expect(getToolType(connection)).To(Equal(api.TOOLTYPE_WANDB))
		})
	})

	Context("get tool type -> external tool -> comet", func() {
		It("should return the type of the external tool", func() {
			connection := api.Connection{
				IsExternal: true,
				ExternalToolInfo: &api.ExternalToolInfo{
					ToolType: api.EXTERNALTOOLTYPE_COMET,
				},
			}
			Expect(getToolType(connection)).To(Equal(api.TOOLTYPE_COMET))
		})
	})

	Context("get tool type -> internal tool", func() {
		It("should return the type of the internal tool", func() {
			connection := api.Connection{
				IsExternal: false,
				InternalToolInfo: &api.InternalToolInfo{
					ToolType: api.INTERNALTOOLTYPE_MLFLOW,
				},
			}
			Expect(getToolType(connection)).To(Equal(api.TOOLTYPE_MLFLOW))
		})
	})

	Context("get tool type -> default", func() {
		It("should return the type custom tool", func() {
			connection := api.Connection{
				IsExternal: false,
			}
			Expect(getToolType(connection)).To(Equal(api.TOOLTYPE_CUSTOM))
		})
	})
})
