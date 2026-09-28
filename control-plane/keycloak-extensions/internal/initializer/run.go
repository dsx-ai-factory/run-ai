/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package initializer

import "fmt"

// Paths is the set of filesystem locations Run operates on.
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

// DefaultPaths is the fixed volume layout this binary runs under in the
// keycloak-extensions container.
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
// Keycloak reads from, then generates keycloak.conf from the secret files.
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
