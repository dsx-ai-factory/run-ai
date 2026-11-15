/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package assets

import (
	"bytes"
	"slices"
	"strings"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/submit"
	"github.com/spf13/pflag"
)

func TestWorkerEnvironmentAliases(t *testing.T) {
	for _, tc := range []struct {
		name       string
		args       []string
		want       []string
		deprecated bool
	}{
		{name: "canonical", args: []string{"--worker-environment-variable=A=x=y", "--worker-environment-variable=B="}, want: []string{"A=x=y", "B="}},
		{name: "legacy", args: []string{"--worker-environment=A=x=y", "--worker-environment=B="}, want: []string{"A=x=y", "B="}, deprecated: true},
		{name: "canonical then legacy", args: []string{"--worker-environment-variable=A=1", "--worker-environment=B=2"}, want: []string{"A=1", "B=2"}, deprecated: true},
		{name: "legacy then canonical", args: []string{"--worker-environment=A=1", "--worker-environment-variable=B=2"}, want: []string{"A=1", "B=2"}, deprecated: true},
		{name: "interleaved", args: []string{"--worker-environment=A=1", "--worker-environment-variable=B=2", "--worker-environment=C=3"}, want: []string{"A=1", "B=2", "C=3"}, deprecated: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			container := &submit.Container{}
			flags := pflag.NewFlagSet("worker", pflag.ContinueOnError)
			var output bytes.Buffer
			flags.SetOutput(&output)
			submit.DistributedInferenceContainerFlagSet(flags, container)
			if err := flags.Parse(tc.args); err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(container.WorkerEnvironmentVariables, tc.want) {
				t.Fatalf("worker variables = %v; want %v", container.WorkerEnvironmentVariables, tc.want)
			}
			warning := "Flag --worker-environment has been deprecated, use --worker-environment-variable instead"
			if strings.Contains(output.String(), warning) != tc.deprecated {
				t.Fatalf("unexpected warning output: %q", output.String())
			}
			for _, name := range []string{"worker-environment-variable", "worker-environment"} {
				values, err := flags.GetStringArray(name)
				if err != nil || !slices.Equal(values, tc.want) {
					t.Fatalf("%s getter = %v, %v; want %v", name, values, err, tc.want)
				}
			}
			canonical, legacy := flags.Lookup("worker-environment-variable"), flags.Lookup("worker-environment")
			if canonical.Hidden || canonical.Deprecated != "" || !legacy.Hidden || legacy.Deprecated == "" {
				t.Fatal("canonical flag must remain visible and supported; legacy alias must be hidden and deprecated")
			}
		})
	}
}
