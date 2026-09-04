/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package packconfig

import (
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/pkg/artifactlock"
)

const workDirSuffix = ".work"

// RunOptions are the CLI-flag-derived knobs that map into the spec. Execution
// knobs (timeouts, preflight staging) live in specrun.ExecOptions.
type RunOptions struct {
	DryRun     bool
	Yes        bool
	Output     spec.OutputFormat
	EventsFile string
}

// ToSpec maps the config and the parsed lock into the orchestrator spec,
// applying defaults and resolving credentials.
func (c *Config) ToSpec(lock *artifactlock.ArtifactLock, opts RunOptions) (*spec.Spec, error) {
	creds, err := resolveCredentials(c.RegistryAuth)
	if err != nil {
		return nil, err
	}

	output := strings.TrimSpace(c.Settings.Output)
	if output == "" {
		output = defaultOutputPath(lock)
	}

	return &spec.Spec{
		Command:        spec.CommandAirgapPack,
		DryRun:         opts.DryRun,
		NonInteractive: opts.Yes,
		Output: spec.Output{
			Format:     opts.Output,
			EventsFile: opts.EventsFile,
		},
		AirgapPack: &spec.AirgapPack{
			Lock:           lock,
			RunaiVersion:   strings.TrimSpace(c.Settings.RunaiVersion),
			Architecture:   strings.TrimSpace(c.Settings.Architecture),
			OutputPath:     output,
			WorkDir:        output + workDirSuffix,
			Credentials:    creds,
			ImageOverrides: toOverrides(c.ImageOverrides),
		},
	}, nil
}

func defaultOutputPath(lock *artifactlock.ArtifactLock) string {
	if lock == nil || lock.Metadata.Version == "" {
		return "runai-airgap.tar"
	}
	return "runai-" + lock.Metadata.Version + "-airgap.tar"
}

func toOverrides(sources []ArtifactSource) []spec.SourceOverride {
	if len(sources) == 0 {
		return nil
	}
	out := make([]spec.SourceOverride, 0, len(sources))
	for _, s := range sources {
		out = append(out, spec.SourceOverride{
			Group:  strings.TrimSpace(s.Group),
			Name:   strings.TrimSpace(s.Name),
			Source: strings.TrimSpace(s.Source),
		})
	}
	return out
}
