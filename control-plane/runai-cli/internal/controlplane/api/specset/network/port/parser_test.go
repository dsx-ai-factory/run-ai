/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package port

import (
	"fmt"
	"testing"

	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	NewFullExample    = "service-type=NodePort,container=1234,external=4321"
	NewMinimalExample = "service-type=NodePort,container=1234"
	InvalidExample    = "bla"

	ErrorMissingMandatory = "cannot create object, missing mandatory fields [container service-type], in value: %s"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "New Port Parser Validation Tests")
}

var _ = Describe("Parser Tests", func() {
	Context("Sanity validation", func() {
		It("Set everything", func() {
			ap := NewParser(NewFullExample)
			err := ap.Parse()
			Expect(err).To(BeNil())
			Expect(*ap.Get(ServiceType)).To(Equal("NodePort"))
			Expect(*ap.Get(Container)).To(Equal("1234"))
			Expect(*ap.Get(External)).To(Equal("4321"))
		})
	})

	Context("New Port validation", func() {
		It("Set everything", func() {
			port, err := NewPort(NewFullExample)
			Expect(err).To(BeNil())
			Expect(port.GetServiceType()).To(Equal(openapi.PORTSERVICETYPE_NODE_PORT))
			Expect(port.GetContainer()).To(Equal(int32(1234)))
			Expect(port.GetExternal()).To(Equal(int32(4321)))
		})

		It("Set minimal", func() {
			port, err := NewPort(NewMinimalExample)
			Expect(err).To(BeNil())
			Expect(port.GetServiceType()).To(Equal(openapi.PORTSERVICETYPE_NODE_PORT))
			Expect(port.GetContainer()).To(Equal(int32(1234)))
			Expect(port.HasExternal()).To(Equal(false))
		})
	})

	Context("Invalid tests", func() {
		It("Invalid string", func() {
			_, err := NewPort(InvalidExample)
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(fmt.Errorf(ErrorMissingMandatory, InvalidExample)))
		})
	})
})
