/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inputcheck

import (
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs/validation"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
)

func init() { checks = append(checks, certificateCheck) }

// certificateCheck validates every provided TLS certificate against the domain
// it will serve: the file decodes as PEM, parses, is currently valid (not
// expired), and its hostname matches. Certificates are only checked when a TLS
// block is present (it is optional). All reads are from disk; nothing reaches
// the network.
func certificateCheck(s *spec.Spec) []Finding {
	var findings []Finding

	if cp := s.ControlPlane; cp != nil {
		src := controlPlaneTLSSource(s)
		if src.tls != nil && src.tls.CertPath != "" {
			findings = append(findings, fromResult(
				validation.ValidateCertificate(src.tls.CertPath, cp.Domain),
				"certificate", src.field+".certPath",
			)...)
		}
	}

	if c := s.Cluster; c != nil {
		src := clusterTLSSource(s)
		if src.tls != nil && src.tls.CertPath != "" {
			findings = append(findings, fromResult(
				validation.ValidateCertificate(src.tls.CertPath, clusterDomain(s)),
				"certificate", src.field+".certPath",
			)...)
		}
	}

	findings = append(findings, wildcardFindings(s)...)

	return findings
}

// wildcardFindings validates the inference wildcard certificate against the
// wildcard domain. The wildcard certificate is optional: when no dedicated
// wildcard TLS is provided, the cluster's certificate is reused. A wildcard SAN
// matches exactly one label, so verification uses a synthetic concrete host
// under the wildcard base rather than the bare "*." domain.
func wildcardFindings(s *spec.Spec) []Finding {
	c := s.Cluster
	if c == nil || c.Inference == nil || c.Inference.WildcardDomain == "" {
		return nil
	}

	// Under Gateway API the Gateway listener terminates inference TLS; the
	// installer creates no wildcard secret and reuses no certificate, so there
	// is nothing to validate here (and the cluster-cert fallback would not be
	// expected to cover the wildcard).
	if isGatewayAPI(s) {
		return nil
	}

	src := inferenceTLSSource(s)
	if src.tls == nil || src.tls.CertPath == "" {
		return nil // TLS is optional overall: nothing to validate
	}

	return wildcardCertificateFindings(
		src,
		c.Inference.WildcardDomain,
		"inference wildcard domain",
	)
}

func wildcardCertificateFindings(src tlsSource, wildcard, label string) []Finding {
	findings := fromResult(
		validation.ValidateCertificate(src.tls.CertPath, sampleWildcardHost(wildcard)),
		"wildcard-certificate", src.field+".certPath",
	)
	for i := range findings {
		findings[i].Message = fmt.Sprintf("%s %q: %s", label, wildcard, findings[i].Message)
	}
	return findings
}

// sampleWildcardHost turns "*.inference.example.com" into a concrete host so
// VerifyHostname exercises the wildcard SAN correctly (a wildcard matches
// exactly one label, never the bare base domain).
func sampleWildcardHost(wildcard string) string {
	if strings.HasPrefix(wildcard, "*.") {
		return "runai-validation." + strings.TrimPrefix(wildcard, "*.")
	}
	return wildcard
}

// clusterDomain returns the domain the cluster certificate must serve,
// mirroring the plan's fallback to the control-plane domain when the cluster
// domain is unset.
func clusterDomain(s *spec.Spec) string {
	if s.Cluster != nil && s.Cluster.Domain != "" {
		return s.Cluster.Domain
	}
	if s.ControlPlane != nil {
		return s.ControlPlane.Domain
	}
	return ""
}
