/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pvc

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	NewFullExample    = "claimname=CLAIM_NAME,storageclass=STORAGE_CLASS,size=SIZE,path=PATH,accessmode-rwo,accessmode-rom,accessmode-rwm,ro,ephemeral"
	NewMinimalExample = "claimname=MINIMAL_NAME,size=SIZE_MINIMAL,path=PATH_MINIMAL"

	ExistingExample            = "claimname=EXIST_NAME,path=PATH_EXIST"
	BadExistingExampleWithSize = "claimname=EXIST_NAME,path=PATH_EXIST,size=SIZE_EXIST"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PVC Parser Validation Tests")
}

var _ = Describe("Parser Tests", func() {
	Context("Sanity validation", func() {
		It("Set everything", func() {
			ap := NewParser(NewFullExample)
			err := ap.Parse()
			Expect(err).To(BeNil())
			Expect(*ap.Get(ClaimName)).To(Equal("CLAIM_NAME"))
			Expect(*ap.Get(StorageClass)).To(Equal("STORAGE_CLASS"))
			Expect(*ap.Get(Size)).To(Equal("SIZE"))
			Expect(*ap.Get(Path)).To(Equal("PATH"))
			Expect(*ap.Get(AccessModeRwo)).To(Equal(""))
			Expect(*ap.Get(AccessModeRwm)).To(Equal(""))
			Expect(*ap.Get(AccessModeRom)).To(Equal(""))
			Expect(*ap.Get(ReadOnly)).To(Equal(""))
			Expect(*ap.Get(Ephemeral)).To(Equal(""))
		})
	})

	Context("New PVC validation", func() {
		It("Set everything", func() {
			pvc, err := NewPvc(NewFullExample)
			Expect(err).To(BeNil())
			Expect(pvc.GetClaimName()).To(Equal("CLAIM_NAME"))
			Expect(pvc.GetPath()).To(Equal("PATH"))
			Expect(pvc.GetEphemeral()).To(Equal(true))
			Expect(pvc.GetExistingPvc()).To(Equal(false))
			Expect(pvc.GetReadOnly()).To(Equal(true))

			ci := pvc.GetClaimInfo()
			Expect(ci).ToNot(BeNil())
			Expect(ci.GetStorageClass()).To(Equal("STORAGE_CLASS"))
			Expect(ci.GetSize()).To(Equal("SIZE"))

			am := ci.GetAccessModes()
			Expect(am.GetReadWriteMany()).To(Equal(true))
			Expect(am.GetReadWriteOnce()).To(Equal(true))
			Expect(am.GetReadOnlyMany()).To(Equal(true))
		})

		It("Set minimal", func() {
			pvc, err := NewPvc(NewMinimalExample)
			Expect(err).To(BeNil())
			Expect(pvc.GetClaimName()).To(Equal("MINIMAL_NAME"))
			Expect(pvc.GetPath()).To(Equal("PATH_MINIMAL"))
			Expect(pvc.GetEphemeral()).To(Equal(false))
			Expect(pvc.GetExistingPvc()).To(Equal(false))
			Expect(pvc.GetReadOnly()).To(Equal(false))

			ci := pvc.GetClaimInfo()
			Expect(ci).ToNot(BeNil())
			Expect(ci.HasStorageClass()).To(Equal(false))
			Expect(ci.GetSize()).To(Equal("SIZE_MINIMAL"))
			Expect(ci.HasAccessModes()).To(Equal(false))
		})
	})

	Context("Existing PVC validation", func() {
		It("Exist success", func() {
			existPvc, err := ExistingPvc(ExistingExample)
			Expect(err).To(BeNil())
			Expect(existPvc.GetClaimName()).To(Equal("EXIST_NAME"))
			Expect(existPvc.GetPath()).To(Equal("PATH_EXIST"))
		})

		It("Exist failed - too many attributes", func() {
			existPvc, err := ExistingPvc(BadExistingExampleWithSize)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(HaveSuffix("dictionary does not contain key: size, value: %s", BadExistingExampleWithSize))
			Expect(existPvc).To(BeNil())
		})
	})
})
