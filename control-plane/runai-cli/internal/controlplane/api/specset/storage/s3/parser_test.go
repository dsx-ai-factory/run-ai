/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package s3

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	NewFullExample    = "name=NAME,bucket=BUCKET,path=PATH,accesskey=ACCESS_KEY,url=URL"
	NewMinimalExample = "name=NAME,bucket=BUCKET,path=PATH"
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
			Expect(*ap.Get(Name)).To(Equal("NAME"))
			Expect(*ap.Get(Bucket)).To(Equal("BUCKET"))
			Expect(*ap.Get(Path)).To(Equal("PATH"))
			Expect(*ap.Get(AccessKey)).To(Equal("ACCESS_KEY"))
			Expect(*ap.Get(Url)).To(Equal("URL"))
		})
	})

	Context("New Host Path validation", func() {
		It("Set everything", func() {
			s3, err := NewS3(NewFullExample)
			Expect(err).To(BeNil())
			Expect(s3.GetName()).To(Equal("NAME"))
			Expect(s3.GetBucket()).To(Equal("BUCKET"))
			Expect(s3.GetPath()).To(Equal("PATH"))
			Expect(s3.GetAccessKeySecret()).To(Equal("ACCESS_KEY"))
			Expect(s3.GetUrl()).To(Equal("URL"))
		})

		It("Set minimal", func() {
			s3, err := NewS3(NewMinimalExample)
			Expect(err).To(BeNil())
			Expect(s3.GetName()).To(Equal("NAME"))
			Expect(s3.GetBucket()).To(Equal("BUCKET"))
			Expect(s3.GetPath()).To(Equal("PATH"))
			Expect(s3.HasUrl()).To(Equal(false))
			aks := s3.AccessKeySecret
			Expect(aks.IsSet()).To(Equal(false))
		})
	})
})
