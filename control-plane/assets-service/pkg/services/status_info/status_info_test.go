/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package status_info

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test status info logic", func() {
	Context("no patching info", func() {
		It("should return with no errors", func() {
			err := PatchStatusInfo(context.Background(), nil, nil, "")
			Expect(err).To(BeNil())
		})
	})
})
