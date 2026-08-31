/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestResetCommandFlagsMatchesPositionalArgsAndAliases(t *testing.T) {
	var flagValue string

	sub := &cobra.Command{Use: "submit APPLICATION_NAME"}
	sub.Flags().StringVar(&flagValue, "values", "", "")

	parent := &cobra.Command{Use: "application", Aliases: []string{"applications"}}
	parent.AddCommand(sub)

	root := &cobra.Command{Use: "runai"}
	root.AddCommand(parent)

	cases := []struct {
		name     string
		cmdSplit []string
	}{
		{name: "canonical name with positional-arg Use", cmdSplit: []string{"application", "submit", "my-app"}},
		{name: "alias invocation", cmdSplit: []string{"applications", "submit", "my-app"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			flagValue = "leaked-from-previous-test"
			if err := sub.Flags().Set("values", "leaked-from-previous-test"); err != nil {
				t.Fatal(err)
			}

			resetCommandFlags(root.Commands(), tc.cmdSplit)

			if flagValue != "" {
				t.Errorf("flag was not reset for %v: got %q", tc.cmdSplit, flagValue)
			}
			if sub.Flags().Changed("values") {
				t.Errorf("flag Changed state was not reset for %v", tc.cmdSplit)
			}
		})
	}
}
