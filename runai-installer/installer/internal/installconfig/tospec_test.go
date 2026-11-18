/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package installconfig

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToSpec_InstallMode(t *testing.T) {
	cases := map[string]spec.InstallMode{
		"":                         spec.InstallModeFull,
		"full_installation":        spec.InstallModeFull,
		"runai_control_plane_only": spec.InstallModeControlPlaneOnly,
		"runai_cluster_only":       spec.InstallModeClusterOnly,
	}
	for in, want := range cases {
		cfg := &Config{Install: Install{Mode: in}, ControlPlane: &ControlPlane{Enabled: true, Domain: "runai.example.com"}}
		assert.Equal(t, want, cfg.ToSpec(RunOptions{}).InstallMode, "mode %q", in)
	}
}

func TestToSpec_FullInstall(t *testing.T) {
	cfg := &Config{
		ControlPlane: &ControlPlane{Enabled: true, Domain: "runai.example.com"},
		Cluster:      &Cluster{Enabled: true, Name: "cluster01"},
	}
	opts := RunOptions{
		DryRun:     true,
		Yes:        true,
		Output:     spec.OutputFormatText,
		EventsFile: "/tmp/events.ndjson",
		OnFailure:  spec.FailurePolicySkip,
		Registry:   &spec.Registry{Type: spec.RegistryTypeDockerConfig, Token: "tok"},
	}

	s := cfg.ToSpec(opts)

	assert.Equal(t, spec.CommandInstall, s.Command)
	assert.Equal(t, spec.InstallModeFull, s.InstallMode, "an empty install.mode defaults to full installation")
	assert.True(t, s.DryRun)
	assert.True(t, s.NonInteractive)
	assert.Equal(t, spec.OutputFormatText, s.Output.Format)
	assert.Equal(t, "/tmp/events.ndjson", s.Output.EventsFile)
	assert.Equal(t, spec.FailurePolicySkip, s.OnStepFailure)
	require.NotNil(t, s.Registry)
	assert.Equal(t, spec.RegistryTypeDockerConfig, s.Registry.Type)
	require.NotNil(t, s.ControlPlane)
	assert.Equal(t, "runai.example.com", s.ControlPlane.Domain)
	require.NotNil(t, s.Cluster)
	assert.Equal(t, "cluster01", s.Cluster.Name)
}

func TestToSpec_MapsAllFields(t *testing.T) {
	cfg := &Config{
		Install: Install{Mode: "full_installation", Connectivity: "connected"},
		TLS:     &TLS{CertPath: "/root.crt", KeyPath: "/root.key", CAChainPath: "/root-ca.crt"},
		ControlPlane: &ControlPlane{
			Enabled:          true,
			Domain:           "runai.example.com",
			Version:          "2.26.0",
			TLS:              &TLS{CertPath: "/cp.crt", KeyPath: "/cp.key", CAChainPath: "/ca.crt"},
			CustomHelmValues: map[string]any{"global.foo": "bar", "replicas": 3},
		},
		Cluster: &Cluster{
			Enabled:         true,
			Name:            "cluster01",
			Domain:          "cluster.example.com",
			Version:         "2.26.1",
			WorkloadRouting: "path",
			ControlPlaneURL: "https://cp.example.com",
			UID:             "cluster-uid",
			bootstrapToken:  "bootstrap-token",
			TLS:             &TLS{CertPath: "/cl.crt", KeyPath: "/cl.key"},
			Inference: &Inference{
				WildcardDomain: "*.inf.cluster.example.com",
				WildcardTLS:    &TLS{CertPath: "/wc.crt", KeyPath: "/wc.key"},
			},
		},
	}
	opts := RunOptions{Env: spec.EnvLab, Registry: &spec.Registry{Type: spec.RegistryTypeDockerConfig, Token: "tok"}}

	s := cfg.ToSpec(opts)

	assert.Equal(t, spec.ConnectivityConnected, s.Connectivity)
	assert.Equal(t, spec.EnvLab, s.Env)
	require.NotNil(t, s.TLS)
	assert.Equal(t, "/root.crt", s.TLS.CertPath)
	assert.Equal(t, "/root.key", s.TLS.KeyPath)
	assert.Equal(t, "/root-ca.crt", s.TLS.CAChainPath)

	require.NotNil(t, s.ControlPlane)
	assert.Equal(t, "2.26.0", s.ControlPlane.Version)
	require.NotNil(t, s.ControlPlane.TLS)
	assert.Equal(t, "/cp.crt", s.ControlPlane.TLS.CertPath)
	assert.Equal(t, "/ca.crt", s.ControlPlane.TLS.CAChainPath)
	assert.Equal(t, "bar", s.ControlPlane.CustomHelmValues["global.foo"])
	assert.Equal(t, "3", s.ControlPlane.CustomHelmValues["replicas"])

	require.NotNil(t, s.Cluster)
	assert.Equal(t, "cluster.example.com", s.Cluster.Domain)
	assert.Equal(t, "2.26.1", s.Cluster.Version)
	assert.Equal(t, "path", s.Cluster.WorkloadRouting)
	assert.Equal(t, "https://cp.example.com", s.Cluster.ControlPlaneURL)
	assert.Equal(t, "cluster-uid", s.Cluster.UID)
	assert.Equal(t, "bootstrap-token", s.Cluster.BootstrapToken)
	require.NotNil(t, s.Cluster.TLS)
	assert.Equal(t, "/cl.key", s.Cluster.TLS.KeyPath)

	require.NotNil(t, s.Cluster.Inference)
	assert.Equal(t, "*.inf.cluster.example.com", s.Cluster.Inference.WildcardDomain)
	require.NotNil(t, s.Cluster.Inference.WildcardTLS)
	assert.Equal(t, "/wc.crt", s.Cluster.Inference.WildcardTLS.CertPath)
	assert.Same(t, s.ControlPlane.TLS, s.ControlPlaneTLS())
	assert.Same(t, s.Cluster.TLS, s.ClusterTLS())
	assert.Same(t, s.Cluster.Inference.WildcardTLS, s.InferenceTLS())
}

