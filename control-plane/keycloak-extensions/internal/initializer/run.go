/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package initializer

import "fmt"

// Paths is the set of filesystem locations Run operates on. Production code
// uses DefaultPaths — Keycloak's fixed volume layout inside the pod; tests
// substitute a temp-dir-rooted set to exercise the full flow without
// touching the real filesystem root.
type Paths struct {
	ExtensionsProvidersDir string
	ExtensionsThemesDir    string
	ExtensionsAgentDir     string

	ProvidersDir string
	ThemesDir    string
	AgentDir     string

	SecretsDir string
	ConfPath   string
}

// DefaultPaths is the fixed contract this binary runs under inside the
// keycloak-extensions container: the shared volumes it and Keycloak mount
// into the pod.
func DefaultPaths() Paths {
	return Paths{
		ExtensionsProvidersDir: "/extensions/providers",
		ExtensionsThemesDir:    "/extensions/themes",
		ExtensionsAgentDir:     "/extensions/agent",
		ProvidersDir:           "/providers",
		ThemesDir:              "/themes",
		AgentDir:               "/agent",
		SecretsDir:             "/opt/keycloak/conf/secrets",
		ConfPath:               "/conf/keycloak.conf",
	}
}

// Run stages the provider/theme/agent payloads into the shared volumes
// Keycloak reads from, then generates keycloak.conf from the mounted secret
// files — the combined behavior of the former copy-extensions.sh and
// write-keycloak-conf.sh scripts.
func Run(p Paths) error {
	if err := CopyExtensions(map[string]string{
		p.ExtensionsProvidersDir: p.ProvidersDir,
		p.ExtensionsThemesDir:    p.ThemesDir,
		p.ExtensionsAgentDir:     p.AgentDir,
	}); err != nil {
		return fmt.Errorf("stage extensions: %w", err)
	}

	if err := WriteKeycloakConf(p.SecretsDir, p.ConfPath); err != nil {
		return fmt.Errorf("write keycloak.conf: %w", err)
	}

	return nil
}
