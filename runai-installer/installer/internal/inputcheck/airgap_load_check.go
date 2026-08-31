/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inputcheck

import (
	"fmt"
	"os"
	"strings"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
)

func init() { checks = append(checks, airgapLoadCheck) }

const airgapLoadCheckName = "airgap-load"

// airgapLoadCheck validates the load settings offline. The bundle's artifact
// lock is not checked here: it travels inside the bundle, so it is only
// readable once the bundle is extracted.
func airgapLoadCheck(s *spec.Spec) []Finding {
	if s.Command != spec.CommandAirgapLoad || s.AirgapLoad == nil {
		return nil
	}
	l := s.AirgapLoad

	var findings []Finding
	findings = append(findings, bundleFindings(l)...)
	if l.WorkingDir == "" {
		findings = append(findings, loadFinding("settings.workingDir", "workingDir is required"))
	}
	findings = append(findings, destinationFindings(l)...)
	return findings
}

func bundleFindings(l *spec.AirgapLoad) []Finding {
	const field = "settings.bundlePath"
	if l.BundlePath == "" {
		return []Finding{loadFinding(field, "bundlePath is required")}
	}
	info, err := os.Stat(l.BundlePath)
	switch {
	case os.IsNotExist(err):
		return []Finding{loadFinding(field, fmt.Sprintf("%q does not exist", l.BundlePath))}
	case err != nil:
		return []Finding{loadFinding(field, fmt.Sprintf("%q cannot be read: %v", l.BundlePath, err))}
	case info.IsDir():
		return []Finding{loadFinding(field, fmt.Sprintf("%q is a directory, not a bundle", l.BundlePath))}
	}
	return nil
}

func destinationFindings(l *spec.AirgapLoad) []Finding {
	const field = "registry.url"
	if l.Destination == "" {
		return []Finding{loadFinding(field, "url is required")}
	}

	var findings []Finding
	if i := strings.Index(l.Destination, "://"); i >= 0 {
		findings = append(findings, loadFinding(field,
			fmt.Sprintf("%q must not carry a scheme: drop the %q", l.Destination, l.Destination[:i+3])))
	}
	host, _, _ := strings.Cut(l.Destination, "/")
	if host != "localhost" && !strings.ContainsAny(host, ".:") {
		findings = append(findings, loadFinding(field,
			fmt.Sprintf("%q must use a fully qualified registry host, localhost, or host:port", l.Destination)))
	}
	// Only the path can carry a tag: a colon in the host is its port.
	if _, path, found := strings.Cut(l.Destination, "/"); found && hasTagOrDigest(path) {
		findings = append(findings, loadFinding(field,
			fmt.Sprintf("%q must be a registry and optional path, without a tag or digest", l.Destination)))
	}
	return findings
}

func loadFinding(field, message string) Finding {
	return Finding{Check: airgapLoadCheckName, Field: field, Message: message}
}
