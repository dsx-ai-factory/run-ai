/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package hostpath

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	NewFullExample    = "path=PATH,mount=MOUNT,mount-propagation=HostToContainer,readwrite"
	NewMinimalExample = "path=PATH,mount=MOUNT"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Host Path Parser Validation Tests")
}

var _ = Describe("Parser Tests", func() {
	Context("Sanity validation", func() {
		It("Set everything", func() {
			ap := NewParser(NewFullExample)
			err := ap.Parse()
			Expect(err).To(BeNil())
			Expect(*ap.Get(Path)).To(Equal("PATH"))
			Expect(*ap.Get(Mount)).To(Equal("MOUNT"))
			Expect(*ap.Get(ReadWrite)).To(Equal(""))
		})
	})

	Context("New Host Path validation", func() {
		It("Set everything", func() {
			host, err := New(NewFullExample)
			Expect(err).To(BeNil())
			Expect(host.GetPath()).To(Equal("PATH"))
			Expect(host.GetMountPath()).To(Equal("MOUNT"))
			Expect(host.GetReadOnly()).To(Equal(false))
		})

		It("Set minimal", func() {
			host, err := New(NewMinimalExample)
			Expect(err).To(BeNil())
			Expect(host.GetPath()).To(Equal("PATH"))
			Expect(host.GetMountPath()).To(Equal("MOUNT"))
			Expect(host.GetReadOnly()).To(Equal(true))
		})
	})
})
