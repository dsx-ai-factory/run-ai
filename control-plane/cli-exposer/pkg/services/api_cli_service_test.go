/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_test

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/cli-exposer/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/cli-exposer/pkg/consts"
	"github.com/run-ai/runai/control-plane/cli-exposer/pkg/document"
	"github.com/run-ai/runai/control-plane/cli-exposer/pkg/services"
	"github.com/run-ai/runai/control-plane/cli-exposer/pkg/version"
	cliversion "github.com/run-ai/runai/control-plane/runai-cli/pkg/version"
	middlewares "github.com/run-ai/runai/runai-application-packages/old-auth-token-middleware"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultCLIVersion = "RUNAI_CLI_MAJOR=2\nRUNAI_CLI_MINOR=50\nRUNAI_CLI_PATCH=0"
)

var _ = Describe("ApiCliService", func() {
	var (
		cliApiService  api.ResearcherCommandLineInterfaceAPIServicer
		ctx            context.Context
		tmpDir         string
		validators     *services.RegexValidators
		cliVersion     cliversion.Version
		htmlDocContent document.HtmlContent
	)

	BeforeEach(func() {
		var err error
		validators, err = services.GetRegexValidators()
		Expect(err).NotTo(HaveOccurred())

		tmpDir, err = os.MkdirTemp("", "test")
		Expect(err).NotTo(HaveOccurred())
		Expect(err).NotTo(HaveOccurred())

		// Create a version file
		fileObj, err := os.CreateTemp(tmpDir, "version")
		Expect(err).NotTo(HaveOccurred())
		cacheFile := fileObj.Name()
		_, err = fileObj.WriteString(defaultCLIVersion)
		Expect(err).NotTo(HaveOccurred())
		version.SetVersionFilePath(cacheFile)
		cliVersion, err = version.GetVersion()
		Expect(err).NotTo(HaveOccurred())

		// Create a static files
		homePath, err := getHomePath()
		Expect(err).NotTo(HaveOccurred())
		document.SetStaticFilePath(filepath.Join(homePath, "internal/static"))
		htmlDocContent, err = document.GetHtmlContent()
		Expect(err).NotTo(HaveOccurred())

		cliApiService = services.NewCliApiService(validators, cliVersion, htmlDocContent)
		Expect(err).NotTo(HaveOccurred())
		ctx = context.WithValue(context.Background(), runaictx.CfTenantId, "12345")
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).NotTo(HaveOccurred())
	})

	// Test Get Version
	Describe("GetVersion", func() {
		It("Should fail on missing version file", func() {
			version.SetVersionFilePath("invalid_path")
			cliVersion, err := version.GetVersion()
			Expect(err).To(HaveOccurred())
			Expect(cliVersion).To(Equal(cliversion.Version{}))
		})

		It("Should fail on invalid semver version in server", func() {
			cliVersion.Version = "invalid_semver_version"
			cliApiService = services.NewCliApiService(validators, cliVersion, htmlDocContent)
			resp, err := cliApiService.GetVersion(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Code).To(Equal(http.StatusInternalServerError))
			Expect(resp.Body).To(Equal("invalid internal CLI version"))
		})

		It("Should fail on missing user agent", func() {
			//ctx := context.WithValue(context.Background(), consts.ContextHostUrl, "http://localhost:8080")
			resp, err := cliApiService.GetVersion(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body).To(Equal("missing user-agent header"))
		})

		It("Should fail on invalid user agent", func() {
			ctx = context.WithValue(ctx, consts.UserAgentHeader, "invalid_user_agent")
			resp, err := cliApiService.GetVersion(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body).To(Equal("invalid User-Agent Header; runai-cli version is invalid"))

		})

		It("Should fail on invalid user agent semver", func() {
			ctx = context.WithValue(ctx, consts.UserAgentHeader, "sdk/invalid_semver_version")
			resp, err := cliApiService.GetVersion(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body).To(Equal("invalid User-Agent Header; runai-cli version is invalid"))
		})

		It("Should fail on version mismatch - client higher version", func() {
			ctx = context.WithValue(ctx, consts.UserAgentHeader, "sdk/2.50.1")
			resp, err := cliApiService.GetVersion(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body).To(HavePrefix("versions mismatch, client SDK version: 2.50.1 is bigger than server SDK version: 2.50.0"))
		})

		It("Should return not need to update - same version", func() {
			ctx = context.WithValue(ctx, consts.UserAgentHeader, "sdk/2.50.0")
			resp, err := cliApiService.GetVersion(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Code).To(Equal(http.StatusOK))
			body := resp.Body.(api.Version)
			Expect(body.SdkVersion).To(Equal("2.50.0"))
			Expect(body.NeedUpgrade).To(Equal(false))
		})

		It("Should return need to update the version", func() {
			ctx = context.WithValue(ctx, consts.UserAgentHeader, "sdk/2.49.0")
			resp, err := cliApiService.GetVersion(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Code).To(Equal(http.StatusOK))
			body := resp.Body.(api.Version)
			Expect(body.SdkVersion).To(Equal("2.50.0"))
			Expect(body.NeedUpgrade).To(Equal(true))
		})
	})

	// Test Get Binary
	Describe("GetBinary", func() {
		Context("When the binary does not exist", func() {
			It("Should return 404", func() {
				operationSystem := "abc"
				arch := "amd64"
				resp, err := cliApiService.GetBinary(ctx, operationSystem, arch)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(http.StatusNotFound))
				Expect(resp.Body).To(Equal(fmt.Sprintf(consts.ErrBinaryNotExist, operationSystem, arch)))

				operationSystem = "darwin"
				arch = "amd"
				resp, err = cliApiService.GetBinary(ctx, operationSystem, arch)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(http.StatusNotFound))
				Expect(resp.Body).To(Equal(fmt.Sprintf(consts.ErrBinaryNotExist, operationSystem, arch)))
			})
		})

	})

	// Test Get Installer Unix
	Describe("GetInstallerUnix", func() {
		It("Should fail on missing token", func() {
			resp, err := cliApiService.GetInstallerUnix(ctx)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body.(string)).To(Equal(consts.ErrMissingToken))
		})

		It("Should fail on missing CP URL", func() {
			ctx := context.WithValue(context.Background(), middlewares.ContextJwtToken, "12345")
			resp, err := cliApiService.GetInstallerUnix(ctx)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body.(string)).To(Equal(consts.ErrorGettingCpURL))
		})

		It("Should fail on missing template file", func() {
			ctx := context.WithValue(context.Background(), middlewares.ContextJwtToken, "12345")
			ctx = context.WithValue(ctx, consts.ContextHostUrl, "http://localhost:8080")
			resp, err := cliApiService.GetInstallerUnix(ctx)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.Code).To(Equal(http.StatusInternalServerError))
			Expect(resp.Body.(string)).To(HavePrefix("failed to parse installer template file"))
		})
	})

	// Test Get Installer Windows
	//Describe("GetInstallerWindows", func() {
	//	It("Should return the installer script", func() {
	//		resp, err := cliApiService.GetInstallerWindows(ctx)
	//		Expect(err).ShouldNot(HaveOccurred())
	//		Expect(resp.Code).To(Equal(http.StatusOK))
	//
	//		body, ok := resp.Body.([]uint8)
	//		Expect(ok).To(BeTrue())
	//		strValue := string(body)
	//		Expect(strValue).To(Equal("Windows installer is not supported yet"))
	//	})
	//})

	// Test Get Manual
	Describe("GetManualDocument", func() {
		Context("When the manual does not exist", func() {
			It("Should return 404", func() {
				docName := "testName.md"
				resp, err := cliApiService.GetManualDocument(ctx, docName)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(http.StatusNotFound))
				Expect(resp.Body).To(Equal(fmt.Sprintf(consts.ErrDocumentNotExist, docName)))
			})
		})
		Context("When the path is invalid", func() {
			It("Should return 403", func() {
				docName := "testName"
				resp, err := cliApiService.GetManualDocument(ctx, docName)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(http.StatusForbidden))
				Expect(resp.Body).To(Equal(fmt.Sprintf(consts.ErrForbiddenDocument, docName)))
			})
		})
	})
})

func getHomePath() (string, error) {
	homePath, err := os.Getwd()
	if err != nil {
		return "", err
	}

	if !strings.Contains(homePath, "/cli-exposer/") {
		return "", fmt.Errorf("couldn't locate CLI Exposer folder")
	}
	homePath = strings.SplitAfter(homePath, "cli-exposer/")[0]
	return homePath, err
}
