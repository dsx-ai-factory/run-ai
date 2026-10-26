/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("sensitive environment variables", func() {
	It("redacts literal token values from a response-only copy", func() {
		hfToken := workloadsApi.EnvironmentVariable{}
		hfToken.SetName("HF_TOKEN")
		hfToken.SetValue("synthetic-hf-marker")
		normal := workloadsApi.EnvironmentVariable{}
		normal.SetName("NORMAL")
		normal.SetValue("visible")
		stored := []workloadsApi.EnvironmentVariable{hfToken, normal}

		redacted := redactPlaintextTokenValues(stored)

		Expect(redacted).To(HaveLen(2))
		Expect(redacted[0].HasValue()).To(BeFalse())
		Expect(redacted[1].GetValue()).To(Equal("visible"))
		Expect(stored[0].GetValue()).To(Equal("synthetic-hf-marker"))
	})

	It("restores a name-only sensitive entry during update", func() {
		update := map[string]interface{}{
			"environmentVariables": []interface{}{
				map[string]interface{}{"name": "HF_TOKEN"},
				map[string]interface{}{"name": "NORMAL", "value": "changed"},
			},
		}
		stored := map[string]interface{}{
			"environmentVariables": []interface{}{
				map[string]interface{}{"name": "HF_TOKEN", "value": "synthetic-hf-marker"},
				map[string]interface{}{"name": "NORMAL", "value": "old"},
			},
		}

		err := restoreRedactedSensitiveEnvironmentVariableValues(update, stored)

		Expect(err).NotTo(HaveOccurred())
		variables := update["environmentVariables"].([]interface{})
		Expect(variables[0].(map[string]interface{})["value"]).To(Equal("synthetic-hf-marker"))
		Expect(variables[1].(map[string]interface{})["value"]).To(Equal("changed"))
	})

	It("restores a redacted sensitive entry that carries exclude false", func() {
		update := map[string]interface{}{
			"environmentVariables": []interface{}{
				map[string]interface{}{"name": "HF_TOKEN", "exclude": false},
			},
		}
		stored := map[string]interface{}{
			"environmentVariables": []interface{}{
				map[string]interface{}{"name": "HF_TOKEN", "value": "synthetic-hf-marker"},
			},
		}

		Expect(restoreRedactedSensitiveEnvironmentVariableValues(update, stored)).To(Succeed())

		variables := update["environmentVariables"].([]interface{})
		Expect(variables[0].(map[string]interface{})["value"]).To(Equal("synthetic-hf-marker"))
	})

	It("rejects a name-only sensitive entry without a stored value", func() {
		update := map[string]interface{}{
			"environmentVariables": []interface{}{map[string]interface{}{"name": "NGC_API_KEY"}},
		}

		err := restoreRedactedSensitiveEnvironmentVariableValues(update, nil)

		Expect(err).To(MatchError("NGC_API_KEY requires a value or credential reference"))
	})
})
