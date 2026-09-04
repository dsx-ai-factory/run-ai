/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config_test

import (
	"flag"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = Describe("MapFlags (config-values parser)", func() {
	var mapFlags config.MapFlags

	BeforeEach(func() {
		mapFlags = make(config.MapFlags)
	})

	Describe("String method", func() {
		It("should return string representation of empty map", func() {
			result := mapFlags.String()
			Expect(result).To(Equal("map[]"))
		})

		It("should return string representation of populated map", func() {
			mapFlags["key1"] = "value1"
			mapFlags["key2"] = "value2"
			result := mapFlags.String()
			Expect(result).To(ContainSubstring("key1:value1"))
			Expect(result).To(ContainSubstring("key2:value2"))
		})
	})

	Describe("Set method", func() {
		Context("with key=value pairs", func() {
			It("should parse single key=value pair", func() {
				err := mapFlags.Set("key=value")
				Expect(err).ToNot(HaveOccurred())
				Expect(mapFlags).To(HaveLen(1))
				Expect(mapFlags["key"]).To(Equal("value"))
			})

			It("should parse multiple key=value pairs separated by semicolon", func() {
				err := mapFlags.Set("key1=value1;key2=value2;key3=value3")
				Expect(err).ToNot(HaveOccurred())
				Expect(mapFlags).To(HaveLen(3))
				Expect(mapFlags["key1"]).To(Equal("value1"))
				Expect(mapFlags["key2"]).To(Equal("value2"))
				Expect(mapFlags["key3"]).To(Equal("value3"))
			})

			It("should handle values with special characters", func() {
				err := mapFlags.Set("license-key=ABC-123-XYZ;url=https://example.com:8080")
				Expect(err).ToNot(HaveOccurred())
				Expect(mapFlags).To(HaveLen(2))
				Expect(mapFlags["license-key"]).To(Equal("ABC-123-XYZ"))
				Expect(mapFlags["url"]).To(Equal("https://example.com:8080"))
			})

			It("should handle values with spaces", func() {
				err := mapFlags.Set("name=John Doe;description=This is a test")
				Expect(err).ToNot(HaveOccurred())
				Expect(mapFlags).To(HaveLen(2))
				Expect(mapFlags["name"]).To(Equal("John Doe"))
				Expect(mapFlags["description"]).To(Equal("This is a test"))
			})

			It("should trim whitespace from keys and values", func() {
				err := mapFlags.Set("  key1  =  value1  ;  key2  =  value2  ")
				Expect(err).ToNot(HaveOccurred())
				Expect(mapFlags).To(HaveLen(2))
				Expect(mapFlags["key1"]).To(Equal("value1"))
				Expect(mapFlags["key2"]).To(Equal("value2"))
			})

			It("should handle empty values", func() {
				err := mapFlags.Set("key1=;key2=value2")
				Expect(err).ToNot(HaveOccurred())
				Expect(mapFlags).To(HaveLen(2))
				Expect(mapFlags["key1"]).To(Equal(""))
				Expect(mapFlags["key2"]).To(Equal("value2"))
			})

			It("should handle values containing equal signs", func() {
				err := mapFlags.Set("equation=a=b+c;query=select * from table where id=123")
				Expect(err).ToNot(HaveOccurred())
				Expect(mapFlags).To(HaveLen(2))
				Expect(mapFlags["equation"]).To(Equal("a=b+c"))
				Expect(mapFlags["query"]).To(Equal("select * from table where id=123"))
			})
		})

		Context("with boolean flags (no value)", func() {
			It("should parse single boolean flag", func() {
				err := mapFlags.Set("enable-feature")
				Expect(err).ToNot(HaveOccurred())
				Expect(mapFlags).To(HaveLen(1))
				Expect(mapFlags["enable-feature"]).To(Equal(""))
			})

			It("should parse multiple boolean flags", func() {
				err := mapFlags.Set("enable-feature-a;enable-feature-b;debug-mode")
				Expect(err).ToNot(HaveOccurred())
				Expect(mapFlags).To(HaveLen(3))
				Expect(mapFlags["enable-feature-a"]).To(Equal(""))
				Expect(mapFlags["enable-feature-b"]).To(Equal(""))
				Expect(mapFlags["debug-mode"]).To(Equal(""))
			})

			It("should trim whitespace from boolean flags", func() {
				err := mapFlags.Set("  enable-feature  ;  another-flag  ")
				Expect(err).ToNot(HaveOccurred())
				Expect(mapFlags).To(HaveLen(2))
				Expect(mapFlags["enable-feature"]).To(Equal(""))
				Expect(mapFlags["another-flag"]).To(Equal(""))
			})
		})

		Context("with mixed formats", func() {
			It("should parse mix of key=value and boolean flags", func() {
				err := mapFlags.Set("license-key=ABC123;enable-feature;url=https://example.com;debug")
				Expect(err).ToNot(HaveOccurred())
				Expect(mapFlags).To(HaveLen(4))
				Expect(mapFlags["license-key"]).To(Equal("ABC123"))
				Expect(mapFlags["enable-feature"]).To(Equal(""))
				Expect(mapFlags["url"]).To(Equal("https://example.com"))
				Expect(mapFlags["debug"]).To(Equal(""))
			})
		})

		Context("with edge cases", func() {
			It("should handle empty string", func() {
				err := mapFlags.Set("")
				Expect(err).ToNot(HaveOccurred())
				Expect(mapFlags).To(HaveLen(0))
			})

			It("should handle only semicolons", func() {
				err := mapFlags.Set(";;;")
				Expect(err).ToNot(HaveOccurred())
				Expect(mapFlags).To(HaveLen(0))
			})

			It("should handle trailing semicolon", func() {
				err := mapFlags.Set("key1=value1;key2=value2;")
				Expect(err).ToNot(HaveOccurred())
				Expect(mapFlags).To(HaveLen(2))
				Expect(mapFlags["key1"]).To(Equal("value1"))
				Expect(mapFlags["key2"]).To(Equal("value2"))
			})

			It("should handle leading semicolon", func() {
				err := mapFlags.Set(";key1=value1;key2=value2")
				Expect(err).ToNot(HaveOccurred())
				Expect(mapFlags).To(HaveLen(2))
				Expect(mapFlags["key1"]).To(Equal("value1"))
				Expect(mapFlags["key2"]).To(Equal("value2"))
			})

			It("should handle consecutive semicolons", func() {
				err := mapFlags.Set("key1=value1;;key2=value2")
				Expect(err).ToNot(HaveOccurred())
				Expect(mapFlags).To(HaveLen(2))
				Expect(mapFlags["key1"]).To(Equal("value1"))
				Expect(mapFlags["key2"]).To(Equal("value2"))
			})
		})

		Context("with multiple Set calls", func() {
			It("should accumulate values from multiple calls", func() {
				err := mapFlags.Set("key1=value1")
				Expect(err).ToNot(HaveOccurred())

				err = mapFlags.Set("key2=value2")
				Expect(err).ToNot(HaveOccurred())

				Expect(mapFlags).To(HaveLen(2))
				Expect(mapFlags["key1"]).To(Equal("value1"))
				Expect(mapFlags["key2"]).To(Equal("value2"))
			})

			It("should overwrite existing keys", func() {
				err := mapFlags.Set("key1=value1")
				Expect(err).ToNot(HaveOccurred())

				err = mapFlags.Set("key1=newvalue")
				Expect(err).ToNot(HaveOccurred())

				Expect(mapFlags).To(HaveLen(1))
				Expect(mapFlags["key1"]).To(Equal("newvalue"))
			})
		})

		Context("with real-world examples", func() {
			It("should parse runai license key configuration", func() {
				err := mapFlags.Set("runai-license-key=XXXX-YYYY-ZZZZ")
				Expect(err).ToNot(HaveOccurred())
				Expect(mapFlags["runai-license-key"]).To(Equal("XXXX-YYYY-ZZZZ"))
			})

			It("should parse complex configuration string", func() {
				configStr := "runai-license-key=ABC123;enable-feature-x;control-plane.url=https://app.run.ai;control-plane.insecure=false;debug-mode"
				err := mapFlags.Set(configStr)
				Expect(err).ToNot(HaveOccurred())
				Expect(mapFlags).To(HaveLen(5))
				Expect(mapFlags["runai-license-key"]).To(Equal("ABC123"))
				Expect(mapFlags["enable-feature-x"]).To(Equal(""))
				Expect(mapFlags["control-plane.url"]).To(Equal("https://app.run.ai"))
				Expect(mapFlags["control-plane.insecure"]).To(Equal("false"))
				Expect(mapFlags["debug-mode"]).To(Equal(""))
			})
		})
	})
})

var _ = Describe("ArrayFlags (pre-req parser)", func() {
	var arrayFlags config.ArrayFlags

	BeforeEach(func() {
		arrayFlags = config.ArrayFlags{}
	})

	Describe("String method", func() {
		It("should return string representation of empty array", func() {
			result := arrayFlags.String()
			Expect(result).To(Equal("[]"))
		})

		It("should return string representation of populated array", func() {
			arrayFlags = config.ArrayFlags{"item1", "item2"}
			result := arrayFlags.String()
			Expect(result).To(Equal("[item1 item2]"))
		})
	})

	Describe("Set method", func() {
		It("should parse single value", func() {
			err := arrayFlags.Set("value1")
			Expect(err).ToNot(HaveOccurred())
			Expect(arrayFlags).To(HaveLen(1))
			Expect(arrayFlags[0]).To(Equal("value1"))
		})

		It("should parse comma-separated values", func() {
			err := arrayFlags.Set("value1,value2,value3")
			Expect(err).ToNot(HaveOccurred())
			Expect(arrayFlags).To(HaveLen(3))
			Expect(arrayFlags).To(Equal(config.ArrayFlags{"value1", "value2", "value3"}))
		})

		It("should trim whitespace around commas", func() {
			err := arrayFlags.Set("value1 , value2 , value3")
			Expect(err).ToNot(HaveOccurred())
			Expect(arrayFlags).To(HaveLen(3))
			Expect(arrayFlags).To(Equal(config.ArrayFlags{"value1", "value2", "value3"}))
		})

		It("should handle values with spaces inside", func() {
			err := arrayFlags.Set("value one,value two")
			Expect(err).ToNot(HaveOccurred())
			Expect(arrayFlags).To(HaveLen(2))
			Expect(arrayFlags).To(Equal(config.ArrayFlags{"value one", "value two"}))
		})
	})
})

var _ = Describe("InstallerMode", func() {
	var mode config.InstallerMode

	Describe("String method", func() {
		It("should return 'tui' for ModeTUI", func() {
			mode = config.ModeTUI
			Expect(mode.String()).To(Equal("tui"))
		})

		It("should return 'lib' for ModeCLI", func() {
			mode = config.ModeCLI
			Expect(mode.String()).To(Equal("lib"))
		})

		It("should return 'tui' for invalid value", func() {
			mode = config.InstallerMode(999)
			Expect(mode.String()).To(Equal("tui"))
		})
	})

	Describe("Set method", func() {
		It("should set ModeTUI for 'tui'", func() {
			err := mode.Set("tui")
			Expect(err).ToNot(HaveOccurred())
			Expect(mode).To(Equal(config.ModeTUI))
		})

		It("should set ModeCLI for 'lib'", func() {
			err := mode.Set("lib")
			Expect(err).ToNot(HaveOccurred())
			Expect(mode).To(Equal(config.ModeCLI))
		})

		It("should handle case insensitivity", func() {
			err := mode.Set("TUI")
			Expect(err).ToNot(HaveOccurred())
			Expect(mode).To(Equal(config.ModeTUI))

			err = mode.Set("LIB")
			Expect(err).ToNot(HaveOccurred())
			Expect(mode).To(Equal(config.ModeCLI))
		})

		It("should default to ModeTUI for invalid value", func() {
			err := mode.Set("invalid")
			Expect(err).ToNot(HaveOccurred())
			Expect(mode).To(Equal(config.ModeTUI))
		})
	})
})

var _ = Describe("GPUType", func() {
	var gpuType config.GPUType

	Describe("String method", func() {
		It("should return 'real' for GPUTypeReal", func() {
			gpuType = config.GPUTypeReal
			Expect(gpuType.String()).To(Equal("real"))
		})

		It("should return 'fake' for GPUTypeFake", func() {
			gpuType = config.GPUTypeFake
			Expect(gpuType.String()).To(Equal("fake"))
		})

		It("should return 'real' for invalid value", func() {
			gpuType = config.GPUType(999)
			Expect(gpuType.String()).To(Equal("real"))
		})
	})

	Describe("Set method", func() {
		It("should set GPUTypeReal for 'real'", func() {
			err := gpuType.Set("real")
			Expect(err).ToNot(HaveOccurred())
			Expect(gpuType).To(Equal(config.GPUTypeReal))
		})

		It("should set GPUTypeFake for 'fake'", func() {
			err := gpuType.Set("fake")
			Expect(err).ToNot(HaveOccurred())
			Expect(gpuType).To(Equal(config.GPUTypeFake))
		})

		It("should handle case insensitivity", func() {
			err := gpuType.Set("REAL")
			Expect(err).ToNot(HaveOccurred())
			Expect(gpuType).To(Equal(config.GPUTypeReal))

			err = gpuType.Set("FAKE")
			Expect(err).ToNot(HaveOccurred())
			Expect(gpuType).To(Equal(config.GPUTypeFake))
		})

		It("should default to GPUTypeReal for invalid value", func() {
			err := gpuType.Set("invalid")
			Expect(err).ToNot(HaveOccurred())
			Expect(gpuType).To(Equal(config.GPUTypeReal))
		})
	})
})

var _ = Describe("RunaiEnvType", func() {
	var envType config.RunaiEnvType

	Describe("String method", func() {
		It("should return 'prod' for EnvProd", func() {
			envType = config.EnvProd
			Expect(envType.String()).To(Equal("prod"))
		})

		It("should return 'test' for EnvTest", func() {
			envType = config.EnvTest
			Expect(envType.String()).To(Equal("test"))
		})

		It("should return 'test' for invalid value", func() {
			envType = config.RunaiEnvType(999)
			Expect(envType.String()).To(Equal("test"))
		})
	})

	Describe("Set method", func() {
		It("should set EnvProd for 'prod'", func() {
			err := envType.Set("prod")
			Expect(err).ToNot(HaveOccurred())
			Expect(envType).To(Equal(config.EnvProd))
		})

		It("should set EnvTest for 'test'", func() {
			err := envType.Set("test")
			Expect(err).ToNot(HaveOccurred())
			Expect(envType).To(Equal(config.EnvTest))
		})

		It("should handle case insensitivity", func() {
			err := envType.Set("PROD")
			Expect(err).ToNot(HaveOccurred())
			Expect(envType).To(Equal(config.EnvProd))

			err = envType.Set("TEST")
			Expect(err).ToNot(HaveOccurred())
			Expect(envType).To(Equal(config.EnvTest))
		})

		It("should default to EnvTest for invalid value", func() {
			err := envType.Set("invalid")
			Expect(err).ToNot(HaveOccurred())
			Expect(envType).To(Equal(config.EnvTest))
		})
	})
})

var _ = Describe("Config Integration", func() {
	// Note: ParseConfig uses flag.Parse() which affects global state,
	// making it difficult to test in isolation. These tests would require
	// resetting flag.CommandLine between tests or using a different approach.
	// For production code, consider refactoring ParseConfig to accept a FlagSet.

	Context("with flag integration", func() {
		It("should work with flag.Var for MapFlags", func() {
			// This demonstrates that MapFlags implements flag.Value interface
			var testFlags config.MapFlags = make(config.MapFlags)
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			fs.Var(&testFlags, "config-values", "test config values")

			err := fs.Parse([]string{"-config-values", "key1=value1;enable-feature"})
			Expect(err).ToNot(HaveOccurred())
			Expect(testFlags).To(HaveLen(2))
			Expect(testFlags["key1"]).To(Equal("value1"))
			Expect(testFlags["enable-feature"]).To(Equal(""))
		})

		It("should work with flag.Var for ArrayFlags", func() {
			// This demonstrates that ArrayFlags implements flag.Value interface
			var testFlags config.ArrayFlags
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			fs.Var(&testFlags, "pre-req", "test pre-req values")

			err := fs.Parse([]string{"-pre-req", "step1,step2,step3"})
			Expect(err).ToNot(HaveOccurred())
			Expect(testFlags).To(HaveLen(3))
			Expect(testFlags).To(Equal(config.ArrayFlags{"step1", "step2", "step3"}))
		})
	})
})
