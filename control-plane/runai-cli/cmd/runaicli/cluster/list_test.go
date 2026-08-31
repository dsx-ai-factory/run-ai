/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster_test

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/cluster/list/manipulate"
	"github.com/spf13/cobra"

	listCommon "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/list"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/cluster"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/settings"
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
		tempDir, err = os.MkdirTemp("", "cluster-test-*")
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
			setOutputValues(&output, []string{"name", "status", "url"}, true)

			err := cluster.HandleColumnsFlags(output, cmd)

			Expect(err).ToNot(HaveOccurred())
			// Verify columns were saved to settings
			savedColumns := settings.GetInstance().GetClusterColumns()
			Expect(savedColumns).To(Equal([]string{"name", "status", "url"}))
			// Verify output message
			Expect(buf.String()).To(ContainSubstring("Default cluster columns set to:"))
			Expect(buf.String()).To(ContainSubstring("[name status url]"))
		})
	})

	Context("Invalid columns with set-default-columns flag", func() {
		It("should return error and not save to settings", func() {
			// Get initial state
			initialColumns := settings.GetInstance().GetClusterColumns()
			initialLen := len(initialColumns)

			setOutputValues(&output, []string{"invalid-column", "another-invalid"}, true)

			err := cluster.HandleColumnsFlags(output, cmd)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unknown column"))
			// Verify no columns were saved (should remain unchanged from initial state)
			savedColumns := settings.GetInstance().GetClusterColumns()
			Expect(len(savedColumns)).To(Equal(initialLen))
		})
	})

	Context("Empty columns with set-default-columns flag", func() {
		It("should save empty columns and show default message", func() {
			setOutputValues(&output, []string{}, true)

			err := cluster.HandleColumnsFlags(output, cmd)

			Expect(err).ToNot(HaveOccurred())
			// Verify empty columns were saved
			savedColumns := settings.GetInstance().GetClusterColumns()
			Expect(len(savedColumns)).To(Equal(0))
			// Verify output shows default columns
			Expect(buf.String()).To(ContainSubstring("Default cluster columns set"))
		})
	})

	Context("Valid columns without set-default-columns flag", func() {
		It("should validate columns but not save to settings", func() {
			// Check initial state
			initialColumns := settings.GetInstance().GetClusterColumns()
			initialLen := len(initialColumns)

			setOutputValues(&output, []string{"name", "status"}, false)

			err := cluster.HandleColumnsFlags(output, cmd)

			Expect(err).ToNot(HaveOccurred())
			// Verify no columns were saved (should remain unchanged)
			savedColumns := settings.GetInstance().GetClusterColumns()
			Expect(len(savedColumns)).To(Equal(initialLen))
			// Verify no output message
			Expect(buf.String()).To(BeEmpty())
		})
	})

	Context("Command output verification", func() {
		It("should print correct message for custom columns", func() {
			setOutputValues(&output, []string{"name", "url", "status"}, true)

			err := cluster.HandleColumnsFlags(output, cmd)

			Expect(err).ToNot(HaveOccurred())
			outputStr := buf.String()
			Expect(outputStr).To(ContainSubstring("Default cluster columns set to:"))
			Expect(outputStr).To(ContainSubstring("[name url status]"))
		})

		It("should print correct message for empty columns", func() {
			setOutputValues(&output, []string{}, true)

			err := cluster.HandleColumnsFlags(output, cmd)

			Expect(err).ToNot(HaveOccurred())
			outputStr := buf.String()
			// Should contain the default columns list
			Expect(outputStr).To(ContainSubstring("Default cluster columns set"))
			Expect(outputStr).To(ContainSubstring(fmt.Sprintf("%s", manipulate.GetDefaultColumns())))
		})
	})

	Context("All valid cluster columns", func() {
		It("should accept all available cluster columns", func() {
			validColumns := []string{"name", "url", "version", "status", "kubernetesdistribution", "distributionversion"}
			setOutputValues(&output, validColumns, true)

			err := cluster.HandleColumnsFlags(output, cmd)

			Expect(err).ToNot(HaveOccurred())
			savedColumns := settings.GetInstance().GetClusterColumns()
			Expect(savedColumns).To(Equal(validColumns))
		})
	})

	Context("Mixed case columns", func() {
		It("should handle mixed case column names", func() {
			mixedCaseColumns := []string{"Name", "URL", "Status"}
			setOutputValues(&output, mixedCaseColumns, true)

			err := cluster.HandleColumnsFlags(output, cmd)

			Expect(err).ToNot(HaveOccurred())
			savedColumns := settings.GetInstance().GetClusterColumns()
			Expect(savedColumns).To(Equal(mixedCaseColumns))
		})
	})

	Context("Whitespace handling", func() {
		It("should handle columns with whitespace", func() {
			columnsWithWhitespace := []string{"  name  ", " url ", "status"}
			setOutputValues(&output, columnsWithWhitespace, true)

			err := cluster.HandleColumnsFlags(output, cmd)

			Expect(err).ToNot(HaveOccurred())
			// The function should still work correctly
			savedColumns := settings.GetInstance().GetClusterColumns()
			Expect(savedColumns).To(Equal(columnsWithWhitespace))
		})
	})

	Context("Duplicate columns", func() {
		It("should handle duplicate column names", func() {
			duplicateColumns := []string{"name", "name", "status"}
			setOutputValues(&output, duplicateColumns, true)

			err := cluster.HandleColumnsFlags(output, cmd)

			Expect(err).ToNot(HaveOccurred())
			savedColumns := settings.GetInstance().GetClusterColumns()
			Expect(savedColumns).To(Equal(duplicateColumns))
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
