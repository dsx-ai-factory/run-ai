/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package keycloakinit

import "fmt"

// Paths is the set of filesystem locations an Initializer operates on.
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

// Initializer stages the keycloak-extensions payload into Keycloak's shared
// volumes and writes keycloak.conf from the mounted secrets.
type Initializer struct {
	Paths Paths
}

// New returns an Initializer configured with DefaultPaths.
func New() *Initializer {
	return &Initializer{Paths: DefaultPaths()}
}

// Run stages the provider/theme/agent payloads into the shared volumes
// Keycloak reads from, then generates keycloak.conf from the secret files.
func (i *Initializer) Run() error {
	p := i.Paths

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
