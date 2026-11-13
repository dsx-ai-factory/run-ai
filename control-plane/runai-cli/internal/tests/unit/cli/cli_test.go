/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli"

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

var errProbe = errors.New("probe failed")

func TestRunCommandHelpersCaptureStreams(t *testing.T) {
	noop := func(*cobra.Command, []string) error { return nil }
	streams := &cobra.Command{
		Use:               "streams-probe",
		PersistentPreRunE: noop,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "normal output")
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "diagnostic")
			return errProbe
		},
	}
	argsProbe := &cobra.Command{
		Use:               "args-probe",
		PersistentPreRunE: noop,
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, a := range args {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), a)
			}
			return nil
		},
	}
	runaicli.RootCmd.AddCommand(streams, argsProbe)
	t.Cleanup(func() { runaicli.RootCmd.RemoveCommand(streams, argsProbe) })

	t.Run("separate", func(t *testing.T) {
		stdout, stderr, err := RunCommandSeparate(context.Background(), "streams-probe", nil)
		if !errors.Is(err, errProbe) {
			t.Fatalf("err = %v, want %v", err, errProbe)
		}
		if stdout != "normal output\n" {
			t.Errorf("stdout = %q, want %q", stdout, "normal output\n")
		}
		if !strings.Contains(stderr, "diagnostic") || strings.Contains(stderr, "normal output") {
			t.Errorf("stderr = %q, want diagnostic only", stderr)
		}
	})

	t.Run("legacy combined keeps chronological order", func(t *testing.T) {
		out, err := RunCommand("streams-probe", nil)
		if !errors.Is(err, errProbe) {
			t.Fatalf("err = %v, want %v", err, errProbe)
		}
		o, d := strings.Index(out, "normal output"), strings.Index(out, "diagnostic")
		if o < 0 || d < 0 || o > d {
			t.Errorf("combined output = %q, want normal output before diagnostic", out)
		}
	})

	t.Run("string split preserves argument boundaries", func(t *testing.T) {
		out, err := RunCommandStringSplit(context.Background(), []string{"args-probe", "a b", "c"}, nil)
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		if out != "a b\nc\n" {
			t.Errorf("out = %q, want %q", out, "a b\nc\n")
		}
	})
}
