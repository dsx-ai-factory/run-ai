/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package keycloakinit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfKey(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{name: "KC_ prefixed", in: "KC_HTTP_PORT", want: "http-port"},
		{name: "KC_ prefixed mixed case tail", in: "KC_Db_Url", want: "db-url"},
		{name: "not KC_ prefixed", in: "SOME_OTHER_VAR", want: "SOME_OTHER_VAR"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := confKey(tc.in); got != tc.want {
				t.Errorf("confKey(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestWriteKeycloakConf(t *testing.T) {
	for _, tc := range []struct {
		name    string
		secrets map[string]string
		want    string
	}{
		{
			name:    "KC_ prefixed keys are transformed",
			secrets: map[string]string{"KC_HTTP_PORT": "8080"},
			want:    "http-port=8080\n",
		},
		{
			name: "admin credentials are skipped",
			secrets: map[string]string{
				"KEYCLOAK_ADMIN":          "admin",
				"KEYCLOAK_ADMIN_PASSWORD": "secret",
				"KC_DB_URL":               "jdbc:postgresql://db/keycloak",
			},
			want: "db-url=jdbc:postgresql://db/keycloak\n",
		},
		{
			name:    "non-KC_ names pass through unchanged",
			secrets: map[string]string{"CUSTOM_FLAG": "true"},
			want:    "CUSTOM_FLAG=true\n",
		},
		{
			name:    "empty secrets dir produces an empty conf file",
			secrets: map[string]string{},
			want:    "",
		},
		{
			name:    "missing secrets dir produces an empty conf file, matching the shell script's unexpanded-glob behavior",
			secrets: nil,
			want:    "",
		},
		{
			name:    "trailing newline in a secret file is stripped, matching $(cat ...)",
			secrets: map[string]string{"KC_HOSTNAME": "runai.local\n\n"},
			want:    "hostname=runai.local\n",
		},
		{
			name: "multiple keys are written in sorted order",
			secrets: map[string]string{
				"KC_HTTP_PORT": "8080",
				"KC_DB_URL":    "jdbc",
			},
			want: "db-url=jdbc\nhttp-port=8080\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var dir string
			if tc.secrets == nil {
				dir = filepath.Join(t.TempDir(), "does-not-exist")
			} else {
				dir = t.TempDir()
			}
			for name, value := range tc.secrets {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(value), 0o600); err != nil {
					t.Fatalf("seed secret %s: %v", name, err)
				}
			}
			confPath := filepath.Join(t.TempDir(), "keycloak.conf")

			if err := WriteKeycloakConf(dir, confPath); err != nil {
				t.Fatalf("WriteKeycloakConf() error = %v", err)
			}

			got, err := os.ReadFile(confPath)
			if err != nil {
				t.Fatalf("read conf: %v", err)
			}
			if string(got) != tc.want {
				t.Errorf("conf content = %q, want %q", got, tc.want)
			}
		})
	}
}

// Reproduces the Kubernetes Secret volume layout: every key is a symlink
// into a timestamped data directory, not a regular file.
func TestWriteKeycloakConf_SymlinkedSecrets(t *testing.T) {
	dir := t.TempDir()

	dataDir := filepath.Join(dir, "..2024_01_01_00_00_00.000000000")
	mustWriteFile(t, filepath.Join(dataDir, "KC_HTTP_PORT"), "8080")

	if err := os.Symlink(dataDir, filepath.Join(dir, "..data")); err != nil {
		t.Fatalf("symlink ..data: %v", err)
	}
	if err := os.Symlink(filepath.Join("..data", "KC_HTTP_PORT"), filepath.Join(dir, "KC_HTTP_PORT")); err != nil {
		t.Fatalf("symlink KC_HTTP_PORT: %v", err)
	}

	confPath := filepath.Join(t.TempDir(), "keycloak.conf")
	if err := WriteKeycloakConf(dir, confPath); err != nil {
		t.Fatalf("WriteKeycloakConf() error = %v", err)
	}

	assertFileContent(t, confPath, "http-port=8080\n")
}
