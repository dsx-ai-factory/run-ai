/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inputcheck

import "github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"

// tlsSource pairs an effective TLS block with its config field path.
type tlsSource struct {
	field string
	tls   *spec.TLS
}

func tlsSources(s *spec.Spec) []tlsSource {
	var out []tlsSource
	appendSource := func(src tlsSource) {
		if src.tls == nil {
			return
		}
		for _, existing := range out {
			if existing.field == src.field && existing.tls == src.tls {
				return
			}
		}
		out = append(out, src)
	}
	if s.ControlPlane != nil {
		appendSource(controlPlaneTLSSource(s))
	}
	if s.Cluster != nil {
		appendSource(clusterTLSSource(s))
		if inferenceTLSApplies(s) {
			appendSource(inferenceTLSSource(s))
		}
	}
	return out
}

func inferenceTLSApplies(s *spec.Spec) bool {
	return s.Cluster != nil &&
		s.Cluster.Inference != nil &&
		s.Cluster.Inference.WildcardDomain != "" &&
		!isGatewayAPI(s)
}

func controlPlaneTLSSource(s *spec.Spec) tlsSource {
	if s.ControlPlane != nil && s.ControlPlane.TLS != nil {
		return tlsSource{"controlPlane.tls", s.ControlPlane.TLS}
	}
	return tlsSource{"tls", s.TLS}
}

func clusterTLSSource(s *spec.Spec) tlsSource {
	if s.Cluster != nil && s.Cluster.TLS != nil {
		return tlsSource{"cluster.tls", s.Cluster.TLS}
	}
	return tlsSource{"tls", s.TLS}
}

func inferenceTLSSource(s *spec.Spec) tlsSource {
	if s.Cluster != nil && s.Cluster.Inference != nil && s.Cluster.Inference.WildcardTLS != nil {
		return tlsSource{"cluster.inference.wildcardTls", s.Cluster.Inference.WildcardTLS}
	}
	return clusterTLSSource(s)
}
