/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	airgaplock "github.com/dsx-ai-factory/run-ai/build/airgap/internal/artifactlock"
	"github.com/dsx-ai-factory/run-ai/build/airgap/pkg/lock"
)

type stringsFlag []string

func (values *stringsFlag) String() string {
	return strings.Join(*values, ",")
}

func (values *stringsFlag) Set(value string) error {
	*values = append(*values, value)
	return nil
}

type assignmentsFlag map[string]string

func (values assignmentsFlag) String() string {
	return ""
}

func (values assignmentsFlag) Set(value string) error {
	name, resolved, found := strings.Cut(value, "=")
	if !found || strings.TrimSpace(name) == "" || strings.TrimSpace(resolved) == "" {
		return fmt.Errorf("expected name=value, got %q", value)
	}
	if _, exists := values[name]; exists {
		return fmt.Errorf("duplicate value for %q", name)
	}
	values[name] = resolved
	return nil
}

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("artifact-lock", flag.ContinueOnError)
	name := flags.String("name", "runai", "")
	version := flags.String("version", "", "")
	helmUsername := flags.String("helm-username", "$oauthtoken", "")
	helmPasswordEnv := flags.String("helm-password-env", "NGC_API_KEY", "")
	helmBinary := flags.String("helm-binary", "helm", "")
	resourceBaseURL := flags.String("ngc-resource-base-url", "https://api.ngc.nvidia.com/v2/org/nvidia/team/runai/resources", "")
	outputDirectory := flags.String("output-dir", "", "")
	var imageLocks stringsFlag
	var chartSources stringsFlag
	var chartVersions stringsFlag
	var cliRoots stringsFlag
	cliVersions := assignmentsFlag{}
	cliSources := assignmentsFlag{}
	cliUsernames := assignmentsFlag{}
	cliPasswordEnvs := assignmentsFlag{}
	cliTokenEnvs := assignmentsFlag{}
	flags.Var(&imageLocks, "image-lock", "")
	flags.Var(&chartSources, "chart-url", "")
	flags.Var(&chartVersions, "chart-version", "")
	flags.Var(&cliRoots, "cli-root", "")
	flags.Var(cliVersions, "cli-version", "")
	flags.Var(cliSources, "cli-source", "")
	flags.Var(cliUsernames, "cli-username", "")
	flags.Var(cliPasswordEnvs, "cli-password-env", "")
	flags.Var(cliTokenEnvs, "cli-token-env", "")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *outputDirectory == "" {
		return fmt.Errorf("--output-dir is required")
	}
	password := os.Getenv(*helmPasswordEnv)
	if password == "" {
		return fmt.Errorf("%s must be set when pulling published artifacts", *helmPasswordEnv)
	}
	cliSourceConfigs := map[string]airgaplock.CLISource{}
	cliNames := map[string]bool{}
	for _, values := range []assignmentsFlag{cliVersions, cliSources, cliUsernames, cliPasswordEnvs, cliTokenEnvs} {
		for name := range values {
			cliNames[name] = true
		}
	}
	for name := range cliNames {
		config := cliSourceConfigs[name]
		config.URL = cliSources[name]
		config.Version = cliVersions[name]
		config.Credentials.Username = cliUsernames[name]
		if env := cliPasswordEnvs[name]; env != "" {
			config.Credentials.Password = os.Getenv(env)
			if config.Credentials.Password == "" {
				return fmt.Errorf("%s must be set for CLI %s", env, name)
			}
		}
		if env := cliTokenEnvs[name]; env != "" {
			config.Credentials.BearerToken = os.Getenv(env)
			if config.Credentials.BearerToken == "" {
				return fmt.Errorf("%s must be set for CLI %s", env, name)
			}
		}
		if config.Credentials.BearerToken != "" && config.Credentials.Username != "" {
			return fmt.Errorf("CLI %s cannot use basic and bearer authentication together", name)
		}
		cliSourceConfigs[name] = config
	}
	artifactLock, err := airgaplock.Build(ctx, airgaplock.Config{
		Name:               *name,
		Version:            *version,
		ImageLocks:         imageLocks,
		ChartSources:       chartSources,
		ChartVersions:      chartVersions,
		CLIRoots:           cliRoots,
		CLISources:         cliSourceConfigs,
		HelmUsername:       *helmUsername,
		HelmPassword:       password,
		NGCResourceBaseURL: *resourceBaseURL,
		NGCToken:           password,
	}, airgaplock.Dependencies{
		HelmRunner: airgaplock.ExecHelmRunner{Binary: *helmBinary},
		Downloader: airgaplock.HTTPDownloader{},
	})
	if err != nil {
		return err
	}
	contents, err := lock.Marshal(artifactLock)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(*outputDirectory, 0o755); err != nil {
		return err
	}
	filename := fmt.Sprintf(
		"%s-%s-%s-%s-%s.yaml",
		artifactLock.Metadata.Name,
		artifactLock.Metadata.Version,
		artifactLock.Spec.Profile,
		artifactLock.Spec.Platform.OS,
		artifactLock.Spec.Platform.Architecture,
	)
	output := filepath.Join(*outputDirectory, filename)
	if err := os.WriteFile(output, contents, 0o644); err != nil {
		return err
	}
	fmt.Printf("Generated artifact lock: %s\n", output)
	return nil
}
