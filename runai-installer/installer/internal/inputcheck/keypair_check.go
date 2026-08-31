/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inputcheck

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/inputs/validation"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
)

func init() { checks = append(checks, keyPairCheck) }

// keyPairCheck verifies that each provided certificate and its private key form
// a matching pair (and both decode), purely from disk. A TLS block with neither
// path set is skipped; the certificate check reports a missing certificate.
func keyPairCheck(s *spec.Spec) []Finding {
	var findings []Finding
	for _, src := range tlsSources(s) {
		if src.tls.CertPath == "" && src.tls.KeyPath == "" {
			continue
		}
		findings = append(findings, fromResult(
			validation.ValidateKeyPair(src.tls.CertPath, src.tls.KeyPath),
			"certificate-key-pair", src.field+".keyPath",
		)...)
	}
	return findings
}
