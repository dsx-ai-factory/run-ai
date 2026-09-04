/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

type TestEnv struct {
	TempDir                string
	CertSet                *CertificateSet
	Domain                 string
	ClusterName            string
	TLSCertPath            string
	TLSKeyPath             string
	JfrogToken             string
	KubernetesDistribution string
}

func SetupTestEnv() (*TestEnv, error) {
	secretsDir := os.Getenv("RUNAI_SECRETS_DIR")
	if secretsDir == "" {
		return nil, fmt.Errorf("RUNAI_SECRETS_DIR environment variable is required")
	}

	tempDir, err := os.MkdirTemp("", "installer-e2e-*")
	if err != nil {
		return nil, err
	}

	if err := CopySecretsFromDir(secretsDir, tempDir); err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to copy secrets: %w", err)
	}

	clusterName := os.Getenv("E2E_CLUSTER_NAME")
	if clusterName == "" {
		clusterName = "installer-e2e"
	}

	jfrogToken := ""
	tokenPath := filepath.Join(tempDir, "runai_jfrog_token")
	if data, err := os.ReadFile(tokenPath); err == nil {
		jfrogToken = string(data)
	}

	env := &TestEnv{
		TempDir:                tempDir,
		Domain:                 TestDomain,
		ClusterName:            clusterName,
		TLSCertPath:            filepath.Join(tempDir, "tls.crt"),
		TLSKeyPath:             filepath.Join(tempDir, "tls.key"),
		JfrogToken:             jfrogToken,
		KubernetesDistribution: os.Getenv("E2E_KUBERNETES_DISTRIBUTION"),
	}

	if err := env.generateCerts(); err != nil {
		os.RemoveAll(tempDir)
		return nil, err
	}

	return env, nil
}

// generateCerts produces a self-signed CA + server cert valid for the test
// domain (and its wildcard) and writes them as tls.crt / tls.key /
// wildcard.crt / wildcard.key / ca.crt in TempDir. Runs unconditionally so
// e2e never depends on org-provided TLS secrets.
func (e *TestEnv) generateCerts() error {
	certSet, err := GenerateSelfSignedCerts(TestDomain, "*."+TestDomain)
	if err != nil {
		return fmt.Errorf("failed to generate self-signed certs: %w", err)
	}
	e.CertSet = certSet

	for src, dst := range map[string]string{
		certSet.ServerCertPath: "tls.crt",
		certSet.ServerKeyPath:  "tls.key",
		certSet.CACertPath:     "ca.crt",
	} {
		if err := CopyFile(src, filepath.Join(e.TempDir, dst)); err != nil {
			return fmt.Errorf("failed to copy %s: %w", dst, err)
		}
	}
	if err := CopyFile(certSet.ServerCertPath, filepath.Join(e.TempDir, "wildcard.crt")); err != nil {
		return fmt.Errorf("failed to copy wildcard cert: %w", err)
	}
	if err := CopyFile(certSet.ServerKeyPath, filepath.Join(e.TempDir, "wildcard.key")); err != nil {
		return fmt.Errorf("failed to copy wildcard key: %w", err)
	}
	return nil
}

// CAChainPath returns the path to the generated CA cert. On OpenShift the
// platform router serves its own (system-trusted) cert and any cert we install
// is ignored, so callers should treat an empty value as "use the system trust
// store".
func (e *TestEnv) CAChainPath() string {
	if e.KubernetesDistribution == "openshift" {
		return ""
	}
	return filepath.Join(e.TempDir, "ca.crt")
}

func (e *TestEnv) WildcardCertPath() string {
	return filepath.Join(e.TempDir, "wildcard.crt")
}

func (e *TestEnv) WildcardKeyPath() string {
	return filepath.Join(e.TempDir, "wildcard.key")
}

func (e *TestEnv) Cleanup() {
	if e.CertSet != nil {
		_ = e.CertSet.Cleanup()
	}
	if e.TempDir != "" {
		removeAllExceptLogs(e.TempDir)
	}
}

// removeAllExceptLogs removes all files in dir except *.log files and subdirectories.
// Log files are preserved so they can be inspected after a test run, including on failure.
// Subdirectories are preserved because they may contain diagnostic artifacts (e.g. postmortem-*/
// created by CollectDiagnostics on test failure) that need to survive until the CI artifact
// upload step runs. All sensitive flat files (certs, keys, tokens) are still deleted.
func removeAllExceptLogs(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) == ".log" {
			continue
		}
		os.RemoveAll(filepath.Join(dir, entry.Name()))
	}
}
