/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("legacy workload sensitive environment variables", func() {
	It("retains sensitive map keys with empty literal values", func() {
		environmentVariables := models.EnvironmentVariables{
			"HF_TOKEN":    "synthetic-hf-value",
			"NGC_API_KEY": "synthetic-ngc-value",
			"NORMAL_VAR":  "visible",
		}

		redactLegacyPlaintextTokenValues(environmentVariables)

		Expect(environmentVariables).To(HaveKeyWithValue("HF_TOKEN", ""))
		Expect(environmentVariables).To(HaveKeyWithValue("NGC_API_KEY", ""))
		Expect(environmentVariables).To(HaveKeyWithValue("NORMAL_VAR", "visible"))
	})

	It("does not add absent sensitive keys", func() {
		environmentVariables := models.EnvironmentVariables{"NORMAL_VAR": "visible"}

		redactLegacyPlaintextTokenValues(environmentVariables)

		Expect(environmentVariables).NotTo(HaveKey("HF_TOKEN"))
		Expect(environmentVariables).NotTo(HaveKey("NGC_API_KEY"))
	})

	It("leaves credential references unchanged", func() {
		environmentVariables := models.EnvironmentVariables{
			"HF_TOKEN": "USERCREDENTIAL:my-hf-credential",
		}

		redactLegacyPlaintextTokenValues(environmentVariables)

		Expect(environmentVariables).To(HaveKeyWithValue("HF_TOKEN", "USERCREDENTIAL:my-hf-credential"))
	})
})
