// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package chart_inspection

import (
	"encoding/json"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Inspection response diagnostics", func() {
	It("api.AiApplicationChartInspectionWarning JSON Always Includes Path", func() {
		encoded, err := json.Marshal(api.AiApplicationChartInspectionWarning{Code: "EXAMPLE", Message: "Example"})
		Expect(err).NotTo(HaveOccurred())

		var got map[string]any
		{
			err := json.Unmarshal(encoded, &got)
			Expect(err).NotTo(HaveOccurred())
		}
		{
			_, exists := got["path"]
			Expect(exists).To(BeTrue(), func() string {
				return fmt.Sprintf("warning path omitted: %s", encoded)
			})
		}
	})
})
