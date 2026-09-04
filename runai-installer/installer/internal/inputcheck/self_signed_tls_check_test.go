/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inputcheck_test

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputcheck"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/stretchr/testify/assert"
)

func TestSelfSignedTLSCheckReportsEveryConflict(t *testing.T) {
	fields := []string{
		"tls.certPath",
		"tls.keyPath",
		"tls.caChainPath",
		"controlPlane.tls",
		"cluster.tls",
		"cluster.inference.wildcardTls",
	}
	s := &spec.Spec{
		SelfSignedTLS: &spec.SelfSignedTLS{Enabled: true, ConflictingFields: fields},
		ControlPlane:  &spec.ControlPlane{Domain: "runai.example.com"},
	}

	report := inputcheck.Validate(s)

	var got []string
	for _, finding := range report.Findings {
		if finding.Check == "self-signed-tls" {
			got = append(got, finding.Field)
		}
	}
	assert.ElementsMatch(t, fields, got)
}

func TestSelfSignedTLSCheckRejectsUnsupportedTopology(t *testing.T) {
	cases := []struct {
		name string
		spec *spec.Spec
	}{
		{
			name: "gateway API",
			spec: &spec.Spec{
				SelfSignedTLS:      &spec.SelfSignedTLS{Enabled: true},
				ControlPlane:       &spec.ControlPlane{Domain: "runai.example.com"},
				ExternalConnection: &spec.ExternalConnection{Type: spec.ExternalConnectionGatewayAPI},
			},
		},
		{
			name: "OpenShift",
			spec: &spec.Spec{
				SelfSignedTLS:          &spec.SelfSignedTLS{Enabled: true},
				ControlPlane:           &spec.ControlPlane{Domain: "runai.example.com"},
				KubernetesDistribution: "openshift",
			},
		},
		{
			name: "missing domain",
			spec: &spec.Spec{
				SelfSignedTLS: &spec.SelfSignedTLS{Enabled: true},
				Cluster:       &spec.Cluster{Name: "cluster"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			report := inputcheck.Validate(tc.spec)
			var found bool
			for _, finding := range report.Findings {
				if finding.Check == "self-signed-tls" {
					found = true
				}
			}
			assert.True(t, found)
		})
	}
}

func TestSelfSignedTLSSatisfiesKnativeCertificateRequirement(t *testing.T) {
	s := &spec.Spec{
		SelfSignedTLS: &spec.SelfSignedTLS{Enabled: true},
		ControlPlane:  &spec.ControlPlane{Domain: "runai.example.com"},
		Cluster: &spec.Cluster{
			Name:      "cluster",
			Inference: &spec.Inference{WildcardDomain: "*.inference.runai.example.com"},
		},
		Features: &spec.Features{
			Knative: &spec.Dependency{Mode: spec.DependencyModeInstall},
		},
		ExternalConnection: &spec.ExternalConnection{Type: spec.ExternalConnectionIngress},
	}

	report := inputcheck.Validate(s)

	for _, finding := range report.Findings {
		assert.NotEqual(t, "knative-requires-certificate", finding.Check)
	}
}
