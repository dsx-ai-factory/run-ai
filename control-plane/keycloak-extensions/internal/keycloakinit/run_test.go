/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package keycloakinit

import (
	"path/filepath"
	"testing"
)

func TestInitializer_Run(t *testing.T) {
	root := t.TempDir()

	paths := Paths{
		ExtensionsProvidersDir: filepath.Join(root, "extensions", "providers"),
		ExtensionsThemesDir:    filepath.Join(root, "extensions", "themes"),
		ExtensionsAgentDir:     filepath.Join(root, "extensions", "agent"),
		ProvidersDir:           filepath.Join(root, "providers"),
		ThemesDir:              filepath.Join(root, "themes"),
		AgentDir:               filepath.Join(root, "agent"),
		SecretsDir:             filepath.Join(root, "secrets"),
		ConfPath:               filepath.Join(root, "conf", "keycloak.conf"),
	}

	mustWriteFile(t, filepath.Join(paths.ExtensionsProvidersDir, "runai-custom-providers.jar"), "jar")
	mustWriteFile(t, filepath.Join(paths.ExtensionsThemesDir, "runai-theme", "login", "theme.ftl"), "theme")
	mustWriteFile(t, filepath.Join(paths.ExtensionsAgentDir, "opentelemetry-javaagent.jar"), "agent")
	mustWriteFile(t, filepath.Join(paths.SecretsDir, "KC_HTTP_PORT"), "8080")
	mustWriteFile(t, filepath.Join(paths.SecretsDir, "KEYCLOAK_ADMIN"), "admin")

	init := &Initializer{Paths: paths}
	if err := init.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	assertFileContent(t, filepath.Join(paths.ProvidersDir, "runai-custom-providers.jar"), "jar")
	assertFileContent(t, filepath.Join(paths.ThemesDir, "runai-theme", "login", "theme.ftl"), "theme")
	assertFileContent(t, filepath.Join(paths.AgentDir, "opentelemetry-javaagent.jar"), "agent")
	assertFileContent(t, paths.ConfPath, "http-port=8080\n")
}
