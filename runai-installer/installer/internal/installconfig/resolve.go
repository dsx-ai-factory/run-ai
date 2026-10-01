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

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
)

// registryAuthMode is the credential scheme a registry config resolves to.
type registryAuthMode int

const (
	authNone registryAuthMode = iota
	authToken
	authBasic
)

// ResolveRegistry picks exactly one registry source and resolves its
// credentials into a spec.Registry. Precedence: the --registry-token-path /
// --ngc-token-path CLI flags override the config file's registry section. With
// neither, it falls back to anonymous NGC, which serves the public Run:ai OSS
// charts and images.
func ResolveRegistry(cfg *Config, registryTokenPath, ngcTokenPath string) (*spec.Registry, error) {
	if registryTokenPath != "" && ngcTokenPath != "" {
		return nil, errors.New("--registry-token-path and --ngc-token-path are mutually exclusive")
	}

	if registryTokenPath != "" || ngcTokenPath != "" {
		typ, path := spec.RegistryTypeDockerConfig, registryTokenPath
		if ngcTokenPath != "" {
			typ, path = spec.RegistryTypeNGC, ngcTokenPath
		}
		token, err := readSecretFile(path, "registry secret")
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
		return &spec.Registry{Type: spec.RegistryTypeNGC}, nil
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
			if out.Token, err = readSecretFile(r.TokenPath, "registry secret"); err != nil {
				return nil, err
			}
		} else {
			out.Token = r.Token
		}
	case authBasic:
		out.Username = r.Username
		if r.PasswordPath != "" {
			if out.Password, err = readSecretFile(r.PasswordPath, "registry secret"); err != nil {
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
	case spec.RegistryTypeDockerConfig, spec.RegistryTypeNGC:
		return nil
	}
	return fmt.Errorf("unknown registry type: %q", t)
}

func ResolveClusterCredentials(cfg *Config) error {
	if cfg == nil || cfg.Cluster == nil {
		return nil
	}
	c := cfg.Cluster
	if c.BootstrapTokenPath != "" && c.ClientSecretPath != "" {
		return errors.New("cluster: bootstrapTokenPath and clientSecretPath are mutually exclusive")
	}
	var err error
	if c.BootstrapTokenPath != "" {
		if c.bootstrapToken, err = readSecretFile(c.BootstrapTokenPath, "cluster bootstrap token"); err != nil {
			return err
		}
	}
	if c.ClientSecretPath != "" {
		if c.clientSecret, err = readSecretFile(c.ClientSecretPath, "cluster client secret"); err != nil {
			return err
		}
	}
	return resolveControlPlaneCredentials(c.ControlPlaneCredentials)
}

func resolveControlPlaneCredentials(cc *ControlPlaneCredentials) error {
	if cc == nil {
		return nil
	}
	if cc.Password != "" && cc.PasswordPath != "" {
		return errors.New("cluster.controlPlaneCredentials: set either password or passwordPath, not both")
	}
	if cc.Token != "" && cc.TokenPath != "" {
		return errors.New("cluster.controlPlaneCredentials: set either token or tokenPath, not both")
	}
	var err error
	cc.password, cc.token = cc.Password, cc.Token
	if cc.PasswordPath != "" {
		if cc.password, err = readSecretFile(cc.PasswordPath, "control plane password"); err != nil {
			return err
		}
	}
	if cc.TokenPath != "" {
		if cc.token, err = readSecretFile(cc.TokenPath, "control plane token"); err != nil {
			return err
		}
	}
	return nil
}

func readSecretFile(path, name string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s %q: %w", name, path, err)
	}
	secret := strings.TrimSpace(string(b))
	if secret == "" {
		return "", fmt.Errorf("%s file %q is empty", name, path)
	}
	return secret, nil
}
