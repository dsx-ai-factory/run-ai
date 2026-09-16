/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cli

import (
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/artifacts"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/actions"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/plan"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/packconfig"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/releaselock"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/specrun"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/terminal"
	"github.com/spf13/cobra"
)

const packLongDescription = `Download every artifact of a Run:ai release and write them to a single
air-gapped bundle.

Two files drive the command. The artifact lock, published by Run:ai, pins the
exact images and charts a release is made of. The config supplies what only you
can know: the credentials for your registries, and any images you want to
substitute or add.

The lock is fetched for you when --airgap-lock-file is not set: the command
looks up the GitHub release of the version in settings.runaiVersion (an exact
tag such as 3.0.5, or a line such as 3.0 for its newest patch) and downloads
the lock for settings.architecture (amd64 unless set). A lock file you provide
is always used as-is.`

const packExample = `  # Build a bundle for the newest 3.0 release, fetching its lock from GitHub
  installer airgap pack --config airgap-pack.yaml

  # Build a bundle from a lock you already have
  installer airgap pack --config airgap-pack.yaml --airgap-lock-file runai-3.0.5-standard-linux-amd64.yaml

  # Report what would be packed without downloading anything
  installer airgap pack --config airgap-pack.yaml --dry-run`

func newAirgapPackCmd() *cobra.Command {
	var (
		configPath       string
		lockFilePath     string
		dryRun           bool
		preflightOnly    bool
		skipPreflight    bool
		yes              bool
		output           string
		eventsFile       string
		executionTimeout time.Duration
	)

	cmd := &cobra.Command{
		Use:     "pack",
		Short:   "Build an air-gapped bundle from a release's artifacts",
		Long:    packLongDescription,
		Example: packExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := packconfig.Load(configPath)
			if err != nil {
				return err
			}

			term := terminal.New(terminal.WithYes(yes))

			lock, source, err := cfg.ResolveLock(cmd.Context(), lockFilePath, releaselock.NewFetcher())
			if err != nil {
				return err
			}
			if source.Tag != "" {
				logging.Infof("[Pack] Using artifact lock %s from release %s (%s)", source.Asset, source.Tag, source.URL)
				if textOutput(output, term) {
					term.Info("Using artifact lock %s from release %s (%s)", source.Asset, source.Tag, source.URL)
				}
			}

			s, err := cfg.ToSpec(lock, packconfig.RunOptions{
				DryRun:     dryRun,
				Yes:        yes,
				Output:     spec.OutputFormat(output),
				EventsFile: eventsFile,
			})
			if err != nil {
				return err
			}

			if err := logging.InitFromDir(globals.LogsDir); err != nil {
				return err
			}

			return specrun.Execute(cmd.Context(), s, packDeps(s), term, specrun.ExecOptions{
				PlanBuilder:   plan.Resolve,
				SkipPreflight: skipPreflight,
				PreflightOnly: preflightOnly,
				StepTimeout:   globals.StepTimeout,
				Timeout:       executionTimeout,
			})
		},
	}

	f := cmd.Flags()
	f.StringVar(&configPath, "config", "", "path to the pack config YAML (required)")
	f.StringVar(&lockFilePath, "airgap-lock-file", "", "path to the release's artifact lock (fetched from the release on GitHub when not set)")
	f.BoolVar(&dryRun, "dry-run", false, "report what would be packed without downloading anything")
	f.BoolVar(&preflightOnly, "preflight-only", false, "run only the preflight checks")
	f.BoolVar(&skipPreflight, "skip-preflight", false, "skip the preflight checks")
	f.BoolVarP(&yes, "yes", "y", false, "non-interactive: assume yes and skip confirmation prompts")
	f.StringVar(&output, "output", "", "output format: text or json (defaults to text on a TTY, json otherwise)")
	f.StringVar(&eventsFile, "events-file", "", "write NDJSON events to this file")
	f.DurationVar(&executionTimeout, "execution-timeout", 120*time.Minute, "total timeout for the run")

	_ = cmd.MarkFlagRequired("config")

	return cmd
}

func textOutput(output string, term *terminal.Terminal) bool {
	switch spec.OutputFormat(output) {
	case spec.OutputFormatText:
		return true
	case spec.OutputFormatNone:
		return term.IsTTY()
	}
	return false
}

// packDeps wires the pack engine. Unlike the install commands it touches no
// cluster: the pack only reads registries and writes files.
func packDeps(s *spec.Spec) orchestrator2.Deps {
	keychain := artifacts.NewKeychain(actions.ToCredentials(s.AirgapPack.Credentials))
	return orchestrator2.Deps{
		RegistryProber: orchestrator2.NewRegistryProber(),
		ImagePuller:    artifacts.NewImagePuller(keychain),
		ChartFetcher:   artifacts.NewChartFetcher(keychain),
		CLIFetcher:     artifacts.NewCLIFetcher(keychain),
	}
}
