/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cmdtxn

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	pathUtil "path"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	clierrors "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/errors"
	"github.com/spf13/cobra"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cmdtxn Tests")
}

var _ = Describe("cmdtxn", Ordered, func() {
	var (
		cmd    *cobra.Command
		tmpDir string
	)

	BeforeAll(func() {
		// create a temporary directory for config file
		var err error
		tmpDir, err = os.MkdirTemp("", "test")
		Expect(err).NotTo(HaveOccurred())
		fullPath := pathUtil.Join(tmpDir, "test_cache.json")
		file, err := os.Create(fullPath)
		Expect(err).NotTo(HaveOccurred())
		defer file.Close()
		_, err = file.WriteString("{}\n")
		Expect(err).NotTo(HaveOccurred())
		err = config.GetInstance().Initializing(config.Options{Debug: true, ConfigPath: tmpDir, ConfigFile: "test_cache.json"})
		Expect(err).NotTo(HaveOccurred())
	})

	AfterAll(func() {
		_ = os.Remove(tmpDir)
	})

	BeforeEach(func() {
		_ = transaction.NewTxn()
		cmd = &cobra.Command{}
	})

	Describe("CmdErrorPrintAndSet", func() {
		It("should print error message and return error", func() {
			err := CmdErrorPrintAndSet(cmd, "Test message", errors.New("test error"))
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("handleErrorsMsg", func() {
		Context("with HttpError", func() {
			It("should return PermissionDenied for 403 Forbidden", func() {
				httpErr := &clierrors.HttpError{Status: "403 Forbidden", Message: "Permission denied"}
				result := handleErrorsMsg(httpErr)
				Expect(result).To(Equal(consts.PermissionDenied))
			})

			It("should return Unauthorized message for 401 Unauthorized", func() {
				httpErr := &clierrors.HttpError{Status: "401 Unauthorized", Message: `{"code":401,
"message":"401 Unauthorized: invalid_grant: Invalid user credentials"}`}
				result := handleErrorsMsg(httpErr)
				Expect(result).To(Equal("401 Unauthorized: invalid_grant: Invalid user credentials"))
			})

			It("should return API error message for other cases", func() {
				httpErr := &clierrors.HttpError{Status: "409 Conflict", Message: `{"code":409,
"message":"Workspace named 'aaa' already exists in the project"}`}
				result := handleErrorsMsg(httpErr)
				Expect(result).To(Equal("Workspace named 'aaa' already exists in the project"))
			})

			It("should return API error message for other cases", func() {
				httpErr := &clierrors.HttpError{Status: "500 Server Error", Message: `{"code":500, "message":"API error"}`}
				result := handleErrorsMsg(httpErr)
				Expect(result).To(Equal("API error"))
			})

			It("should return API error message without marshaling", func() {
				httpErr := &clierrors.HttpError{Status: "500 Server Error", Message: `{"codes":"500"", "message":"API error"}`}
				result := handleErrorsMsg(httpErr)
				Expect(result).To(Equal(`{"codes":"500"", "message":"API error"}`))
			})

			It("should parse error response wrapped in a JSON array with compliance issues", func() {
				httpErr := &clierrors.HttpError{Status: "409 Conflict", Message: `[{"code":409,"message":"resource already exists","complianceIssues":[{"field":"name","details":"duplicate name"}]}]`}
				result := handleErrorsMsg(httpErr)
				Expect(result).To(Equal("resource already exists\nField: name, Details: duplicate name"))
			})

			It("should keep compliance and skip redundant message when details repeat the message", func() {
				httpErr := &clierrors.HttpError{Status: "409 Conflict", Message: `[{"code":409,"message":"workload already exists","complianceIssues":[{"field":"name","details":"workload already exists"}]}]`}
				result := handleErrorsMsg(httpErr)
				Expect(result).To(Equal("\nField: name, Details: workload already exists"))
			})

			It("should parse error response wrapped in a JSON array without compliance issues", func() {
				httpErr := &clierrors.HttpError{Status: "400 Bad Request", Message: `[{"code":400,"message":"some API error"}]`}
				result := handleErrorsMsg(httpErr)
				Expect(result).To(Equal("some API error"))
			})

			It("should parse JSON array with multiple error entries", func() {
				httpErr := &clierrors.HttpError{Status: "400 Bad Request", Message: `[{"code":400,"message":"first error"},{"code":400,"message":"second error"}]`}
				result := handleErrorsMsg(httpErr)
				Expect(result).To(Equal("first error\nsecond error"))
			})
		})

		Context("with url.Error", func() {
			It("should return BadUrlMsg for DNS errors", func() {
				dnsErr := &net.DNSError{Name: "example.com", Err: "DNS error"}
				netErr := &net.OpError{Err: dnsErr}
				urlErr := &url.Error{Err: netErr}
				result := handleErrorsMsg(urlErr)
				Expect(result).To(Equal(fmt.Sprintf(consts.BadUrlMsg, "example.com", "DNS error")))
			})
		})

		It("should return error message for other errors", func() {
			err := errors.New("generic error")
			result := handleErrorsMsg(err)
			Expect(result).To(Equal("generic error"))
		})
	})

	Describe("CmdErrorValidationMessage", func() {
		It("should print validation error message and return error", func() {
			err := CmdErrorValidationMessage(cmd, "Validation error: %s", errors.New("invalid input"))
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("CmdErrorPrintAndSetSuffix", func() {
		It("should print error message with suffix and return error", func() {
			err := CmdErrorPrintAndSetSuffix(cmd, "Test message", errors.New("test error"), " (suffix)")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("CmdErrorSetMissingParams", func() {
		It("should return error with missing params message", func() {
			err := CmdErrorSetMissingParams(cmd)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(MissingParamsErrorMsg))
		})
	})
})
