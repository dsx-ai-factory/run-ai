/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inputcheck

import (
	"fmt"
	"strings"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/run-ai/runai/runai-installer/installer/pkg/artifactlock"
)

func init() { checks = append(checks, airgapPackCheck) }

const airgapPackCheckName = "airgap-pack"

// airgapPackCheck validates the pack settings against the artifact lock they
// will be packed from: the requested release and architecture must be the ones
// the lock pins, and every registry and override must be actionable.
func airgapPackCheck(s *spec.Spec) []Finding {
	if s.Command != spec.CommandAirgapPack || s.AirgapPack == nil || s.AirgapPack.Lock == nil {
		return nil
	}
	p := s.AirgapPack

	var findings []Finding
	findings = append(findings, lockSelectionFindings(p)...)
	findings = append(findings, credentialFindings(p)...)
	findings = append(findings, overrideFindings(p)...)
	return findings
}

func lockSelectionFindings(p *spec.AirgapPack) []Finding {
	var findings []Finding
	if p.RunaiVersion != "" && p.RunaiVersion != p.Lock.Metadata.Version {
		findings = append(findings, packFinding("settings.runaiVersion",
			fmt.Sprintf("the artifact lock is for %s, not %s", p.Lock.Metadata.Version, p.RunaiVersion)))
	}
	if p.Architecture == "" {
		return findings
	}
	if !artifactlock.IsValidArchitecture(p.Architecture) {
		findings = append(findings, packFinding("settings.architecture",
			fmt.Sprintf("unsupported architecture %q: must be %s or %s", p.Architecture, artifactlock.ArchAMD64, artifactlock.ArchARM64)))
	} else if arch := p.Lock.Spec.Platform.Architecture; p.Architecture != arch {
		findings = append(findings, packFinding("settings.architecture",
			fmt.Sprintf("the artifact lock is for %s, not %s", arch, p.Architecture)))
	}
	return findings
}

func credentialFindings(p *spec.AirgapPack) []Finding {
	var findings []Finding
	seen := make(map[string]bool, len(p.Credentials))
	for i, c := range p.Credentials {
		field := fmt.Sprintf("registryAuth[%d].registry", i)
		switch {
		case c.Prefix == "":
			findings = append(findings, packFinding(field, "registry is required"))
		case seen[c.Prefix]:
			findings = append(findings, packFinding(field, fmt.Sprintf("duplicate registry %q", c.Prefix)))
		default:
			seen[c.Prefix] = true
		}
	}
	return findings
}

func overrideFindings(p *spec.AirgapPack) []Finding {
	var findings []Finding
	seen := make(map[string]bool, len(p.ImageOverrides))
	for i, o := range p.ImageOverrides {
		key := o.Name
		if o.Group != "" {
			key = o.Group + "/" + o.Name
		}
		if o.Name == "" {
			findings = append(findings, packFinding(fmt.Sprintf("imageOverrides[%d].name", i), "name is required"))
		} else if seen[key] {
			findings = append(findings, packFinding(fmt.Sprintf("imageOverrides[%d].name", i), fmt.Sprintf("duplicate override %q", key)))
		} else {
			seen[key] = true
			groups := imageGroups(p.Lock, o.Name)
			if o.Group != "" && !containsGroup(groups, o.Group) {
				findings = append(findings, packFinding(fmt.Sprintf("imageOverrides[%d].group", i),
					fmt.Sprintf("image %q does not exist in group %q", o.Name, o.Group)))
			} else if o.Group == "" && len(groups) > 1 {
				findings = append(findings, packFinding(fmt.Sprintf("imageOverrides[%d].name", i),
					fmt.Sprintf("image %q exists in groups %s: set group", o.Name, strings.Join(groups, ", "))))
			}
		}

		field := fmt.Sprintf("imageOverrides[%d].source", i)
		if o.Source == "" {
			findings = append(findings, packFinding(field, "source is required"))
		} else if !hasTagOrDigest(o.Source) {
			findings = append(findings, packFinding(field,
				fmt.Sprintf("%q must carry an explicit tag or digest", o.Source)))
		}
	}
	return findings
}

func imageGroups(lock *artifactlock.ArtifactLock, name string) []string {
	var groups []string
	for _, group := range lock.Spec.ImageLocks {
		for _, image := range group.Images {
			if image.Name == name {
				groups = append(groups, group.Name)
			}
		}
	}
	return groups
}

func containsGroup(groups []string, group string) bool {
	for _, candidate := range groups {
		if candidate == group {
			return true
		}
	}
	return false
}

// hasTagOrDigest reports whether a reference pins a tag or digest, ignoring any
// port in the registry host.
func hasTagOrDigest(ref string) bool {
	name := ref[strings.LastIndex(ref, "/")+1:]
	return strings.ContainsAny(name, ":@")
}

func packFinding(field, message string) Finding {
	return Finding{Check: airgapPackCheckName, Field: field, Message: message}
}
