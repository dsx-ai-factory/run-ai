//go:build e2e

/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package e2e_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/e2e/utils"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
)

var testEnv *utils.TestEnv

// installerBinaryPath is the installer binary compiled once for the whole suite
// in BeforeSuite and shared by every spec that drives the CLI.
var installerBinaryPath string

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Installer E2E Suite")
}

var _ = BeforeSuite(func() {
	var err error
	testEnv, err = utils.SetupTestEnv()
	Expect(err).ToNot(HaveOccurred())

	err = logging.InitFromDir(testEnv.TempDir)
	Expect(err).ToNot(HaveOccurred())

	installerBinaryPath, err = buildInstallerBinary()
	Expect(err).ToNot(HaveOccurred())
})

var _ = ReportAfterEach(func(report SpecReport) {
	if report.Failed() {
		utils.CollectDiagnostics(testEnv.TempDir)
	}
})

var _ = AfterSuite(func() {
	if installerBinaryPath != "" {
		os.Remove(installerBinaryPath)
	}
	if testEnv != nil {
		testEnv.Cleanup()
	}
})
