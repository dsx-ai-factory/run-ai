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

// skippedSecretNames are secret files Keycloak's bootstrap-admin bootstrap
// consumes directly; they must not be written into keycloak.conf.
var skippedSecretNames = map[string]bool{
	"KEYCLOAK_ADMIN":          true,
	"KEYCLOAK_ADMIN_PASSWORD": true,
}

// WriteKeycloakConf reads every regular file directly under secretsDir and
// writes confPath as `key=value` lines, one per file: KC_-prefixed names
// become Keycloak's kebab-case config keys, other names pass through
// unchanged, and the bootstrap admin credentials are skipped.
func WriteKeycloakConf(secretsDir, confPath string) error {
	// A missing secrets dir is not an error: the original shell script's
	// unexpanded `"$S"/*` glob never matches `[ -f "$file" ]` either, so it
	// silently produced an empty conf file in that case too.
	entries, err := os.ReadDir(secretsDir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read secrets dir %s: %w", secretsDir, err)
	}

	// Stat (not e.Type(), which uses Lstat and never follows symlinks) each
	// entry: Kubernetes Secret volumes present every key as a symlink into a
	// timestamped data directory, the same way POSIX `[ -f "$file" ]` (used
	// by the original script) follows symlinks to test the resolved target.
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		info, err := os.Stat(filepath.Join(secretsDir, e.Name()))
		if err != nil {
			if os.IsNotExist(err) {
				continue // dangling symlink; not a secret key
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
		// Matches bash `$(cat "$file")`, which strips trailing newlines.
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

// confKey converts a KC_-prefixed secret name into Keycloak's kebab-case
// config key (KC_HTTP_PORT -> http-port); other names are left unchanged.
func confKey(name string) string {
	if !strings.HasPrefix(name, "KC_") {
		return name
	}
	stripped := strings.TrimPrefix(name, "KC_")
	return strings.ToLower(strings.ReplaceAll(stripped, "_", "-"))
}
