/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cli

import (
	"context"
	"os"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/artifacts"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/installconfig"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/plan"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/specrun"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/terminal"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient"
	"github.com/spf13/cobra"
)

func newInstallCmd() *cobra.Command {
	var (
		configPath       string
		dryRun           bool
		preflightOnly    bool
		skipPreflight    bool
		yes              bool
		output           string
		eventsFile       string
		statusFile       string
		onFailure        string
		kubeconfig       string
		env              string
		jfrogTokenPath   string
		ngcTokenPath     string
		devMode          bool
		executionTimeout time.Duration
	)

	cmd := &cobra.Command{
		Use:     "install",
		Aliases: []string{"install-new"},
		Short:   "Install Run:ai using the orchestrator (config-driven)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if kubeconfig != "" {
				if err := setHelmKubeconfig(kubeconfig); err != nil {
					return err
				}
			}
			cfg, err := installconfig.Load(configPath)
			if err != nil {
				return err
			}

			if err := installconfig.ResolveClusterCredentials(cfg); err != nil {
				return err
			}

			registry, err := installconfig.ResolveRegistry(cfg, jfrogTokenPath, ngcTokenPath)
			if err != nil {
				return err
			}

			opts := installconfig.RunOptions{
				DryRun:        dryRun,
				PreflightOnly: preflightOnly,
				SkipPreflight: skipPreflight,
				Yes:           yes,
				Output:        spec.OutputFormat(output),
				EventsFile:    eventsFile,
				StatusFile:    statusFile,
				OnFailure:     spec.FailurePolicy(onFailure),
				Env:           spec.Env(env),
				Registry:      registry,
				DevMode:       devMode,
			}
			s := cfg.ToSpec(opts)
			if s.UsesAirgapLoad() && !cmd.Flags().Changed("execution-timeout") {
				executionTimeout = 3 * time.Hour
			}

			if err := logging.InitFromDir(globals.LogsDir); err != nil {
				return err
			}

			term := terminal.New(terminal.WithYes(yes))

			return specrun.Execute(cmd.Context(), s, orchestrator2.Deps{}, term, specrun.ExecOptions{
				PlanBuilder:   plan.Resolve,
				ConfirmPrompt: "Proceed with installation?",
				SkipPreflight: skipPreflight,
				PreflightOnly: preflightOnly,
				OnFailureSet:  cmd.Flags().Changed("on-failure"),
				StepTimeout:   globals.StepTimeout,
				Timeout:       executionTimeout,
				DepsBuilder: func(context.Context) (orchestrator2.Deps, error) {
					k8sClients, err := k8sclient.Connect(kubeconfig)
					if err != nil {
						return orchestrator2.Deps{}, err
					}
					version := ""
					if s.ControlPlane != nil {
						version = s.ControlPlane.Version
					} else if s.Cluster != nil {
						version = s.Cluster.Version
					}
					deps, err := orchestrator2.ProductionDeps(k8sClients.GetClient(), k8sClients.GetConfig(), version)
					if err != nil {
						return orchestrator2.Deps{}, err
					}
					if s.UsesAirgapLoad() && !s.Airgap.SkipPrepare {
						credential := s.AirgapLoad.Credential
						deps.ImagePusher, err = artifacts.NewImagePusher(artifacts.Credential{
							Prefix:             credential.Prefix,
							Username:           credential.Username,
							Password:           credential.Password,
							CAChainPath:        credential.CAChainPath,
							Insecure:           credential.Insecure,
							InsecureSkipVerify: credential.InsecureSkipVerify,
						})
						if err != nil {
							return orchestrator2.Deps{}, err
						}
					}
					return deps, nil
				},
			})
		},
	}

	f := cmd.Flags()
	f.StringVar(&configPath, "config", "", "path to the installation config YAML (required)")
	f.BoolVar(&dryRun, "dry-run", false, "plan the installation without applying changes")
	f.BoolVar(&preflightOnly, "preflight-only", false, "run only preflight checks")
	f.BoolVar(&skipPreflight, "skip-preflight", false, "skip preflight checks")
	f.BoolVarP(&yes, "yes", "y", false, "non-interactive: assume yes and skip confirmation prompts")
	f.StringVar(&output, "output", "", "output format: text or json (defaults to text on a TTY, json otherwise)")
	f.StringVar(&eventsFile, "events-file", "", "write NDJSON events to this file")
	f.StringVar(&statusFile, "status-file", "", "write the final status document to this file")
	f.StringVar(&onFailure, "on-failure", "", "failure policy: abort (default) or skip. Applies to step failures when not interactive, and to input/preflight failures in any mode (when set, it overrides the interactive prompt)")
	f.StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig (overrides default discovery)")
	f.StringVar(&env, "env", "prod", "Run:ai environment selecting chart repos: prod, staging, lab, or test")
	f.StringVar(&jfrogTokenPath, "jfrog-token-path", "", "path to a file containing the JFrog token")
	f.StringVar(&ngcTokenPath, "ngc-token-path", "", "path to a file containing the NGC token")
	f.BoolVar(&devMode, "dev-mode", false, "relax input validation for development/testing (e.g. allow a Run:ai version not in the compatibility matrix)")
	f.DurationVar(&executionTimeout, "execution-timeout", 50*time.Minute, "total timeout for the run")

	_ = cmd.MarkFlagRequired("config")

	return cmd
}

func setHelmKubeconfig(path string) error {
	return os.Setenv("KUBECONFIG", path)
}
