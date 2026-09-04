/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package usage_info

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test usage info logic", func() {
	Context("no patching info", func() {
		It("should return with no errors", func() {
			err := PatchUsageInfo(context.Background(), api.ASSETKIND_ENVIRONMENT, nil, nil)
			Expect(err).To(BeNil())
		})
	})
})
