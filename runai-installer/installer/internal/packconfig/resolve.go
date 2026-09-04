/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package packconfig

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
)

// tokenUser is the username a registry token authenticates as, following the
// NGC convention.
const tokenUser = "$oauthtoken"

// resolveCredentials reads every registry's configured secret and flattens it
// into the username and password the pack authenticates with.
func resolveCredentials(auths []RegistryAuth) ([]spec.RegistryCredential, error) {
	if len(auths) == 0 {
		return nil, nil
	}
	out := make([]spec.RegistryCredential, 0, len(auths))
	for _, a := range auths {
		username, password, err := ResolveAuth(a)
		if err != nil {
			return nil, err
		}
		bearerToken, hasBearerToken, err := resolveSecret(a.Registry, "bearerToken", a.BearerToken, a.BearerTokenPath, a.BearerTokenEnv)
		if err != nil {
			return nil, err
		}
		if hasBearerToken && (username != "" || password != "") {
			return nil, fmt.Errorf("%s: bearer token authentication cannot be combined with another scheme", a.Registry)
		}
		out = append(out, spec.RegistryCredential{
			Prefix:             strings.TrimSpace(a.Registry),
			Username:           username,
			Password:           password,
			BearerToken:        bearerToken,
			CAChainPath:        a.CAChainPath,
			Insecure:           a.Insecure,
			InsecureSkipVerify: a.InsecureSkipVerify,
		})
	}
	return out, nil
}

// ResolveAuth returns the credentials of one registry, enforcing that only one
// authentication scheme is configured.
func ResolveAuth(a RegistryAuth) (username, password string, err error) {
	label := a.Registry
	if label == "" {
		label = "registryAuth entry"
	}

	token, hasToken, err := resolveSecret(label, "token", a.Token, a.TokenPath, a.TokenEnv)
	if err != nil {
		return "", "", err
	}
	pass, hasPassword, err := resolveSecret(label, "password", a.Password, a.PasswordPath, a.PasswordEnv)
	if err != nil {
		return "", "", err
	}
	hasUsername := strings.TrimSpace(a.Username) != ""

	switch {
	case hasToken && (hasPassword || hasUsername):
		return "", "", fmt.Errorf("%s: set either a token or a username and password, not both", label)
	case hasToken:
		return tokenUser, token, nil
	case hasPassword && !hasUsername:
		return "", "", fmt.Errorf("%s: a password requires a username", label)
	case hasUsername && !hasPassword:
		return "", "", fmt.Errorf("%s: a username requires a password", label)
	case hasUsername:
		return strings.TrimSpace(a.Username), pass, nil
	}
	return "", "", nil
}

// resolveSecret returns the value of the one configured source of a secret.
func resolveSecret(label, field, inline, path, env string) (string, bool, error) {
	sources := 0
	for _, s := range []string{inline, path, env} {
		if strings.TrimSpace(s) != "" {
			sources++
		}
	}
	switch {
	case sources == 0:
		return "", false, nil
	case sources > 1:
		return "", false, fmt.Errorf("%s: set only one of %s, %sPath or %sEnv", label, field, field, field)
	}

	switch {
	case strings.TrimSpace(inline) != "":
		return inline, true, nil
	case strings.TrimSpace(path) != "":
		value, err := readSecretFile(path)
		if err != nil {
			return "", false, fmt.Errorf("%s: read %sPath: %w", label, field, err)
		}
		return value, true, nil
	default:
		value := strings.TrimSpace(os.Getenv(env))
		if value == "" {
			return "", false, fmt.Errorf("%s: environment variable %s is not set", label, env)
		}
		return value, true, nil
	}
}

func readSecretFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	value := strings.TrimSpace(string(content))
	if value == "" {
		return "", errors.New("file is empty")
	}
	return value, nil
}
