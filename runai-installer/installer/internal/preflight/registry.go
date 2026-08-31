/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package preflight

import (
	"encoding/base64"
	"encoding/json"
	"sort"
	"strings"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/actions"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/steps"
)

func init() { builders = append(builders, registryBuilder) }

// registryBuilder assembles the registry check. It needs a RegistryProber, so it
// returns nil when one is not wired (e.g. component tests not exercising it).
func registryBuilder(s *spec.Spec, deps orchestrator2.Deps) orchestrator2.Step {
	if deps.RegistryProber == nil || s.Registry == nil {
		return nil
	}
	targets, configErr := registryTargets(s)
	if len(targets) == 0 && configErr == "" {
		return nil
	}
	return steps.NewActionStepTitled("preflight-registry", "Registry access", actions.CheckRegistry{Targets: targets, ConfigError: configErr})
}

// registryTargets resolves the registries to probe from the spec. For air-gap it
// is the customer mirror (pull + push + project). For a connected install it is
// every image registry embedded in the pull-secret token (auth + pull). The
// second return is a config error (e.g. a malformed token) surfaced as a warning
// rather than silently producing no check.
func registryTargets(s *spec.Spec) ([]actions.RegistryTarget, string) {
	ca := s.Registry.CAChainPath

	if s.Airgap != nil {
		host, project := splitHostProject(s.Registry.Host)
		if host == "" {
			return nil, "Air-gap registry host is not configured"
		}
		repo := "runai-preflight-probe"
		if project != "" {
			repo = project + "/runai-preflight-probe"
		}
		return []actions.RegistryTarget{{
			Name:               "air-gap mirror " + s.Registry.Host,
			Host:               host,
			Project:            project,
			Repo:               repo,
			Username:           s.Registry.Username,
			Password:           s.Registry.Password,
			CAPath:             ca,
			CheckPush:          true,
			Insecure:           s.Registry.Insecure,
			InsecureSkipVerify: s.Registry.InsecureSkipVerify,
		}}, ""
	}

	dockerConfig, ok := s.Registry.ImagePullConfig()
	if !ok {
		return nil, ""
	}
	return connectedTargets(dockerConfig, s.Registry, ca)
}

// connectedTargets resolves image registry probes from a pull secret.
func connectedTargets(dockerConfig string, reg *spec.Registry, ca string) ([]actions.RegistryTarget, string) {
	raw, err := base64.StdEncoding.DecodeString(dockerConfig)
	if err != nil {
		return nil, "Registry token is not valid base64"
	}
	var cfg struct {
		Auths map[string]struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Auth     string `json:"auth"`
		} `json:"auths"`
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, "Registry token is not a valid docker config (.dockerconfigjson)"
	}
	if len(cfg.Auths) == 0 {
		return nil, "Registry token contains no credentials"
	}

	hosts := make([]string, 0, len(cfg.Auths))
	for host := range cfg.Auths {
		hosts = append(hosts, host)
	}
	sort.Strings(hosts)

	targets := make([]actions.RegistryTarget, 0, len(hosts))
	for _, host := range hosts {
		a := cfg.Auths[host]
		user, pass := a.Username, a.Password
		if user == "" && a.Auth != "" {
			if dec, err := base64.StdEncoding.DecodeString(a.Auth); err == nil {
				if i := strings.IndexByte(string(dec), ':'); i >= 0 {
					user, pass = string(dec)[:i], string(dec)[i+1:]
				}
			}
		}
		targets = append(targets, actions.RegistryTarget{
			Name:               "image registry " + hostOnly(host),
			Host:               hostOnly(host),
			Username:           user,
			Password:           pass,
			CAPath:             ca,
			Insecure:           reg.Insecure,
			InsecureSkipVerify: reg.InsecureSkipVerify,
		})
	}
	return targets, ""
}

// splitHostProject splits a registry reference like "harbor.corp.com/runai" into
// host ("harbor.corp.com") and the first path segment as the project ("runai").
func splitHostProject(ref string) (host, project string) {
	ref = hostAndPath(ref)
	parts := strings.SplitN(ref, "/", 2)
	host = parts[0]
	if len(parts) == 2 {
		project = strings.SplitN(parts[1], "/", 2)[0]
	}
	return host, project
}

// hostOnly returns just the host of a registry reference (no scheme, no path).
func hostOnly(ref string) string {
	return strings.SplitN(hostAndPath(ref), "/", 2)[0]
}

func hostAndPath(ref string) string {
	if i := strings.Index(ref, "://"); i >= 0 {
		ref = ref[i+3:]
	}
	return strings.TrimSuffix(ref, "/")
}
