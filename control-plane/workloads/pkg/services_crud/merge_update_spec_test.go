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

	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
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
})
