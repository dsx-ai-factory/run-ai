/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package upgrade

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/os"
)

var _ = Describe("Factory", func() {
	var factory *Factory

	BeforeEach(func() {
		factory = &Factory{
			unixFlow:    &MockUpgrader1{},
			windowsFlow: &MockUpgrader{},
		}
	})

	Describe("CreateFlow", func() {
		Context("with Windows OS", func() {
			BeforeEach(func() {

			})

			It("returns a Windows upgrader", func() {
				upgrader, err := factory.CreateFlow(os.Windows)
				Expect(err).NotTo(HaveOccurred())
				Expect(upgrader).To(BeAssignableToTypeOf(&MockUpgrader{}))
			})
		})

		Context("with linux OS", func() {
			It("returns a Unix upgrader", func() {
				upgrader, err := factory.CreateFlow(os.Linux)
				Expect(err).NotTo(HaveOccurred())
				Expect(upgrader).To(BeAssignableToTypeOf(&MockUpgrader1{}))
			})
		})
		Context("with darwin OS", func() {
			It("returns a Unix upgrader", func() {
				upgrader, err := factory.CreateFlow(os.Darwin)
				Expect(err).NotTo(HaveOccurred())
				Expect(upgrader).To(BeAssignableToTypeOf(&MockUpgrader1{}))
			})
		})

		Context("with unsupported OS", func() {
			It("returns an error", func() {
				upgrader, err := factory.CreateFlow(os.Unknown)
				Expect(err).To(HaveOccurred())
				Expect(upgrader).To(BeNil())
			})
		})
	})

})

type MockUpgrader struct{}

func (m *MockUpgrader) Upgrade(ctx context.Context, execShell string) (string, error) {
	return "", nil
}

type MockUpgrader1 struct{}

func (m *MockUpgrader1) Upgrade(ctx context.Context, execShell string) (string, error) {
	return "", nil
}
