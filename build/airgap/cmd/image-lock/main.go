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

	"github.com/dsx-ai-factory/run-ai/build/airgap/internal/imagelock"
	"github.com/dsx-ai-factory/run-ai/build/airgap/internal/thirdparty"
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

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("image-lock", flag.ContinueOnError)
	name := flags.String("name", "runai", "")
	version := flags.String("version", "", "")
	registry := flags.String("registry", "", "")
	platformValue := flags.String("platform", "", "")
	output := flags.String("output", "", "")
	maxWorkers := flags.Int("max-workers", 10, "")
	chart := flags.String("chart", "", "")
	var roots stringsFlag
	flags.Var(&roots, "root", "")
	if err := flags.Parse(args); err != nil {
		return err
	}
	platform, err := parsePlatform(*platformValue)
	if err != nil {
		return err
	}
	if *output == "" {
		return fmt.Errorf("--output is required")
	}
	if *chart == "" {
		return fmt.Errorf("--chart is required")
	}
	thirdParty, err := thirdparty.FromChart(*chart)
	if err != nil {
		return err
	}
	outputPath := expandOutputPath(*output, *version)
	imageLock, err := imagelock.Build(ctx, imagelock.Config{
		Name:       *name,
		Version:    *version,
		Registry:   *registry,
		Platform:   platform,
		Roots:      roots,
		MaxWorkers: *maxWorkers,
		ThirdParty: thirdParty,
	}, imagelock.RemoteResolver{})
	if err != nil {
		return err
	}
	contents, err := lock.Marshal(imageLock)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(outputPath, contents, 0o644); err != nil {
		return err
	}
	fmt.Printf("Generated image lock: %s\n", outputPath)
	return nil
}

func parsePlatform(value string) (lock.Platform, error) {
	osName, architecture, found := strings.Cut(value, "/")
	if !found || osName == "" || architecture == "" || strings.Contains(architecture, "/") {
		return lock.Platform{}, fmt.Errorf("invalid platform %q, expected os/architecture", value)
	}
	return lock.Platform{OS: osName, Architecture: architecture}, nil
}

func expandOutputPath(pattern, version string) string {
	return strings.ReplaceAll(pattern, "{version}", version)
}
