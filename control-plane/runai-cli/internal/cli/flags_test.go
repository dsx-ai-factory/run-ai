/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cli_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cli"
	"github.com/spf13/cobra"
)

var _ = Describe("GetOutputTypeFromFlags", func() {
	It("returns json when only --json is set", func() {
		cmd := &cobra.Command{}
		cmd.Flags().Bool(cli.JsonOutput, false, "")
		cmd.Flags().Bool(cli.YamlOutput, false, "")
		cmd.Flags().Bool(cli.TableOutput, false, "")
		cmd.Flags().Set(cli.JsonOutput, "true")
		Expect(cli.GetOutputTypeFromFlags(cmd)).To(Equal(cli.JsonOutput))
	})

	It("returns yaml when only --yaml is set", func() {
		cmd := &cobra.Command{}
		cmd.Flags().Bool(cli.JsonOutput, false, "")
		cmd.Flags().Bool(cli.YamlOutput, false, "")
		cmd.Flags().Bool(cli.TableOutput, false, "")
		cmd.Flags().Set(cli.YamlOutput, "true")
		Expect(cli.GetOutputTypeFromFlags(cmd)).To(Equal(cli.YamlOutput))
	})

	It("returns table when only --table is set", func() {
		cmd := &cobra.Command{}
		cmd.Flags().Bool(cli.JsonOutput, false, "")
		cmd.Flags().Bool(cli.YamlOutput, false, "")
		cmd.Flags().Set(cli.TableOutput, "true")
		Expect(cli.GetOutputTypeFromFlags(cmd)).To(Equal(cli.TableOutput))
	})
	It("returns table when no output flags are set (default)", func() {
		cmd := &cobra.Command{}
		cmd.Flags().Bool(cli.JsonOutput, false, "")
		cmd.Flags().Bool(cli.YamlOutput, false, "")
		cmd.Flags().Bool(cli.TableOutput, false, "")
		// Do not set any flag
		Expect(cli.GetOutputTypeFromFlags(cmd)).To(Equal(cli.TableOutput))
	})
})
