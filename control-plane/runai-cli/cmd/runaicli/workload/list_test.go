/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload_test

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/list/manipulate"
	"github.com/spf13/cobra"

	listCommon "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/workload"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/settings"
)

var _ = Describe("HandleColumnsFlags", func() {
	var (
		cmd     *cobra.Command
		output  listCommon.Output
		buf     *bytes.Buffer
		tempDir string
		err     error
	)

	BeforeEach(func() {
		buf = &bytes.Buffer{}
		cmd = &cobra.Command{}
		cmd.SetOut(buf)
		output = listCommon.Output{}

		// Create temporary directory for test configs
		tempDir, err = os.MkdirTemp("", "workload-test-*")
		Expect(err).ToNot(HaveOccurred())

		// Initialize config and settings with temp directory
		configInstance := config.GetInstance()
		err = configInstance.Initializing(config.Options{ConfigPath: tempDir, ConfigFile: "test_config.json"})
		Expect(err).ToNot(HaveOccurred())

		settingsInstance := settings.GetInstance()
		err = settingsInstance.Initializing(tempDir)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		// Clean up temporary directory
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
	})

	Context("Valid columns with set-default-columns flag", func() {
		It("should save valid columns to settings", func() {
			// Set test values using helper function
			setOutputValues(&output, []string{"name", "status", "project"}, true)

			err := workload.HandleColumnsFlags(output, cmd)

			Expect(err).ToNot(HaveOccurred())
			// Verify columns were saved to settings
			savedColumns := settings.GetInstance().GetWorkloadColumns()
			Expect(savedColumns).To(Equal([]string{"name", "status", "project"}))
			// Verify output message
			Expect(buf.String()).To(ContainSubstring("Default workload columns set to:"))
		})
	})

	Context("Invalid columns with set-default-columns flag", func() {
		It("should return error and not save to settings", func() {
			// Get initial state
			initialColumns := settings.GetInstance().GetWorkloadColumns()
			initialLen := len(initialColumns)

			setOutputValues(&output, []string{"invalid-column", "another-invalid"}, true)

			err := workload.HandleColumnsFlags(output, cmd)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid column names"))
			// Verify no columns were saved (should remain unchanged from initial state)
			savedColumns := settings.GetInstance().GetWorkloadColumns()
			Expect(len(savedColumns)).To(Equal(initialLen))
		})
	})

	Context("Empty columns with set-default-columns flag", func() {
		It("should save empty columns and show default message", func() {
			setOutputValues(&output, []string{}, true)

			err := workload.HandleColumnsFlags(output, cmd)

			Expect(err).ToNot(HaveOccurred())
			// Verify empty columns were saved
			savedColumns := settings.GetInstance().GetWorkloadColumns()
			Expect(len(savedColumns)).To(Equal(0))
			// Verify output shows default columns
			Expect(buf.String()).To(ContainSubstring("Default workload columns set"))
		})
	})

	Context("Valid columns without set-default-columns flag", func() {
		It("should validate columns but not save to settings", func() {
			// Check initial state
			initialColumns := settings.GetInstance().GetWorkloadColumns()
			initialLen := len(initialColumns)

			setOutputValues(&output, []string{"name", "status"}, false)

			err := workload.HandleColumnsFlags(output, cmd)

			Expect(err).ToNot(HaveOccurred())
			// Verify no columns were saved (should remain unchanged)
			savedColumns := settings.GetInstance().GetWorkloadColumns()
			Expect(len(savedColumns)).To(Equal(initialLen))
			// Verify no output message
			Expect(buf.String()).To(BeEmpty())
		})
	})

	Context("Command output verification", func() {
		It("should print correct message for custom columns", func() {
			setOutputValues(&output, []string{"name", "type", "status"}, true)

			err := workload.HandleColumnsFlags(output, cmd)

			Expect(err).ToNot(HaveOccurred())
			outputStr := buf.String()
			Expect(outputStr).To(ContainSubstring("Default workload columns set to:"))
			Expect(outputStr).To(ContainSubstring("[name type status]"))
		})

		It("should print correct message for empty columns", func() {
			setOutputValues(&output, []string{}, true)

			err := workload.HandleColumnsFlags(output, cmd)

			Expect(err).ToNot(HaveOccurred())
			outputStr := buf.String()
			// Should contain the default columns list
			Expect(outputStr).To(ContainSubstring("Default workload columns set"))
			Expect(outputStr).To(ContainSubstring(fmt.Sprintf("%s", manipulate.GetDefaultColumns())))
		})
	})
})

// Helper function to set private fields for testing
func setOutputValues(output *listCommon.Output, columns []string, setDefaultColumns bool) {
	// Use flag binding to set private fields
	cmd := &cobra.Command{}
	listCommon.FlagOutputType(cmd, output)
	listCommon.FlagOutputTypeWithColumns(cmd, output)

	// Set flag values
	if len(columns) > 0 {
		_ = cmd.Flags().Set(listCommon.ColumnsStr, strings.Join(columns, ","))
	}
	if setDefaultColumns {
		_ = cmd.Flags().Set(listCommon.SetDefaultColumnsStr, "true")
	}
}
