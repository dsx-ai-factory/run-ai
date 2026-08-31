/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package loadconfig defines the airgap load config file and maps it onto the
// orchestrator spec, mirroring packconfig's conventions.
package loadconfig

type Config struct {
	Settings Settings `yaml:"settings,omitempty"`
	Registry Registry `yaml:"registry,omitempty"`
}

type Settings struct {
	BundlePath string `yaml:"bundlePath,omitempty"`
	WorkingDir string `yaml:"workingDir,omitempty"`
	ValuesDir  string `yaml:"valuesDir,omitempty"`
}

// Registry is the destination the bundle's images are pushed to, authenticated
// with the same schemes a pack's registry is.
type Registry struct {
	URL                string `yaml:"url,omitempty"`
	Username           string `yaml:"username,omitempty"`
	Password           string `yaml:"password,omitempty"`
	PasswordPath       string `yaml:"passwordPath,omitempty"`
	PasswordEnv        string `yaml:"passwordEnv,omitempty"`
	Token              string `yaml:"token,omitempty"`
	TokenPath          string `yaml:"tokenPath,omitempty"`
	TokenEnv           string `yaml:"tokenEnv,omitempty"`
	CAChainPath        string `yaml:"caChainPath,omitempty"`
	Insecure           bool   `yaml:"insecure,omitempty"`
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify,omitempty"`
}
