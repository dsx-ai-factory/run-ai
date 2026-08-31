/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cmv

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	NewFullExample    = "name=NAME,path=PATH,subpath=SUBPATH,default-mode=DEFAULT_MODE"
	NewPartialExample = "name=NAME,path=PATH"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "New ConfigMap Volume Parser Validation Tests")
}

var _ = Describe("Parser Tests", func() {
	Context("Sanity validation", func() {
		It("Set everything", func() {
			ap := NewParser(NewFullExample)
			err := ap.Parse()
			Expect(err).To(BeNil())
			Expect(*ap.Get(Name)).To(Equal("NAME"))
			Expect(*ap.Get(Path)).To(Equal("PATH"))
			Expect(*ap.Get(SubPath)).To(Equal("SUBPATH"))
			Expect(*ap.Get(DefaultMode)).To(Equal("DEFAULT_MODE"))
		})
	})

	Context("New ConfigMap Volume validation", func() {
		It("Set everything", func() {
			cm, err := NewConfigMapVolume(NewFullExample)
			Expect(err).To(BeNil())
			Expect(cm.GetConfigMap()).To(Equal("NAME"))
			Expect(cm.GetMountPath()).To(Equal("PATH"))
			Expect(cm.GetSubPath()).To(Equal("SUBPATH"))
			Expect(cm.GetDefaultMode()).To(Equal("DEFAULT_MODE"))
		})
		It("Set only mandatory fields", func() {
			cm, err := NewConfigMapVolume(NewPartialExample)
			Expect(err).To(BeNil())
			Expect(cm.GetConfigMap()).To(Equal("NAME"))
			Expect(cm.GetMountPath()).To(Equal("PATH"))
			Expect(cm.SubPath.Get()).To(BeNil())
			Expect(cm.DefaultMode.Get()).To(BeNil())
		})
	})
})
