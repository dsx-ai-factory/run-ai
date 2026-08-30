package cert

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const SelfSignedDirectoryName = "runai_certs"

type SelfSignedCertificateSet struct {
	OutputDir  string
	CertPath   string
	KeyPath    string
	CACertPath string
	CAKeyPath  string
}

func EnsureSelfSignedCertificateSet(baseDir string, dnsNames []string) (*SelfSignedCertificateSet, error) {
	names := normalizeDNSNames(dnsNames)
	if len(names) == 0 {
		return nil, errors.New("at least one DNS name is required")
	}
	outputDir := filepath.Join(baseDir, SelfSignedDirectoryName)
	set := certificateSetAt(outputDir)
	if _, err := os.Stat(outputDir); err == nil {
		entries, err := os.ReadDir(outputDir)
		if err != nil {
			return nil, err
		}
		if len(entries) > 0 {
			if err := validateCertificateSet(set, names); err != nil {
				return nil, fmt.Errorf("existing self-signed certificate files are invalid: %w", err)
			}
			return set, nil
		}
		if err := os.Remove(outputDir); err != nil {
			return nil, err
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	tempDir, err := os.MkdirTemp(baseDir, ".runai_certs-")
	if err != nil {
		return nil, err
	}
	removeTemp := true
	defer func() {
		if removeTemp {
			_ = os.RemoveAll(tempDir)
		}
	}()

	tempSet := certificateSetAt(tempDir)
	if err := generateCertificateSet(tempSet, names); err != nil {
		return nil, err
	}
	if err := validateCertificateSet(tempSet, names); err != nil {
		return nil, err
	}
	if err := os.Rename(tempDir, outputDir); err != nil {
		return nil, err
	}
	removeTemp = false
	return set, nil
}

func certificateSetAt(outputDir string) *SelfSignedCertificateSet {
	return &SelfSignedCertificateSet{
		OutputDir:  outputDir,
		CertPath:   filepath.Join(outputDir, "tls.crt"),
		KeyPath:    filepath.Join(outputDir, "tls.key"),
		CACertPath: filepath.Join(outputDir, "ca.crt"),
		CAKeyPath:  filepath.Join(outputDir, "ca.key"),
	}
}
