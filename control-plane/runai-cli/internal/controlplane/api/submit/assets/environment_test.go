/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package assets

import (
	"slices"
	"strings"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/submit"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/spec/spectype"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/flags/update"
	"github.com/spf13/pflag"
)

func TestEnvironmentAssetFlags(t *testing.T) {
	for _, operation := range []string{"submit", "update"} {
		t.Run(operation, func(t *testing.T) {
			for _, tc := range []struct {
				name      string
				args      []string
				asset     string
				variables []string
				wantError bool
			}{
				{name: "omitted"},
				{name: "variables without asset", args: []string{"--environment-variable=A=x=y", "-e", "B="}, variables: []string{"A=x=y", "B="}},
				{name: "repeated asset uses last name", args: []string{"--environment=first", "--environment=second"}, asset: "second"},
				{name: "existing variables survive rejected input", args: []string{"-e", "EXISTING=value", "--environment=INVALID=value"}, variables: []string{"EXISTING=value"}, wantError: true},
				{name: "asset", args: []string{"--environment", "my-environment"}, asset: "my-environment"},
				{name: "asset and variables", args: []string{"--environment=my-environment", "--environment-variable=A=x=y", "-e", "B="}, asset: "my-environment", variables: []string{"A=x=y", "B="}},
				{name: "key value rejected", args: []string{"--environment", "A=value"}, wantError: true},
				{name: "equals syntax rejected", args: []string{"--environment=A=x=y"}, wantError: true},
				{name: "empty value rejected", args: []string{"--environment=A="}, wantError: true},
				{name: "asset then key value rejected", args: []string{"--environment=my-environment", "--environment=A=value"}, asset: "my-environment", wantError: true},
			} {
				t.Run(tc.name, func(t *testing.T) {
					c := &submit.Container{}
					flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
					if operation == "submit" {
						submit.ContainerFlagSet(flags, c, spectype.Training, api.UndefinedFramework)
					} else {
						update.ContainerUpdateFlagSet(flags, c, spectype.Inference)
					}
					err := flags.Parse(tc.args)
					if tc.wantError {
						if err == nil || !strings.Contains(err.Error(), "use --environment-variable KEY=VALUE") {
							t.Fatalf("expected migration error, got %v", err)
						}
					} else if err != nil {
						t.Fatal(err)
					}
					if c.EnvironmentAsset != tc.asset || !slices.Equal(c.EnvironmentVariables, tc.variables) {
						t.Fatalf("got asset %q and variables %v; want %q and %v", c.EnvironmentAsset, c.EnvironmentVariables, tc.asset, tc.variables)
					}
					asset, err := flags.GetString("environment")
					if err != nil || asset != tc.asset {
						t.Fatalf("asset flag getter returned %q, %v; want %q", asset, err, tc.asset)
					}
					flag := flags.Lookup("environment")
					if flag.Hidden || flag.Deprecated != "" {
						t.Fatal("environment asset selector must remain visible and supported")
					}
				})
			}
		})
	}
}
