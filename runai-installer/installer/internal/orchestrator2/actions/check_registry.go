/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
)

// RegistryTarget is one registry the check probes (read-only).
type RegistryTarget struct {
	Name               string // human label, e.g. "image registry runai.jfrog.io"
	Host               string
	Repo               string
	Project            string
	Username           string
	Password           string
	CAPath             string
	CheckPush          bool
	Insecure           bool
	InsecureSkipVerify bool
}

// CheckRegistry is a read-only preflight check that verifies the configured
// registries are reachable, the credentials authenticate and can pull, and (for
// the air-gap mirror) that push is permitted and the project exists. It never
// writes to the registry: push is verified via a token scope grant. Findings are
// surfaced as warnings only.
type CheckRegistry struct {
	// ConfigError, when set, is a problem resolving the registry config itself
	// (e.g. a malformed pull-secret token) — reported before any probing.
	ConfigError string
	Targets     []RegistryTarget
}

func (a CheckRegistry) Kind() string { return "preflight.registry" }

func (a CheckRegistry) Args() map[string]any { return map[string]any{} }

func (a CheckRegistry) Describe() string {
	return "verify registry reachability, authentication and pull/push permissions (warning only)"
}

func (a CheckRegistry) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	findings := make([]map[string]any, 0)
	add := func(target, issue string) {
		findings = append(findings, map[string]any{"registry": target, "issue": issue})
	}

	if a.ConfigError != "" {
		add("registry credentials", a.ConfigError)
	}

	for _, t := range a.Targets {
		res, err := deps.RegistryProber.Check(ctx, orchestrator2.RegistryCheckRequest{
			Host:               t.Host,
			Repo:               t.Repo,
			Project:            t.Project,
			Username:           t.Username,
			Password:           t.Password,
			CAPath:             t.CAPath,
			CheckPush:          t.CheckPush,
			Insecure:           t.Insecure,
			InsecureSkipVerify: t.InsecureSkipVerify,
		})
		switch {
		case err != nil:
			add(t.Name, "Registry check failed"+detailSuffix(err.Error()))
			continue
		case !res.Reachable:
			add(t.Name, "Registry is not reachable"+detailSuffix(res.Detail))
			continue
		case res.ResponseError != "":
			add(t.Name, "Registry returned an unexpected response"+detailSuffix(res.ResponseError))
			continue
		case !res.Authenticated:
			add(t.Name, "Invalid registry credentials")
			continue
		}
		if t.Repo != "" && !res.CanPull {
			add(t.Name, "Credentials cannot pull images from registry")
		}
		if t.Project != "" && res.ProjectChecked && !res.ProjectExists {
			add(t.Name, fmt.Sprintf("Project %q does not exist in registry", t.Project))
		}
		if t.CheckPush {
			switch {
			case !res.PushChecked:
				add(t.Name, "Could not verify push permission")
			case !res.CanPush:
				add(t.Name, "Credentials cannot push images to registry")
			}
		}
	}

	if len(findings) == 0 {
		return orchestrator2.ActionResult{}, nil
	}
	return orchestrator2.ActionResult{
		Status: orchestrator2.StatusWarn,
		Reason: fmt.Sprintf("%d registry issue(s) found", len(findings)),
		Data:   map[string]any{"registryIssues": findings},
	}, nil
}

func detailSuffix(detail string) string {
	if detail == "" {
		return ""
	}
	return ": " + detail
}
