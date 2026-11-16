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

// The fixtures below reuse these Helm value keys; naming them keeps one
// fixture from silently testing a different key than the others.
const (
	keyGlobal  = "global"
	keyDNS     = "dns"
	keyArgs    = "args"
	keyName    = "name"
	keyEnabled = "enabled"
	valueTrue  = "true"
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

// Custom helm values are handed to Helm as --set arguments, which address
// nested values with dotted paths. A nested map in the config must therefore be
// flattened instead of being rendered as a Go map string.
func TestToSpec_CustomHelmValuesNestedMaps(t *testing.T) {
	cfg := &Config{
		ControlPlane: &ControlPlane{
			Enabled: true,
			Domain:  "runai.example.com",
			CustomHelmValues: map[string]any{
				"supportEntitled": true,
				"telemetryService": map[string]any{
					"enabled": true,
				},
				keyGlobal: map[string]any{
					"fipsMode": "only",
					"ingress":  map[string]any{"enabled": false},
				},
			},
		},
	}

	values := cfg.ToSpec(RunOptions{}).ControlPlane.CustomHelmValues

	assert.Equal(t, map[string]string{
		"supportEntitled":          "true",
		"telemetryService.enabled": "true",
		"global.fipsMode":          "only",
		"global.ingress.enabled":   "false",
	}, values)
}

func TestToSpec_CustomHelmValuesShapes(t *testing.T) {
	cases := []struct {
		name string
		in   map[string]any
		want map[string]string
	}{
		{
			name: "empty block maps to no overrides",
			in:   map[string]any{},
			want: nil,
		},
		{
			name: "already dotted keys are preserved",
			in:   map[string]any{"clusterConfig.global.subdomainSupport": false},
			want: map[string]string{"clusterConfig.global.subdomainSupport": "false"},
		},
		{
			name: "scalars keep their config type",
			in:   map[string]any{"replicas": 3, "ratio": 1.5, "name": "runai"},
			want: map[string]string{"replicas": "3", "ratio": "1.5", "name": "runai"},
		},
		{
			name: "lists use helm brace syntax",
			in:   map[string]any{keyDNS: []any{"1.1.1.1", "8.8.8.8"}},
			want: map[string]string{keyDNS: "{1.1.1.1,8.8.8.8}"},
		},
		{
			name: "commas inside list elements are escaped",
			in:   map[string]any{keyArgs: []any{"a,b", `c\d`}},
			want: map[string]string{keyArgs: `{a\,b,c\\d}`},
		},
		{
			name: "a map nested under a dotted key is still flattened",
			in:   map[string]any{"global.foo": map[string]any{"bar": 1}},
			want: map[string]string{"global.foo.bar": "1"},
		},
		{
			name: "empty nested maps do not clear chart defaults",
			in:   map[string]any{keyGlobal: map[string]any{}},
			want: nil,
		},
		{
			name: "braces inside list elements are escaped",
			in:   map[string]any{keyArgs: []any{"a}b", "{c}"}},
			want: map[string]string{keyArgs: `{a\}b,\{c\}}`},
		},
		{
			name: "braces inside scalar values are escaped",
			in:   map[string]any{"note": "a}b{c"},
			want: map[string]string{"note": `a\}b\{c`},
		},
		{
			// Helm's --set has no empty-list form: `key={}` sets a list holding
			// one empty string, so an empty list has to emit nothing.
			name: "empty lists do not become a one element list",
			in:   map[string]any{keyDNS: []any{}, keyArgs: []string{}},
			want: nil,
		},
		{
			// Both spellings name global.foo. Sorted iteration makes the result
			// independent of map order, and the key spelled out in full wins.
			name: "a dotted key wins over the same path from a nested map",
			in:   map[string]any{keyGlobal: map[string]any{"foo": "from-nested"}, "global.foo": "from-dotted"},
			want: map[string]string{"global.foo": "from-dotted"},
		},
		{
			// Helm splits keys on dots, so a nested key containing one has to be
			// escaped or the chart receives nested maps instead of that key.
			name: "dots in nested keys stay part of the key",
			in:   map[string]any{"podAnnotations": map[string]any{"prometheus.io/scrape": valueTrue}},
			want: map[string]string{`podAnnotations.prometheus\.io/scrape`: valueTrue},
		},
		{
			name: "nested keys escape every helm key separator",
			in:   map[string]any{"a": map[string]any{"b,c": 1, "d=e": 2, `f\g`: 3, "h[i": 4}},
			want: map[string]string{`a.b\,c`: "1", `a.d\=e`: "2", `a.f\\g`: "3", `a.h\[i`: "4"},
		},
		{
			name: "an indexed entry escapes the dots in its keys",
			in:   map[string]any{"ports": []any{map[string]any{"prometheus.io/scrape": true}}},
			want: map[string]string{`ports[0].prometheus\.io/scrape`: valueTrue},
		},
		{
			// Helm's brace syntax carries scalars only, so an entry that is a map
			// has to be addressed by index or the chart receives the Go map
			// string that this change exists to remove.
			name: "a list of maps uses indexed keys",
			in:   map[string]any{"tolerations": []any{map[string]any{"key": "nvidia.com/gpu", "operator": "Exists"}}},
			want: map[string]string{
				"tolerations[0].key":      "nvidia.com/gpu",
				"tolerations[0].operator": "Exists",
			},
		},
		{
			// The outer list needs indexes because one entry is a list, while the
			// inner list of scalars keeps the compact brace form Helm accepts on
			// the right hand side of an indexed key.
			name: "a list with a nested list addresses the outer entries by index",
			in:   map[string]any{keyArgs: []any{"plain", []any{"a", "b"}}},
			want: map[string]string{
				"args[0]": "plain",
				"args[1]": "{a,b}",
			},
		},
		{
			name: "an indexed list keeps helm escaping and types",
			in:   map[string]any{"ports": []any{map[string]any{keyName: "a,b", "number": 8080, keyEnabled: true}}},
			want: map[string]string{
				"ports[0].name":    `a\,b`,
				"ports[0].number":  "8080",
				"ports[0].enabled": valueTrue,
			},
		},
		{
			// Helm parses a bare `null` as an empty value; `%v` would have sent
			// the literal "<nil>".
			name: "null keeps its meaning",
			in:   map[string]any{"affinity": nil, keyArgs: []any{nil, "x"}},
			want: map[string]string{"affinity": "null", keyArgs: "{null,x}"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &Config{
				ControlPlane: &ControlPlane{Enabled: true, Domain: "runai.example.com", CustomHelmValues: tc.in},
			}
			assert.Equal(t, tc.want, cfg.ToSpec(RunOptions{}).ControlPlane.CustomHelmValues)
		})
	}
}
