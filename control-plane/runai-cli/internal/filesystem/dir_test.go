/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package filesystem

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("IsDirWritable", func() {
	It("should return true when dir exists and writable", func() {
		tempDir, err := os.MkdirTemp("", "writable-test")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(tempDir)

		err = IsDirWritable(tempDir)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should return true when dir doesn't exist and writable", func() {
		tempDir, err := os.MkdirTemp("", "parent-dir")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(tempDir)

		nestedDir := tempDir + "/nested/doesnt/exist"
		err = IsDirWritable(nestedDir)
		Expect(err).NotTo(HaveOccurred())

		_, err = os.Stat(nestedDir)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should return false when dir is not writable", func() {
		tempDir, err := os.MkdirTemp("", "readonly-test")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(tempDir)

		var readOnlyMode os.FileMode = 0444
		err = os.Chmod(tempDir, readOnlyMode)
		Expect(err).NotTo(HaveOccurred())

		err = IsDirWritable(tempDir)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("directory is not writable"))
	})
})
