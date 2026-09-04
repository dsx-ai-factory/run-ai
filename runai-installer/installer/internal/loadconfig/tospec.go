/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package loadconfig

import (
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/artifacts"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/packconfig"
)

// RunOptions are the CLI-flag-derived knobs that map into the spec. Execution
// knobs (timeouts, preflight staging) live in specrun.ExecOptions.
type RunOptions struct {
	DryRun     bool
	Yes        bool
	Output     spec.OutputFormat
	EventsFile string
}

// ToSpec maps the config into the orchestrator spec, applying defaults and
// resolving the destination registry's credentials.
func (c *Config) ToSpec(opts RunOptions) (*spec.Spec, error) {
	destination := artifacts.NormalizeTarget(strings.TrimSuffix(strings.TrimSpace(c.Registry.URL), "/"))

	username, password, err := packconfig.ResolveAuth(packconfig.RegistryAuth{
		Registry:     destination,
		Username:     c.Registry.Username,
		Password:     c.Registry.Password,
		PasswordPath: c.Registry.PasswordPath,
		PasswordEnv:  c.Registry.PasswordEnv,
		Token:        c.Registry.Token,
		TokenPath:    c.Registry.TokenPath,
		TokenEnv:     c.Registry.TokenEnv,
	})
	if err != nil {
		return nil, err
	}

	workingDir := strings.TrimSpace(c.Settings.WorkingDir)
	valuesDir := strings.TrimSpace(c.Settings.ValuesDir)
	if valuesDir == "" {
		valuesDir = workingDir
	}

	return &spec.Spec{
		Command:        spec.CommandAirgapLoad,
		DryRun:         opts.DryRun,
		NonInteractive: opts.Yes,
		Output: spec.Output{
			Format:     opts.Output,
			EventsFile: opts.EventsFile,
		},
		AirgapLoad: &spec.AirgapLoad{
			BundlePath:  strings.TrimSpace(c.Settings.BundlePath),
			WorkingDir:  workingDir,
			ValuesDir:   valuesDir,
			Destination: destination,
			Credential: spec.RegistryCredential{
				Prefix:             destination,
				Username:           username,
				Password:           password,
				CAChainPath:        c.Registry.CAChainPath,
				Insecure:           c.Registry.Insecure,
				InsecureSkipVerify: c.Registry.InsecureSkipVerify,
			},
		},
	}, nil
}
