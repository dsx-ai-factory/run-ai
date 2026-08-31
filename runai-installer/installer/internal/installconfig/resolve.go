/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package installconfig

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
)

// registryAuthMode is the credential scheme a registry config resolves to.
type registryAuthMode int

const (
	authNone registryAuthMode = iota
	authToken
	authBasic
)

// ResolveRegistry picks exactly one registry source and resolves its
// credentials into a spec.Registry. Precedence: the --jfrog-token-path /
// --ngc-token-path CLI flags override the config file's registry section.
func ResolveRegistry(cfg *Config, jfrogTokenPath, ngcTokenPath string) (*spec.Registry, error) {
	if jfrogTokenPath != "" && ngcTokenPath != "" {
		return nil, errors.New("--jfrog-token-path and --ngc-token-path are mutually exclusive")
	}

	if jfrogTokenPath != "" || ngcTokenPath != "" {
		typ, path := spec.RegistryTypeJFrog, jfrogTokenPath
		if ngcTokenPath != "" {
			typ, path = spec.RegistryTypeNGC, ngcTokenPath
		}
		token, err := readSecretFile(path)
		if err != nil {
			return nil, err
		}
		reg := &spec.Registry{Type: typ, Token: token}
		if cfg != nil && cfg.Registry != nil {
			reg.Host = cfg.Registry.Host
			reg.CAChainPath = cfg.Registry.CAChainPath
			reg.Insecure = cfg.Registry.Insecure
			reg.InsecureSkipVerify = cfg.Registry.InsecureSkipVerify
			reg.ChartRepository = cfg.Registry.ChartRepository
		}
		return reg, nil
	}

	if cfg == nil || cfg.Registry == nil {
		return nil, errors.New("no registry credentials: set --jfrog-token-path, --ngc-token-path, or registry in the config")
	}
	return resolveRegistryFromConfig(cfg.Registry)
}

// resolveRegistryFromConfig validates the credential combination and reads any
// file-based secrets into a spec.Registry.
func resolveRegistryFromConfig(r *Registry) (*spec.Registry, error) {
	mode, err := validateRegistryAuth(r)
	if err != nil {
		return nil, err
	}

	out := &spec.Registry{
		Type:               spec.RegistryType(r.Type),
		Host:               r.Host,
		CAChainPath:        r.CAChainPath,
		Insecure:           r.Insecure,
		InsecureSkipVerify: r.InsecureSkipVerify,
		ChartRepository:    r.ChartRepository,
	}

	switch mode {
	case authToken:
		if err := validateRegistryType(out.Type); err != nil {
			return nil, err
		}
		if r.TokenPath != "" {
			if out.Token, err = readSecretFile(r.TokenPath); err != nil {
				return nil, err
			}
		} else {
			out.Token = r.Token
		}
	case authBasic:
		out.Username = r.Username
		if r.PasswordPath != "" {
			if out.Password, err = readSecretFile(r.PasswordPath); err != nil {
				return nil, err
			}
		} else {
			out.Password = r.Password
		}
	case authNone:
		// Unauthenticated registry (e.g. an open mirror): host only.
	}
	return out, nil
}

// validateRegistryAuth checks that the registry credential fields form exactly
// one valid auth scheme and reports which one. It performs no file I/O so it is
// safe to call during in-memory input validation.
func validateRegistryAuth(r *Registry) (registryAuthMode, error) {
	hasUser := r.Username != ""
	hasPassword := r.Password != ""
	hasPasswordPath := r.PasswordPath != ""
	hasToken := r.Token != ""
	hasTokenPath := r.TokenPath != ""

	if hasPassword && hasPasswordPath {
		return authNone, errors.New("registry: set either password or passwordPath, not both")
	}
	if hasToken && hasTokenPath {
		return authNone, errors.New("registry: set either token or tokenPath, not both")
	}

	usesBasic := hasUser || hasPassword || hasPasswordPath
	usesToken := hasToken || hasTokenPath

	switch {
	case usesBasic && usesToken:
		return authNone, errors.New("registry: basic auth (username/password) and token auth are mutually exclusive")
	case usesBasic:
		if !hasUser {
			return authNone, errors.New("registry: a password requires a username")
		}
		if !hasPassword && !hasPasswordPath {
			return authNone, errors.New("registry: a username requires a password or passwordPath")
		}
		return authBasic, nil
	case usesToken:
		return authToken, nil
	default:
		return authNone, nil
	}
}

func validateRegistryType(t spec.RegistryType) error {
	switch t {
	case spec.RegistryTypeJFrog, spec.RegistryTypeNGC:
		return nil
	}
	return fmt.Errorf("unknown registry type: %q", t)
}

func readSecretFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read registry secret %q: %w", path, err)
	}
	secret := strings.TrimSpace(string(b))
	if secret == "" {
		return "", fmt.Errorf("registry secret file %q is empty", path)
	}
	return secret, nil
}
