/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cli_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cli"
	openapi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
)

var _ = Describe("NullableResourcesFlag", func() {
	var (
		nr *cli.NullableResourcesFlag
	)

	BeforeEach(func() {
		nr = cli.NewNullableResourcesFlag(&[]openapi.ResourcesNullable{})
	})

	Describe("Set", func() {
		It("should set a valid resources value with default values for gpu-deserved,gpu-limit", func() {
			err := nr.Set("nodepool=some-nodepool")
			Expect(err).NotTo(HaveOccurred())
			Expect(nr.IsSet()).To(BeTrue())
			Expect(nr.String()).To(Equal("nodepool=some-nodepool,gpu-deserved=0.00,gpu-limit=-1.00"))
		})

		It("should fail on invalid prop name", func() {
			err := nr.Set("node=some-nodepool")
			Expect(err).To(HaveOccurred())
			Expect(nr.IsSet()).To(BeFalse())
		})
	})

	Describe("Set with gpu values", func() {
		It("should set a valid resources value with default values for gpu-deserved,gpu-limit", func() {
			err := nr.Set("nodepool=some-nodepool,gpu-deserved=2.00,gpu-limit=2.1")
			Expect(err).NotTo(HaveOccurred())
			Expect(nr.IsSet()).To(BeTrue())
			Expect(nr.String()).To(Equal("nodepool=some-nodepool,gpu-deserved=2.00,gpu-limit=2.10"))
		})

		It("should set default gpu-limit on gpu-limit not explicitly set", func() {
			err := nr.Set("nodepool=some-nodepool,gpu-deserved=3.00")
			Expect(err).NotTo(HaveOccurred())
			Expect(nr.IsSet()).To(BeTrue())
			Expect(nr.String()).To(Equal("nodepool=some-nodepool,gpu-deserved=3.00,gpu-limit=-1.00"))
		})

		It("should fail on gpu-deserved is higher than gpu-limit", func() {
			err := nr.Set("nodepool=some-nodepool,gpu-deserved=3.00,gpu-limit=2.1")
			Expect(err).To(HaveOccurred())
			Expect(nr.IsSet()).To(BeFalse())
			Expect(nr.String()).To(Equal(""))
		})
	})
})