func TestToSpec_TLSDefaults(t *testing.T) {
	cfg := &Config{
		TLS:          &TLS{CertPath: "/root.crt", KeyPath: "/root.key"},
		ControlPlane: &ControlPlane{Enabled: true, Domain: "runai.example.com"},
		Cluster: &Cluster{
			Enabled:   true,
			Name:      "cluster01",
			Inference: &Inference{WildcardDomain: "*.inference.runai.example.com"},
		},
	}

	s := cfg.ToSpec(RunOptions{})

	require.NotNil(t, s.TLS)
	assert.Same(t, s.TLS, s.ControlPlaneTLS())
	assert.Same(t, s.TLS, s.ClusterTLS())
	assert.Same(t, s.TLS, s.InferenceTLS())
}

func TestToSpec_SelfSignedTLS(t *testing.T) {
	cfg := &Config{
		TLS: &TLS{
			GenerateSelfSigned: true,
			CertPath:           "/root.crt",
			KeyPath:            "/root.key",
			CAChainPath:        "/root-ca.crt",
		},
		ControlPlane: &ControlPlane{
			Enabled: true,
			Domain:  "cp.example.com",
			TLS:     &TLS{CertPath: "/cp.crt"},
		},
		Cluster: &Cluster{
			Enabled: true,
			Name:    "cluster01",
			Domain:  "cluster.example.com",
			TLS:     &TLS{CertPath: "/cluster.crt"},
			Inference: &Inference{
				WildcardDomain: "*.inference.cluster.example.com",
				WildcardTLS:    &TLS{CertPath: "/inference.crt"},
			},
		},
	}

	s := cfg.ToSpec(RunOptions{})

	require.True(t, s.SelfSignedTLSEnabled())
	assert.Equal(t, []string{
		"tls.certPath",
		"tls.keyPath",
		"tls.caChainPath",
		"controlPlane.tls",
		"cluster.tls",
		"cluster.inference.wildcardTls",
	}, s.SelfSignedTLS.ConflictingFields)
	assert.Equal(t, []string{
		"*.cluster.example.com",
		"*.cp.example.com",
		"*.inference.cluster.example.com",
		"cluster.example.com",
		"cp.example.com",
	}, s.SelfSignedTLSDNSNames())
}

func TestToSpec_ControlPlaneDisabled(t *testing.T) {
	cfg := &Config{
		ControlPlane: &ControlPlane{Enabled: false, Domain: "runai.example.com"},
		Cluster:      &Cluster{Enabled: true, Name: "cluster01"},
	}

	s := cfg.ToSpec(RunOptions{})

	assert.Nil(t, s.ControlPlane, "disabled control plane must not be mapped")
	require.NotNil(t, s.Cluster)
}

func TestToSpec_ClusterDisabled(t *testing.T) {
	cfg := &Config{
		ControlPlane: &ControlPlane{Enabled: true, Domain: "runai.example.com"},
		Cluster:      &Cluster{Enabled: false, Name: "cluster01"},
	}

	s := cfg.ToSpec(RunOptions{})

	require.NotNil(t, s.ControlPlane)
	assert.Nil(t, s.Cluster, "disabled cluster must not be mapped")
}

