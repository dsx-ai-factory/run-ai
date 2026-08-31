/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package secretvolume

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	FirstSecretName  = "SecretName"
	SecondSecretName = "SecondSecretNameSecondName"
	FirstPath        = "MountPath"
	SecondPath       = "SecondPath"
	TestDefaultMode  = "0666"
)

var (
	FirstFullExample  = fmt.Sprintf("path=%s,name=%s,default-mode=%s", FirstPath, FirstSecretName, TestDefaultMode)
	SecondFullExample = fmt.Sprintf("path=%s,name=%s", SecondPath, SecondSecretName)
	InvalidParameter  = fmt.Sprintf("path=%s,secret=%s", FirstPath, FirstSecretName)
	MissingParameter  = fmt.Sprintf("path=%s", FirstPath)
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SecretName Volume Parser Validation Tests")
}

var _ = Describe("Parser Tests", func() {
	Context("Sanity validation", func() {
		It("Set everything", func() {
			ap := NewParser(FirstFullExample)
			err := ap.Parse()
			Expect(err).To(BeNil())
			Expect(*ap.Get(SecretName)).To(Equal("SecretName"))
			Expect(*ap.Get(MountPath)).To(Equal("MountPath"))
			Expect(*ap.Get(DefaultMode)).To(Equal(TestDefaultMode))
		})
	})

	Context("New SecretName Volume validation", func() {
		It("Set everything", func() {
			sv, err := New(FirstFullExample)
			Expect(err).To(BeNil())
			Expect(sv.GetSecret()).To(Equal("SecretName"))
			Expect(sv.GetMountPath()).To(Equal("MountPath"))
			Expect(sv.GetDefaultMode()).To(Equal(TestDefaultMode))
		})
		It("Invalid parameter", func() {
			sv, err := New(InvalidParameter)
			Expect(err.Error()).To(HaveSuffix("parser (SecretVolumeParser), dictionary does not contain key: secret, value: %s", InvalidParameter))
			Expect(sv).To(BeNil())
		})
		It("Missing parameter", func() {
			sv, err := New(MissingParameter)
			Expect(err.Error()).To(HaveSuffix("cannot create object, missing mandatory fields [name], in value: %s", MissingParameter))
			Expect(sv).To(BeNil())
		})
		Context("Multiple parameters ", func() {
			It("Valid Parameters", func() {
				multipleParams := []string{FirstFullExample, SecondFullExample}
				svList, err := NewList(multipleParams)
				Expect(err).To(BeNil())
				for _, sv := range svList {
					Expect(sv.GetSecret() == FirstSecretName || sv.GetSecret() == SecondSecretName)
					Expect(sv.GetMountPath() == FirstPath || sv.GetSecret() == SecondPath)
				}
			})
			It("Invalid Parameter", func() {
				multipleParams := []string{FirstFullExample, InvalidParameter}
				svList, err := NewList(multipleParams)
				Expect(err.Error()).To(HaveSuffix("parser (SecretVolumeParser), dictionary does not contain key: secret, value: %s", InvalidParameter))
				Expect(svList).To(BeNil())
			})
			It("Missing field in one param", func() {
				multipleParams := []string{FirstFullExample, MissingParameter}
				svList, err := NewList(multipleParams)
				Expect(err.Error()).To(HaveSuffix("cannot create object, missing mandatory fields [name], in value: %s", MissingParameter))
				Expect(svList).To(BeNil())
			})
		})

	})
})
