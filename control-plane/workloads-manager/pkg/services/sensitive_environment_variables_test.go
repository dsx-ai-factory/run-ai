/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("sensitive environment variable response sanitization", func() {
	It("redacts nested raw and effective manifests without mutating persisted maps", func() {
		raw := map[string]interface{}{
			"metadata": map[string]interface{}{
				"labels": []interface{}{map[string]interface{}{"name": "HF_TOKEN", "value": "not-an-env-secret"}},
			},
			"spec": map[string]interface{}{
				"containers": []interface{}{map[string]interface{}{
					"env": []interface{}{
						map[string]interface{}{"name": "HF_TOKEN", "value": "synthetic-hf-value"},
						map[string]interface{}{"name": "NORMAL", "value": "visible"},
					},
				}},
			},
		}
		response := api.WorkloadV2{RawManifest: raw, WorkloadManifest: raw}

		redactWorkloadV2SensitiveValues(&response)

		redactedEnv := response.RawManifest["spec"].(map[string]interface{})["containers"].([]interface{})[0].(map[string]interface{})["env"].([]interface{})
		Expect(redactedEnv[0].(map[string]interface{})).NotTo(HaveKey("value"))
		Expect(redactedEnv[1].(map[string]interface{})["value"]).To(Equal("visible"))
		label := response.RawManifest["metadata"].(map[string]interface{})["labels"].([]interface{})[0].(map[string]interface{})
		Expect(label["value"]).To(Equal("not-an-env-secret"))
		storedEnv := raw["spec"].(map[string]interface{})["containers"].([]interface{})[0].(map[string]interface{})["env"].([]interface{})
		Expect(storedEnv[0].(map[string]interface{})["value"]).To(Equal("synthetic-hf-value"))
	})

	It("redacts NIM response literals without changing the submitted spec", func() {
		name, value := "NGC_API_KEY", "synthetic-ngc-value"
		environmentVariables := []api.EnvironmentVariable{{Name: &name, Value: &value}}
		response := api.NimServiceResponse{Spec: api.NimServiceSpec{EnvironmentVariables: &environmentVariables}}

		redactNimServiceResponseSensitiveValues(&response)

		Expect((*response.Spec.EnvironmentVariables)[0].Value).To(BeNil())
		Expect(environmentVariables[0].Value).To(Equal(&value))
	})

	It("restores a name-only token in a raw manifest update", func() {
		stored := map[string]interface{}{
			"spec": map[string]interface{}{
				"containers": []interface{}{map[string]interface{}{
					"env": []interface{}{map[string]interface{}{"name": "HF_TOKEN", "value": "synthetic-hf-value"}},
				}},
			},
		}
		update := map[string]interface{}{
			"spec": map[string]interface{}{
				"containers": []interface{}{map[string]interface{}{
					"env": []interface{}{map[string]interface{}{"name": "HF_TOKEN"}},
				}},
			},
		}

		Expect(restoreRedactedSensitiveManifestValues(update, stored)).To(Succeed())

		updatedEnv := update["spec"].(map[string]interface{})["containers"].([]interface{})[0].(map[string]interface{})["env"].([]interface{})
		Expect(updatedEnv[0].(map[string]interface{})["value"]).To(Equal("synthetic-hf-value"))
	})

	It("matches reordered containers by name before restoring token values", func() {
		stored := map[string]interface{}{
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"name": "first",
						"env":  []interface{}{map[string]interface{}{"name": "HF_TOKEN", "value": "first-token"}},
					},
					map[string]interface{}{
						"name": "second",
						"env":  []interface{}{map[string]interface{}{"name": "HF_TOKEN", "value": "second-token"}},
					},
				},
			},
		}
		update := map[string]interface{}{
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"name": "second",
						"env":  []interface{}{map[string]interface{}{"name": "HF_TOKEN"}},
					},
					map[string]interface{}{
						"name": "first",
						"env":  []interface{}{map[string]interface{}{"name": "HF_TOKEN"}},
					},
				},
			},
		}

		Expect(restoreRedactedSensitiveManifestValues(update, stored)).To(Succeed())

		containers := update["spec"].(map[string]interface{})["containers"].([]interface{})
		secondEnv := containers[0].(map[string]interface{})["env"].([]interface{})
		firstEnv := containers[1].(map[string]interface{})["env"].([]interface{})
		Expect(secondEnv[0].(map[string]interface{})["value"]).To(Equal("second-token"))
		Expect(firstEnv[0].(map[string]interface{})["value"]).To(Equal("first-token"))
	})

	It("rejects a name-only token without a stored literal", func() {
		update := map[string]interface{}{
			"spec": map[string]interface{}{
				"containers": []interface{}{map[string]interface{}{
					"env": []interface{}{map[string]interface{}{"name": "HF_TOKEN"}},
				}},
			},
		}

		Expect(restoreRedactedSensitiveManifestValues(update, nil)).To(MatchError("HF_TOKEN requires a value or credential reference"))
	})

	It("restores a name-only NIM token during update", func() {
		name, storedValue := "NGC_API_KEY", "synthetic-ngc-value"
		updateVariables := []api.EnvironmentVariable{{Name: &name}}
		storedVariables := []api.EnvironmentVariable{{Name: &name, Value: &storedValue}}
		update := api.NimServiceSpec{EnvironmentVariables: &updateVariables}

		Expect(restoreRedactedNimSensitiveValues(&update, api.NimServiceSpec{EnvironmentVariables: &storedVariables})).To(Succeed())

		Expect((*update.EnvironmentVariables)[0].Value).To(Equal(&storedValue))
	})

	It("rejects a name-only NIM token without a stored literal", func() {
		name := "NGC_API_KEY"
		updateVariables := []api.EnvironmentVariable{{Name: &name}}
		update := api.NimServiceSpec{EnvironmentVariables: &updateVariables}

		Expect(restoreRedactedNimSensitiveValues(&update, api.NimServiceSpec{})).
			To(MatchError("NGC_API_KEY requires a value or credential reference"))
	})
})
