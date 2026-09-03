/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package installconfig

import "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"

func toSpecSelfSignedTLS(c *Config) *spec.SelfSignedTLS {
	if c.TLS == nil || !c.TLS.GenerateSelfSigned {
		return nil
	}
	return &spec.SelfSignedTLS{
		Enabled:           true,
		ConflictingFields: selfSignedTLSConflictingFields(c),
	}
}

func selfSignedTLSConflictingFields(c *Config) []string {
	var conflicts []string
	if c.TLS.CertPath != "" {
		conflicts = append(conflicts, "tls.certPath")
	}
	if c.TLS.KeyPath != "" {
		conflicts = append(conflicts, "tls.keyPath")
	}
	if c.TLS.CAChainPath != "" {
		conflicts = append(conflicts, "tls.caChainPath")
	}
	if c.ControlPlane != nil && c.ControlPlane.TLS != nil {
		conflicts = append(conflicts, "controlPlane.tls")
	}
	if c.Cluster == nil {
		return conflicts
	}
	if c.Cluster.TLS != nil {
		conflicts = append(conflicts, "cluster.tls")
	}
	if c.Cluster.Inference != nil && c.Cluster.Inference.WildcardTLS != nil {
		conflicts = append(conflicts, "cluster.inference.wildcardTls")
	}
	return conflicts
}
