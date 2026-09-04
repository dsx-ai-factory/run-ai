/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cli

import (
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/artifacts"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/loadconfig"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/actions"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/plan"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/specrun"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/terminal"
	"github.com/spf13/cobra"
)

const loadLongDescription = `Unpack an air-gapped bundle, push its images to your registry, and write the
Helm values that install Run:ai from it.

The command needs no internet access: everything it installs from comes out of
the bundle. It does not touch Kubernetes — the values it writes are what you
pass to the install command, to Helm, or to your own deployment system.`

const loadExample = `  # Push a bundle to an internal registry
  installer airgap load --config airgap-load.yaml

  # Report what would be pushed without uploading anything
  installer airgap load --config airgap-load.yaml --dry-run`

func newAirgapLoadCmd() *cobra.Command {
	var (
		configPath       string
		dryRun           bool
		preflightOnly    bool
		skipPreflight    bool
		yes              bool
		output           string
		eventsFile       string
		executionTimeout time.Duration
	)

	cmd := &cobra.Command{
		Use:     "load",
		Short:   "Push an air-gapped bundle to a registry and write its Helm values",
		Long:    loadLongDescription,
		Example: loadExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadconfig.Load(configPath)
			if err != nil {
				return err
			}

			s, err := cfg.ToSpec(loadconfig.RunOptions{
				DryRun:     dryRun,
				Yes:        yes,
				Output:     spec.OutputFormat(output),
				EventsFile: eventsFile,
			})
			if err != nil {
				return err
			}

			if err := logging.InitFromDir(globals.logsDir); err != nil {
				return err
			}

			term := terminal.New(terminal.WithYes(yes))
			deps, err := loadDeps(s)
			if err != nil {
				return err
			}

			return specrun.Execute(cmd.Context(), s, deps, term, specrun.ExecOptions{
				PlanBuilder:   plan.Resolve,
				SkipPreflight: skipPreflight,
				PreflightOnly: preflightOnly,
				StepTimeout:   globals.stepTimeout,
				Timeout:       executionTimeout,
			})
		},
	}

	f := cmd.Flags()
	f.StringVar(&configPath, "config", "", "path to the load config YAML (required)")
	f.BoolVar(&dryRun, "dry-run", false, "report what would be pushed without uploading anything")
	f.BoolVar(&preflightOnly, "preflight-only", false, "run only the preflight checks")
	f.BoolVar(&skipPreflight, "skip-preflight", false, "skip the preflight checks")
	f.BoolVarP(&yes, "yes", "y", false, "non-interactive: assume yes and skip confirmation prompts")
	f.StringVar(&output, "output", "", "output format: text or json (defaults to text on a TTY, json otherwise)")
	f.StringVar(&eventsFile, "events-file", "", "write NDJSON events to this file")
	f.DurationVar(&executionTimeout, "execution-timeout", 3*time.Hour, "total timeout for the run")

	_ = cmd.MarkFlagRequired("config")

	return cmd
}

// loadDeps wires the load engine. Like the pack it touches no cluster: it reads
// the bundle, writes to one registry, and writes files.
func loadDeps(s *spec.Spec) (orchestrator2.Deps, error) {
	credential := actions.ToCredentials([]spec.RegistryCredential{s.AirgapLoad.Credential})[0]
	pusher, err := artifacts.NewImagePusher(credential)
	if err != nil {
		return orchestrator2.Deps{}, err
	}
	return orchestrator2.Deps{ImagePusher: pusher}, nil
}
