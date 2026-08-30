package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/run-ai/runai/runai-common-packages/logging"
	shared "github.com/run-ai/runai/runai-installer/installer/e2e/utils"
	"github.com/run-ai/runai/runai-installer/provisioner-service/pkg/config"
)

type TestEnv struct {
	TempDir string
	CertSet *shared.CertificateSet
}

func SetupTestEnv() (*TestEnv, error) {
	secretsDir := os.Getenv("RUNAI_SECRETS_DIR")
	if secretsDir == "" {
		return nil, fmt.Errorf("RUNAI_SECRETS_DIR environment variable is required")
	}

	tempDir, err := os.MkdirTemp("", "provisioner-e2e-*")
	if err != nil {
		return nil, err
	}

	env := &TestEnv{
		TempDir: tempDir,
	}

	if err := CopySecretsFromDir(secretsDir, tempDir); err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to copy secrets: %w", err)
	}

	os.Setenv("SECRETS_MOUNT_PATH", tempDir)
	os.Setenv("RUNAI_TLS_CERT_PATH", filepath.Join(tempDir, "tls.crt"))
	os.Setenv("RUNAI_TLS_KEY_PATH", filepath.Join(tempDir, "tls.key"))
	os.Setenv("RUNAI_WILDCARD_TLS_CERT_PATH", filepath.Join(tempDir, "wildcard.crt"))
	os.Setenv("RUNAI_WILDCARD_TLS_KEY_PATH", filepath.Join(tempDir, "wildcard.key"))

	logging.Init(nil)

	if err := env.EnableSelfSignedCA(); err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to generate self-signed certs: %w", err)
	}

	if TestK8sDistribution == "openshift" {
		os.Unsetenv("RUNAI_CA_CHAIN_PATH")
		config.Init(tempDir)
	}

	return env, nil
}

func (e *TestEnv) Cleanup() {
	if e.CertSet != nil {
		_ = e.CertSet.Cleanup()
	}
	if e.TempDir != "" {
		os.RemoveAll(e.TempDir)
	}

	os.Unsetenv("SECRETS_MOUNT_PATH")
	os.Unsetenv("RUNAI_TLS_CERT_PATH")
	os.Unsetenv("RUNAI_TLS_KEY_PATH")
	os.Unsetenv("RUNAI_CA_CHAIN_PATH")
	os.Unsetenv("RUNAI_WILDCARD_TLS_CERT_PATH")
	os.Unsetenv("RUNAI_WILDCARD_TLS_KEY_PATH")
}

func (e *TestEnv) EnableSelfSignedCA() error {
	certSet, err := shared.GenerateSelfSignedCerts(TestDomain, TestWildcardDomain)
	if err != nil {
		return fmt.Errorf("failed to generate self-signed certs: %w", err)
	}
	e.CertSet = certSet

	if err := shared.CopyFile(certSet.ServerCertPath, filepath.Join(e.TempDir, "tls.crt")); err != nil {
		return fmt.Errorf("failed to copy server cert: %w", err)
	}
	if err := shared.CopyFile(certSet.ServerKeyPath, filepath.Join(e.TempDir, "tls.key")); err != nil {
		return fmt.Errorf("failed to copy server key: %w", err)
	}
	if err := shared.CopyFile(certSet.CACertPath, filepath.Join(e.TempDir, "ca.crt")); err != nil {
		return fmt.Errorf("failed to copy CA cert: %w", err)
	}
	if err := shared.CopyFile(certSet.ServerCertPath, filepath.Join(e.TempDir, "wildcard.crt")); err != nil {
		return fmt.Errorf("failed to copy wildcard cert: %w", err)
	}
	if err := shared.CopyFile(certSet.ServerKeyPath, filepath.Join(e.TempDir, "wildcard.key")); err != nil {
		return fmt.Errorf("failed to copy wildcard key: %w", err)
	}

	os.Setenv("RUNAI_CA_CHAIN_PATH", filepath.Join(e.TempDir, "ca.crt"))
	config.Init(e.TempDir)

	return nil
}

// DisableSelfSignedCA is kept for backward compatibility but is now a no-op:
// self-signed certs are generated unconditionally in SetupTestEnv and shared
// by every spec, so disabling them between tests would break subsequent specs.
func (e *TestEnv) DisableSelfSignedCA() {}