func TestToSpec_ControlPlaneCredentials(t *testing.T) {
	cfg := &Config{
		Install: Install{Mode: "runai_cluster_only"},
		Cluster: &Cluster{
			Enabled:                 true,
			Name:                    "cluster01",
			Domain:                  "cluster.example.com",
			ControlPlaneURL:         "https://cp.example.com",
			ControlPlaneCredentials: &ControlPlaneCredentials{Username: "admin@example.com", Password: "inline-pw"},
		},
	}

	s := cfg.ToSpec(RunOptions{})

	require.NotNil(t, s.Cluster.ControlPlaneCredentials)
	assert.Equal(t, "admin@example.com", s.Cluster.ControlPlaneCredentials.Username)
	assert.Equal(t, "inline-pw", s.Cluster.ControlPlaneCredentials.Password, "inline secrets map without the resolver")
	assert.Empty(t, s.Cluster.ControlPlaneCredentials.Token)
	assert.Empty(t, s.Cluster.UID)

	cfg.Cluster.ControlPlaneCredentials = &ControlPlaneCredentials{TokenPath: "/t", token: "resolved-token"}
	s = cfg.ToSpec(RunOptions{})
	assert.Equal(t, "resolved-token", s.Cluster.ControlPlaneCredentials.Token, "file-based secrets use the resolved value")

	cfg.Cluster.ControlPlaneCredentials = nil
	assert.Nil(t, cfg.ToSpec(RunOptions{}).Cluster.ControlPlaneCredentials)
}

func TestToStringMap_NestedAndDotted(t *testing.T) {
	cases := map[string]struct {
		in   map[string]any
		want map[string]string
	}{
		"empty": {in: nil, want: nil},
		"scalars unchanged": {
			in:   map[string]any{"supportEntitled": true, "replicas": 3, "name": "x"},
			want: map[string]string{"supportEntitled": "true", "replicas": "3", "name": "x"},
		},
		"dotted key stays as is": {
			in:   map[string]any{"telemetryService.enabled": true},
			want: map[string]string{"telemetryService.enabled": "true"},
		},
		"nested map flattened": {
			in: map[string]any{
				"telemetryService": map[string]any{"enabled": true},
				"global":           map[string]any{"fipsMode": "only"},
			},
			want: map[string]string{"telemetryService.enabled": "true", "global.fipsMode": "only"},
		},
		"deeply nested": {
			in:   map[string]any{"a": map[string]any{"b": map[string]any{"c": 1}}},
			want: map[string]string{"a.b.c": "1"},
		},
		"dotted and nested mixed": {
			in: map[string]any{
				"supportEntitled":  true,
				"global.fipsMode":  "only",
				"telemetryService": map[string]any{"telemetryAuth": map[string]any{"secretKey": "k"}},
			},
			want: map[string]string{
				"supportEntitled":                          "true",
				"global.fipsMode":                          "only",
				"telemetryService.telemetryAuth.secretKey": "k",
			},
		},
		"dotted key with nested value": {
			in:   map[string]any{"a.b": map[string]any{"c": 1}},
			want: map[string]string{"a.b.c": "1"},
		},
		"dot inside nested key is escaped": {
			in:   map[string]any{"ann": map[string]any{"foo.bar/baz": "v"}},
			want: map[string]string{`ann.foo\.bar/baz`: "v"},
		},
		"scalar list in helm brace syntax": {
			in:   map[string]any{"hosts": []any{"a", "b,c", `d\e`}},
			want: map[string]string{"hosts": `{a,b\,c,d\\e}`},
		},
		"empty string value": {
			in:   map[string]any{"a": ""},
			want: map[string]string{"a": ""},
		},
		"null value": {
			in:   map[string]any{"a": nil, "b": map[string]any{"c": nil}},
			want: map[string]string{"a": "null", "b.c": "null"},
		},
		"numbers and bools": {
			in:   map[string]any{"i": 1, "big": int64(9007199254740993), "f": 1.5, "whole": 2.0, "b": false},
			want: map[string]string{"i": "1", "big": "9007199254740993", "f": "1.5", "whole": "2", "b": "false"},
		},
		"empty list is skipped": {
			in:   map[string]any{"a": []any{}, "b": 1},
			want: map[string]string{"b": "1"},
		},
		"list of maps uses indexed paths": {
			in: map[string]any{"hosts": []any{
				map[string]any{"name": "a", "port": 80},
				map[string]any{"name": "b"},
			}},
			want: map[string]string{"hosts[0].name": "a", "hosts[0].port": "80", "hosts[1].name": "b"},
		},
		"nested list inside list": {
			in:   map[string]any{"a": []any{[]any{"x", "y"}, []any{"z"}}},
			want: map[string]string{"a[0]": "{x,y}", "a[1]": "{z}"},
		},
		"map in list in map": {
			in:   map[string]any{"a": map[string]any{"l": []any{map[string]any{"k": "v"}}}},
			want: map[string]string{"a.l[0].k": "v"},
		},
		"non-string keys": {
			in:   map[string]any{"a": map[any]any{1: "x", "b": true}},
			want: map[string]string{"a.1": "x", "a.b": "true"},
		},
		"nested and dotted conflict resolves deterministically": {
			in:   map[string]any{"a": map[string]any{"b": 1}, "a.b": 2},
			want: map[string]string{"a.b": "2"},
		},
		"all empty maps gives nil": {
			in:   map[string]any{"a": map[string]any{}, "b": map[string]any{"c": map[string]any{}}},
			want: nil,
		},
		"empty nested map skipped": {
			in:   map[string]any{"a": map[string]any{}, "b": 1},
			want: map[string]string{"b": "1"},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, toStringMap(tc.in))
		})
	}
}
