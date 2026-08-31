/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package assets

import (
	"bytes"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/submit"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ProcessRawEnvironment", func() {
	var (
		c   *submit.Container
		buf bytes.Buffer
		err error
	)

	BeforeEach(func() {
		c = &submit.Container{}
		buf.Reset()
	})

	Describe("Environment Asset Name", func() {
		It("should parse asset name (no = sign)", func() {
			c.RawEnvironment = []string{"my-environment"}
			err = c.ProcessRawEnvironment(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.EnvironmentAsset).To(Equal("my-environment"))
			Expect(c.EnvironmentVariables).To(BeEmpty())
			Expect(buf.String()).To(BeEmpty())
		})

		It("should parse env var with = sign in value", func() {
			c.RawEnvironment = []string{"env1=x=y"}
			err = c.ProcessRawEnvironment(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.EnvironmentAsset).To(BeEmpty())
			Expect(c.EnvironmentVariables).To(ContainElement("env1=x=y"))
			Expect(buf.String()).To(ContainSubstring(consts.DeprecatedEnvironmentKeyValueMsg))
		})

		It("should handle empty environment", func() {
			c.RawEnvironment = []string{}
			err = c.ProcessRawEnvironment(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.EnvironmentAsset).To(BeEmpty())
		})
	})

	Describe("Environment Variable (key=value format - deprecated)", func() {
		It("should parse key=value as environment variable and print deprecation", func() {
			c.RawEnvironment = []string{"MY_VAR=my_value"}
			err = c.ProcessRawEnvironment(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.EnvironmentVariables).To(ContainElement("MY_VAR=my_value"))
			Expect(c.EnvironmentAsset).To(BeEmpty())
			Expect(buf.String()).To(ContainSubstring(consts.DeprecatedEnvironmentKeyValueMsg))
		})

		It("should parse multiple env vars", func() {
			c.RawEnvironment = []string{"VAR1=val1", "VAR2=val2"}
			err = c.ProcessRawEnvironment(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.EnvironmentVariables).To(ContainElements("VAR1=val1", "VAR2=val2"))
			Expect(buf.String()).To(ContainSubstring(consts.DeprecatedEnvironmentKeyValueMsg))
		})

		It("should append to existing environment variables", func() {
			c.EnvironmentVariables = []string{"EXISTING=value"}
			c.RawEnvironment = []string{"NEW_VAR=new_value"}
			err = c.ProcessRawEnvironment(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.EnvironmentVariables).To(ContainElements("EXISTING=value", "NEW_VAR=new_value"))
		})
	})

	Describe("Mixed usage", func() {
		It("should handle both env vars and asset name", func() {
			c.RawEnvironment = []string{"MY_VAR=value", "my-environment"}
			err = c.ProcessRawEnvironment(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.EnvironmentVariables).To(ContainElement("MY_VAR=value"))
			Expect(c.EnvironmentAsset).To(Equal("my-environment"))
			Expect(buf.String()).To(ContainSubstring(consts.DeprecatedEnvironmentKeyValueMsg))
		})
	})
})
