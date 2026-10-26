/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
)

var _ = Describe("mergeUpdateSpecIntoFullSpec (generic partial)", func() {
	ctx := context.Background()

	It("merges a root-level field from an untyped (any) partial into the full spec", func() {
		fullSpec := workloadsApi.SupersetSpec{}
		err := json.Unmarshal([]byte(`{"image":"image:v1","command":"cmd1"}`), &fullSpec)
		Expect(err).To(BeNil())

		// partial provided as a generic value (any) — not a typed *InferenceUpdateSpecSpec.
		// This is what lets the shared orchestrator merge any workload type's partial.
		var partial any = map[string]interface{}{"image": "image:v2"}

		mergeErr := mergeUpdateSpecIntoFullSpec(ctx, &fullSpec, partial)
		Expect(mergeErr).To(BeNil())

		// changed field overwritten
		Expect(fullSpec.Image.IsSet()).To(BeTrue())
		Expect(*fullSpec.Image.Get()).To(Equal("image:v2"))
		// untouched field preserved
		Expect(fullSpec.Command.IsSet()).To(BeTrue())
		Expect(*fullSpec.Command.Get()).To(Equal("cmd1"))
	})

	It("preserves a stored sensitive literal from a redacted update placeholder", func() {
		fullSpec := workloadsApi.SupersetSpec{}
		Expect(json.Unmarshal([]byte(`{"environmentVariables":[{"name":"HF_TOKEN","value":"synthetic-hf-marker"},{"name":"NORMAL","value":"old"}]}`), &fullSpec)).To(Succeed())
		partial := map[string]interface{}{
			"environmentVariables": []interface{}{
				map[string]interface{}{"name": "HF_TOKEN"},
				map[string]interface{}{"name": "NORMAL", "value": "changed"},
			},
		}

		Expect(mergeUpdateSpecIntoFullSpec(ctx, &fullSpec, partial)).To(BeNil())

		result, err := json.Marshal(fullSpec)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(result)).To(ContainSubstring(`"name":"HF_TOKEN","value":"synthetic-hf-marker"`))
		Expect(string(result)).To(ContainSubstring(`"name":"NORMAL","value":"changed"`))
	})

	It("allows explicit replacement, credential conversion, and deletion", func() {
		fullSpec := workloadsApi.SupersetSpec{}
		Expect(json.Unmarshal([]byte(`{"environmentVariables":[{"name":"HF_TOKEN","value":"synthetic-hf-marker"}]}`), &fullSpec)).To(Succeed())

		Expect(mergeUpdateSpecIntoFullSpec(ctx, &fullSpec, map[string]interface{}{
			"environmentVariables": []interface{}{map[string]interface{}{"name": "HF_TOKEN", "value": "replacement"}},
		})).To(BeNil())
		result, err := json.Marshal(fullSpec)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(result)).To(ContainSubstring(`"value":"replacement"`))

		Expect(mergeUpdateSpecIntoFullSpec(ctx, &fullSpec, map[string]interface{}{
			"environmentVariables": []interface{}{map[string]interface{}{
				"name":   "HF_TOKEN",
				"secret": map[string]interface{}{"name": "hf-secret", "key": "HF_TOKEN"},
			}},
		})).To(BeNil())
		result, err = json.Marshal(fullSpec)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(result)).To(ContainSubstring(`"secret":{`))
		Expect(string(result)).To(ContainSubstring(`"name":"hf-secret"`))
		Expect(string(result)).To(ContainSubstring(`"key":"HF_TOKEN"`))
		Expect(string(result)).NotTo(ContainSubstring(`"value":"replacement"`))

		Expect(mergeUpdateSpecIntoFullSpec(ctx, &fullSpec, map[string]interface{}{
			"environmentVariables": []interface{}{},
		})).To(BeNil())
		result, err = json.Marshal(fullSpec)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(result)).To(ContainSubstring(`"environmentVariables":[]`))
	})
})
