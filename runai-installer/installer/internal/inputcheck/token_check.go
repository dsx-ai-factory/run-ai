/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inputcheck

import (
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs/validation"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
)

func init() { checks = append(checks, tokenCheck) }

// tokenCheck validates the registry token format. It is skipped for air-gapped
// installs (which authenticate to a mirror, not via a registry token) and for
// non-token auth. Only docker config tokens have a defined (base64) format to verify.
func tokenCheck(s *spec.Spec) []Finding {
	if s.IsAirGapped() || s.Registry == nil {
		return nil
	}
	if s.Registry.Type != spec.RegistryTypeDockerConfig || s.Registry.Token == "" {
		return nil
	}
	return fromResult(
		validation.ValidatePullSecret(s.Registry.Token),
		"token-format", "registry.token",
	)
}
