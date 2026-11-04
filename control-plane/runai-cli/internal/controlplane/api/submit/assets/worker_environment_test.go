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

func TestWorkerEnvironmentVariableFlag(t *testing.T) {
	for _, tc := range []struct {
		name      string
		args      []string
		want      []string
		wantError bool
	}{
		{name: "canonical", args: []string{"--worker-environment-variable=A=x=y", "--worker-environment-variable=B="}, want: []string{"A=x=y", "B="}},
		{name: "removed flag", args: []string{"--worker-environment=A=1"}, wantError: true},
		{name: "removed flag with separate value", args: []string{"--worker-environment", "A=1"}, wantError: true},
		{name: "canonical then removed flag", args: []string{"--worker-environment-variable=A=1", "--worker-environment=B=2"}, wantError: true},
		{name: "removed flag then canonical", args: []string{"--worker-environment=A=1", "--worker-environment-variable=B=2"}, wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			container := &submit.Container{}
			flags := pflag.NewFlagSet("worker", pflag.ContinueOnError)
			var output bytes.Buffer
			flags.SetOutput(&output)
			submit.DistributedInferenceContainerFlagSet(flags, container)
			if flags.Lookup("worker-environment") != nil {
				t.Fatal("removed flag must not be registered")
			}
			canonical := flags.Lookup("worker-environment-variable")
			if canonical == nil || canonical.Hidden || canonical.Deprecated != "" {
				t.Fatal("worker-environment-variable must remain visible and supported")
			}
			err := flags.Parse(tc.args)
			if tc.wantError {
				if err == nil || !strings.Contains(err.Error(), "unknown flag: --worker-environment") {
					t.Fatalf("parse error = %v; want unknown flag: --worker-environment", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(container.WorkerEnvironmentVariables, tc.want) {
				t.Fatalf("worker variables = %v; want %v", container.WorkerEnvironmentVariables, tc.want)
			}
			if output.Len() != 0 {
				t.Fatalf("unexpected output: %q", output.String())
			}
		})
	}
}
