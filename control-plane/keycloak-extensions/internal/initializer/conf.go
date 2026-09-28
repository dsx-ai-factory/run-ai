/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package initializer

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// skippedSecretNames are consumed by Keycloak's bootstrap-admin bootstrap
// directly and must not be written into keycloak.conf.
var skippedSecretNames = map[string]bool{
	"KEYCLOAK_ADMIN":          true,
	"KEYCLOAK_ADMIN_PASSWORD": true,
}

// WriteKeycloakConf writes confPath as `key=value` lines, one per regular
// file under secretsDir: KC_-prefixed names become Keycloak's kebab-case
// config keys, other names pass through unchanged.
func WriteKeycloakConf(secretsDir, confPath string) error {
	// The original shell script's unexpanded `"$S"/*` glob never matches
	// `[ -f "$file" ]` either, so a missing dir isn't an error here.
	entries, err := os.ReadDir(secretsDir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read secrets dir %s: %w", secretsDir, err)
	}

	// os.Stat, not e.Type() (which uses Lstat and never follows symlinks):
	// Kubernetes Secret volumes present every key as a symlink, the same way
	// POSIX `[ -f "$file" ]` in the original script follows symlinks.
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		info, err := os.Stat(filepath.Join(secretsDir, e.Name()))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("stat secret %s: %w", e.Name(), err)
		}
		if info.Mode().IsRegular() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	var b strings.Builder
	for _, name := range names {
		if skippedSecretNames[name] {
			continue
		}
		value, err := os.ReadFile(filepath.Join(secretsDir, name))
		if err != nil {
			return fmt.Errorf("read secret %s: %w", name, err)
		}
		// bash `$(cat "$file")` strips trailing newlines; match it.
		fmt.Fprintf(&b, "%s=%s\n", confKey(name), strings.TrimRight(string(value), "\n"))
	}

	if err := os.MkdirAll(filepath.Dir(confPath), 0o755); err != nil {
		return fmt.Errorf("create parent dir for %s: %w", confPath, err)
	}
	if err := os.WriteFile(confPath, []byte(b.String()), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", confPath, err)
	}
	return nil
}

// confKey converts KC_HTTP_PORT -> http-port; other names are unchanged.
func confKey(name string) string {
	if !strings.HasPrefix(name, "KC_") {
		return name
	}
	stripped := strings.TrimPrefix(name, "KC_")
	return strings.ToLower(strings.ReplaceAll(stripped, "_", "-"))
}
