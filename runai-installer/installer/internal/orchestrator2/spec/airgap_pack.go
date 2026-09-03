/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package spec

import (
	"errors"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/pkg/artifactlock"
)

// AirgapPack is the fully resolved input of the airgap pack command: the lock
// is parsed and the credentials are read, so nothing downstream performs
// configuration I/O. It shares none of the install settings on Spec, only the
// run-level envelope.
type AirgapPack struct {
	Lock           *artifactlock.ArtifactLock
	RunaiVersion   string
	Architecture   string
	OutputPath     string
	WorkDir        string
	Credentials    []RegistryCredential
	ImageOverrides []SourceOverride
}

// RegistryCredential authenticates to one registry. Prefix is the match key: a
// host, or a host and path prefix scoping the credentials to one project.
type RegistryCredential struct {
	Prefix             string
	Username           string
	Password           string
	BearerToken        string
	CAChainPath        string
	Insecure           bool
	InsecureSkipVerify bool
}

// SourceOverride redirects a named artifact to a different source.
type SourceOverride struct {
	Group  string
	Name   string
	Source string
}

func (p *AirgapPack) validate() error {
	if p == nil {
		return errors.New("airgap pack settings are required")
	}
	if p.Lock == nil {
		return errors.New("an artifact lock is required")
	}
	if p.OutputPath == "" {
		return errors.New("output path is required")
	}
	if p.WorkDir == "" {
		return errors.New("working directory is required")
	}
	if arch := p.Lock.Spec.Platform.Architecture; !artifactlock.IsValidArchitecture(arch) {
		return fmt.Errorf("artifact lock has an unsupported architecture: %q", arch)
	}
	return nil
}
