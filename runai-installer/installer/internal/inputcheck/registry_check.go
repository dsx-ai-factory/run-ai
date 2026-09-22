/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inputcheck

import (
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
)

func init() { checks = append(checks, registryHostCheck, registryAirgapAuthCheck) }

// registryAirgapAuthCheck rejects token-based registry auth for air-gapped
// installs. The image push runs setup.sh with REGISTRY_USERNAME/REGISTRY_PASSWORD;
// a registry token is never forwarded to it, so a token-authenticated mirror
// would silently fail to push. Air-gapped mirrors must use username/password,
// or no credentials for an open mirror. Token auth is meaningful only for
// connected installs that target Run:ai's hosted registries.
func registryAirgapAuthCheck(s *spec.Spec) []Finding {
	if !s.IsAirGapped() || s.Registry == nil || !isTokenAuth(s.Registry) {
		return nil
	}
	return []Finding{{
		Check:   "registry-airgap-auth",
		Field:   "registry",
		Message: "air-gapped installs must authenticate to the mirror with username/password (or no credentials), not a token",
	}}
}

// registryHostCheck enforces that a registry host is present whenever the
// installer cannot target a default Run:ai registry:
//
//   - Air-gapped installs always push to and pull from the customer mirror, so a
//     host is mandatory regardless of how the mirror authenticates. Without it
//     the image push (setup.sh) and the chart pulls run against an empty
//     registry.
//   - Basic-auth and unauthenticated registries identify a specific mirror by
//     host; without it the resolved registry is not actionable.
//
// Token auth (docker config or NGC) targets Run:ai's known registries, where the host is
// optional, so it is exempt for connected installs.
func registryHostCheck(s *spec.Spec) []Finding {
	if s.IsAirGapped() {
		if s.Registry == nil || strings.TrimSpace(s.Registry.Host) == "" {
			return []Finding{{
				Check:   "registry-host",
				Field:   "registry.host",
				Message: "registry host is required for air-gapped installs (images are pushed to and pulled from the mirror)",
			}}
		}
		return nil
	}

	if s.Registry == nil || targetsKnownRegistry(s.Registry) {
		return nil
	}

	if strings.TrimSpace(s.Registry.Host) == "" {
		return []Finding{{
			Check:   "registry-host",
			Field:   "registry.host",
			Message: "registry host is required for basic-auth and unauthenticated registries",
		}}
	}
	return nil
}

// isTokenAuth reports whether the registry authenticates with a registry token
// rather than basic auth or no credentials.
func isTokenAuth(r *spec.Registry) bool {
	return r.Token != ""
}

func targetsKnownRegistry(r *spec.Registry) bool {
	return r.Token != "" || r.Type == spec.RegistryTypeNGC
}
