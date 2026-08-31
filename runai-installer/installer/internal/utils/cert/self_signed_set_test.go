/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cert

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsureSelfSignedCertificateSet(t *testing.T) {
	baseDir := t.TempDir()
	names := []string{"*.runai.example.com", "runai.example.com", "*.inference.runai.example.com", "runai.example.com"}

	set, err := EnsureSelfSignedCertificateSet(baseDir, names)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(baseDir, SelfSignedDirectoryName), set.OutputDir)

	leaf, err := readCertificate(set.CertPath)
	require.NoError(t, err)
	assert.Equal(t, []string{"*.inference.runai.example.com", "*.runai.example.com", "runai.example.com"}, leaf.DNSNames)

	ca, err := readCertificate(set.CACertPath)
	require.NoError(t, err)
	require.NoError(t, leaf.CheckSignatureFrom(ca))

	for _, path := range []string{set.CertPath, set.KeyPath, set.CACertPath, set.CAKeyPath} {
		info, err := os.Stat(path)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	}
	info, err := os.Stat(set.OutputDir)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o700), info.Mode().Perm())

	certBefore, err := os.ReadFile(set.CertPath)
	require.NoError(t, err)
	reused, err := EnsureSelfSignedCertificateSet(baseDir, names)
	require.NoError(t, err)
	certAfter, err := os.ReadFile(reused.CertPath)
	require.NoError(t, err)
	assert.Equal(t, certBefore, certAfter)

	_, err = EnsureSelfSignedCertificateSet(baseDir, []string{"other.example.com"})
	require.ErrorContains(t, err, "do not match requested names")
}

func TestEnsureSelfSignedCertificateSetRejectsPartialDirectory(t *testing.T) {
	baseDir := t.TempDir()
	outputDir := filepath.Join(baseDir, SelfSignedDirectoryName)
	require.NoError(t, os.Mkdir(outputDir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(outputDir, "tls.crt"), []byte("partial"), 0o600))

	_, err := EnsureSelfSignedCertificateSet(baseDir, []string{"runai.example.com"})
	require.ErrorContains(t, err, "existing self-signed certificate files are invalid")
}

func TestEnsureSelfSignedCertificateSetGeneratesInEmptyDirectory(t *testing.T) {
	baseDir := t.TempDir()
	outputDir := filepath.Join(baseDir, SelfSignedDirectoryName)
	require.NoError(t, os.Mkdir(outputDir, 0o755))

	set, err := EnsureSelfSignedCertificateSet(baseDir, []string{"runai.example.com"})

	require.NoError(t, err)
	for _, path := range []string{set.CertPath, set.KeyPath, set.CACertPath, set.CAKeyPath} {
		_, err := os.Stat(path)
		require.NoError(t, err)
	}
	info, err := os.Stat(outputDir)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o700), info.Mode().Perm())
}
