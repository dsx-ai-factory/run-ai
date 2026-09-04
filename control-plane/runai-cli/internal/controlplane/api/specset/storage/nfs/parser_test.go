/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nfs

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	NewFullExample    = "path=PATH,server=SERVER,mountpath=MOUNT_PATH,readwrite"
	NewMinimalExample = "path=PATH,server=SERVER,mountpath=MOUNT_PATH"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "NFS Parser Validation Tests")
}

var _ = Describe("Parser Tests", func() {
	Context("Sanity validation", func() {
		It("Set everything", func() {
			ap := NewParser(NewFullExample)
			err := ap.Parse()
			Expect(err).To(BeNil())
			Expect(*ap.Get(Path)).To(Equal("PATH"))
			Expect(*ap.Get(Server)).To(Equal("SERVER"))
			Expect(*ap.Get(MountPath)).To(Equal("MOUNT_PATH"))
			Expect(*ap.Get(ReadWrite)).To(Equal(""))
		})
	})

	Context("New NFS validation", func() {
		It("Set everything", func() {
			nfs, err := NewNfs(NewFullExample)
			Expect(err).To(BeNil())
			Expect(nfs.GetPath()).To(Equal("PATH"))
			Expect(nfs.GetServer()).To(Equal("SERVER"))
			Expect(nfs.GetMountPath()).To(Equal("MOUNT_PATH"))
			Expect(nfs.GetReadOnly()).To(Equal(false))
		})

		It("Set minimal", func() {
			nfs, err := NewNfs(NewMinimalExample)
			Expect(err).To(BeNil())
			Expect(nfs.GetPath()).To(Equal("PATH"))
			Expect(nfs.GetServer()).To(Equal("SERVER"))
			Expect(nfs.GetMountPath()).To(Equal("MOUNT_PATH"))
			Expect(nfs.GetReadOnly()).To(Equal(true))
		})
	})
})
