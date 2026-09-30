/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package install_v2_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/component-tests/harness"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/installconfig"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/specrun"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests drive the full CLI entrypoint (Config -> ToSpec -> specrun.Execute)
// with intentionally broken inputs and assert that offline input validation
// aborts the run before any work: a single input_validation_failed event is
// emitted carrying the findings, the run returns an error, and no helm/k8s/API
// side effects occur.

const validationDomain = "test.runai.local"

func findValidationEvent(events []map[string]any) map[string]any {
	for _, e := range events {
		if e["kind"] == "input_validation_failed" {
			return e
		}
	}
	return nil
}

func findingMessages(ev map[string]any) string {
	findings, _ := ev["findings"].([]any)
	var sb strings.Builder
	for _, f := range findings {
		m, _ := f.(map[string]any)
		fmt.Fprintf(&sb, "%v: %v\n", m["field"], m["message"])
	}
	return sb.String()
}

func TestInputValidation_Failures(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)

	cases := []struct {
		name    string
		setup   func(t *testing.T, cfg *installconfig.Config, opts *installconfig.RunOptions)
		wantSub string
	}{
		{
			name: "missing certificate file",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.ControlPlane.TLS = &installconfig.TLS{
					CertPath: filepath.Join(t.TempDir(), "absent.crt"),
					KeyPath:  filepath.Join(t.TempDir(), "absent.key"),
				}
			},
			wantSub: "Cannot read certificate",
		},
		{
			name: "malformed PEM",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				path := filepath.Join(t.TempDir(), "tls.crt")
				require.NoError(t, os.WriteFile(path, []byte("not a certificate"), 0600))
				cfg.ControlPlane.TLS = &installconfig.TLS{CertPath: path, KeyPath: path}
			},
			wantSub: "Invalid PEM certificate",
		},
		{
			name: "expired certificate",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cert, key := harness.GenCert(t, []string{validationDomain}, time.Now().Add(-time.Hour))
				cfg.ControlPlane.TLS = &installconfig.TLS{CertPath: cert, KeyPath: key}
			},
			wantSub: "expired",
		},
		{
			name: "control-plane hostname mismatch",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cert, key := harness.GenCert(t, []string{"other.example.com"}, future)
				cfg.ControlPlane.TLS = &installconfig.TLS{CertPath: cert, KeyPath: key}
			},
			wantSub: "Hostname mismatch",
		},
		{
			name: "cluster hostname mismatch",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cert, key := harness.GenCert(t, []string{"wrong.example.com"}, future)
				cfg.Cluster.TLS = &installconfig.TLS{CertPath: cert, KeyPath: key}
			},
			wantSub: "Hostname mismatch",
		},
		{
			name: "explicit wildcard cert does not cover wildcard domain",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cert, key := harness.GenCert(t, []string{"*.wrong.example.com"}, future)
				cfg.Cluster.Inference = &installconfig.Inference{
					WildcardDomain: "*.inference.test.runai.local",
					WildcardTLS:    &installconfig.TLS{CertPath: cert, KeyPath: key},
				}
			},
			// The message names the configured wildcard domain, not just the
			// synthetic validation host.
			wantSub: `wildcard domain "*.inference.test.runai.local"`,
		},
		{
			name: "fallback cluster cert does not cover wildcard domain",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				// Valid for the cluster domain but not the inference wildcard;
				// with no dedicated wildcard cert the cluster cert is reused.
				cert, key := harness.GenCert(t, []string{"test.runai.local"}, future)
				cfg.Cluster.TLS = &installconfig.TLS{CertPath: cert, KeyPath: key}
				cfg.Cluster.Inference = &installconfig.Inference{WildcardDomain: "*.inference.test.runai.local"}
			},
			wantSub: "Hostname mismatch",
		},
		{
			name: "certificate and key do not match",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cert, _ := harness.GenCert(t, []string{validationDomain}, future)
				_, otherKey := harness.GenCert(t, []string{validationDomain}, future)
				cfg.ControlPlane.TLS = &installconfig.TLS{CertPath: cert, KeyPath: otherKey}
			},
			wantSub: "do not match",
		},
		{
			name: "invalid control-plane FQDN format",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.ControlPlane.Domain = "not_a_valid_domain"
			},
			wantSub: "Invalid FQDN format",
		},
		{
			name: "invalid inference wildcard FQDN format",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				// Missing the leading "*." makes it an invalid wildcard.
				cfg.Cluster.Inference = &installconfig.Inference{WildcardDomain: "inference.test.runai.local"}
			},
			wantSub: "Invalid wildcard FQDN format",
		},
		{
			name: "missing CA chain file",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cert, key := harness.GenCert(t, []string{validationDomain}, future)
				cfg.ControlPlane.TLS = &installconfig.TLS{
					CertPath:    cert,
					KeyPath:     key,
					CAChainPath: filepath.Join(t.TempDir(), "absent-ca.crt"),
				}
			},
			wantSub: "not accessible",
		},
		{
			name: "invalid docker config token format",
			setup: func(t *testing.T, _ *installconfig.Config, opts *installconfig.RunOptions) {
				opts.Registry = &spec.Registry{Type: spec.RegistryTypeDockerConfig, Token: "not valid base64!!"}
			},
			wantSub: "base64-encoded docker config",
		},
		{
			name: "missing chart repository for a docker config registry",
			setup: func(t *testing.T, _ *installconfig.Config, opts *installconfig.RunOptions) {
				t.Setenv(step.ChartRepoBaseEnv, "")
				opts.Registry = &spec.Registry{Type: spec.RegistryTypeDockerConfig, Token: "dG9r"}
			},
			wantSub: "chart repository is required",
		},
		{
			name: "missing chart repository for a non-prod NGC registry",
			setup: func(t *testing.T, _ *installconfig.Config, opts *installconfig.RunOptions) {
				t.Setenv(step.NGCNonProdChartRepoEnv, "")
				opts.Env = spec.EnvStaging
				opts.Registry = &spec.Registry{Type: spec.RegistryTypeNGC, Token: "ngc-key"}
			},
			wantSub: step.NGCNonProdChartRepoEnv,
		},
		{
			name: "missing registry CA chain file",
			setup: func(t *testing.T, _ *installconfig.Config, opts *installconfig.RunOptions) {
				opts.Registry.CAChainPath = filepath.Join(t.TempDir(), "absent-registry-ca.crt")
			},
			wantSub: "registry.caChainPath",
		},
		{
			name: "invalid cluster FQDN format",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.Cluster.Domain = "not_a_valid_domain"
			},
			wantSub: "cluster.domain",
		},
		{
			name: "invalid GPU operator type",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.Prerequisites = &installconfig.Prerequisites{
					GPUOperator: &installconfig.Dependency{Mode: "install", Type: "simulated"},
				}
			},
			wantSub: "invalid GPU operator type",
		},
		{
			name: "invalid kubernetes distribution",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.Install.KubernetesDistribution = "k3s"
			},
			wantSub: "install.kubernetesDistribution",
		},
		{
			name: "prometheus install rejected on openshift",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.Install.KubernetesDistribution = "openshift"
				cfg.ExternalConnection = nil // OpenShift uses Routes
				cfg.Prerequisites = &installconfig.Prerequisites{
					Prometheus: &installconfig.Dependency{Mode: "install"},
				}
			},
			wantSub: "installing Prometheus is not supported on OpenShift",
		},
		{
			name: "dynamo without grove",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.Features = &installconfig.Features{Dynamo: &installconfig.Dependency{Mode: "install"}}
			},
			wantSub: "Dynamo requires Grove",
		},
		{
			name: "invalid system nodes mode",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.SystemNodes = &installconfig.SystemNodes{Enabled: true, Mode: "hard", Nodes: []string{"n1"}}
			},
			wantSub: "invalid mode",
		},
		{
			name: "system nodes managed without selection",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.SystemNodes = &installconfig.SystemNodes{Enabled: true}
			},
			wantSub: "exactly one of systemNodes.nodes or systemNodes.selector",
		},
		{
			name: "knative without inference wildcard domain",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.Features = &installconfig.Features{Knative: &installconfig.Dependency{Mode: "install"}}
			},
			wantSub: "Knative requires an inference wildcard domain",
		},
		{
			name: "knative with wildcard domain but no certificate",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				// Wildcard domain set, but no wildcard cert and no cluster cert
				// to fall back to: the fronting Ingress would reference a secret
				// that is never created.
				cfg.Cluster.TLS = nil
				cfg.Cluster.Inference = &installconfig.Inference{WildcardDomain: "*.inference.test.runai.local"}
				cfg.Features = &installconfig.Features{Knative: &installconfig.Dependency{Mode: "install"}}
			},
			wantSub: "Knative requires a wildcard certificate",
		},
		{
			name: "kubeflow with custom helm values",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.Features = &installconfig.Features{Kubeflow: &installconfig.Dependency{
					Mode:             "install",
					CustomHelmValues: map[string]any{"image.tag": "v1.8.1"},
				}}
			},
			wantSub: "features.kubeflow.customHelmValues",
		},
		{
			name: "mpi operator with custom helm values",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.Features = &installconfig.Features{MPIOperator: &installconfig.Dependency{
					Mode:             "install",
					CustomHelmValues: map[string]any{"image.tag": "v0.6.0"},
				}}
			},
			wantSub: "features.mpiOperator.customHelmValues",
		},
		{
			name: "external connection missing on non-openshift",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.ExternalConnection = nil
			},
			wantSub: "externalConnection is required",
		},
		{
			name: "non-gateway external connection set on openshift",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.Install.KubernetesDistribution = "openshift"
				// externalConnection is left set to ingress (from the base config).
			},
			wantSub: "on OpenShift only \"gatewayApi\" may be set",
		},
		{
			name: "gateway api without name and namespace on openshift",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.Install.KubernetesDistribution = "openshift"
				cfg.ExternalConnection = &installconfig.ExternalConnection{Type: "gatewayApi", GatewayAPI: &installconfig.GatewayAPIConnection{Name: "gw"}}
			},
			wantSub: "gatewayApi.name and gatewayApi.namespace are required",
		},
		{
			name: "ingress without class name",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.ExternalConnection = &installconfig.ExternalConnection{Type: "ingress", Ingress: &installconfig.IngressConnection{}}
			},
			wantSub: "ingress.className is required",
		},
		{
			name: "gateway api without name and namespace",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.ExternalConnection = &installconfig.ExternalConnection{Type: "gatewayApi", GatewayAPI: &installconfig.GatewayAPIConnection{Name: "gw"}}
			},
			wantSub: "gatewayApi.name and gatewayApi.namespace are required",
		},
		{
			name: "invalid external connection type",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.ExternalConnection = &installconfig.ExternalConnection{Type: "loadbalancer"}
			},
			wantSub: "invalid externalConnection type",
		},
		{
			name: "route type on non-openshift",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.ExternalConnection = &installconfig.ExternalConnection{Type: "route"}
			},
			wantSub: "route is selected automatically on OpenShift",
		},
		{
			name: "gateway api set when type is ingress",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.ExternalConnection = &installconfig.ExternalConnection{
					Type:       "ingress",
					Ingress:    &installconfig.IngressConnection{ClassName: "nginx"},
					GatewayAPI: &installconfig.GatewayAPIConnection{Name: "gw", Namespace: "gw-system"},
				}
			},
			wantSub: "gatewayApi must not be set when type is ingress",
		},
		{
			name: "ingress set when type is gateway api",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.ExternalConnection = &installconfig.ExternalConnection{
					Type:       "gatewayApi",
					GatewayAPI: &installconfig.GatewayAPIConnection{Name: "gw", Namespace: "gw-system"},
					Ingress:    &installconfig.IngressConnection{ClassName: "nginx"},
				}
			},
			wantSub: "ingress must not be set when type is gatewayApi",
		},
		{
			name: "kubeflow with custom version",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.Features = &installconfig.Features{Kubeflow: &installconfig.Dependency{
					Mode:    "install",
					Version: "v1.7.0",
				}}
			},
			wantSub: "features.kubeflow.version",
		},
		{
			name: "mpi operator with custom version",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.Features = &installconfig.Features{MPIOperator: &installconfig.Dependency{
					Mode:    "install",
					Version: "v0.5.0",
				}}
			},
			wantSub: "features.mpiOperator.version",
		},
		{
			name: "gpu custom helm values rejected on openshift",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.Install.KubernetesDistribution = "openshift"
				cfg.ExternalConnection = nil // OpenShift uses Routes
				cfg.Prerequisites = &installconfig.Prerequisites{GPUOperator: &installconfig.Dependency{
					Mode:             "install",
					Type:             "real",
					CustomHelmValues: map[string]any{"driver.enabled": true},
				}}
			},
			wantSub: "custom Helm values are not supported on OpenShift",
		},
		{
			name: "gpu openshift custom resource rejected on vanilla",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.Prerequisites = &installconfig.Prerequisites{GPUOperator: &installconfig.Dependency{
					Mode:      "install",
					Type:      "real",
					OpenShift: &installconfig.DependencyOpenShift{CustomResource: map[string]any{"driver": map[string]any{"version": "550"}}},
				}}
			},
			wantSub: "openshift.customResource only applies",
		},
		{
			name: "openshift custom resource rejected for a helm-only prerequisite",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.Install.KubernetesDistribution = "openshift"
				cfg.ExternalConnection = nil
				cfg.Prerequisites = &installconfig.Prerequisites{Prometheus: &installconfig.Dependency{
					Mode:      "install",
					OpenShift: &installconfig.DependencyOpenShift{CustomResource: map[string]any{"foo": "bar"}},
				}}
			},
			wantSub: "openshift.customResource only applies",
		},
		{
			name: "network operator custom helm values rejected on openshift",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.Install.KubernetesDistribution = "openshift"
				cfg.ExternalConnection = nil
				cfg.Prerequisites = &installconfig.Prerequisites{NetworkOperator: &installconfig.Dependency{
					Mode:             "install",
					CustomHelmValues: map[string]any{"sriov.enabled": true},
				}}
			},
			wantSub: "custom Helm values are not supported on OpenShift",
		},
		{
			name: "nim operator custom resource rejected (no managed CR)",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.Install.KubernetesDistribution = "openshift"
				cfg.ExternalConnection = nil
				cfg.Features = &installconfig.Features{NIMOperator: &installconfig.Dependency{
					Mode:      "install",
					OpenShift: &installconfig.DependencyOpenShift{CustomResource: map[string]any{"foo": "bar"}},
				}}
			},
			wantSub: "openshift.customResource only applies",
		},
		{
			name: "knative custom resource rejected on a non-openshift install",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.Cluster.Inference = &installconfig.Inference{WildcardDomain: "*.inference.test.runai.local"}
				cfg.Features = &installconfig.Features{Knative: &installconfig.Dependency{
					Mode:      "install",
					OpenShift: &installconfig.DependencyOpenShift{CustomResource: map[string]any{"config": map[string]any{"autoscaler": map[string]any{"stable-window": "60s"}}}},
				}}
			},
			wantSub: "openshift.customResource only applies",
		},
		{
			name: "invalid installPlanApproval value",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.Install.KubernetesDistribution = "openshift"
				cfg.ExternalConnection = nil
				cfg.Prerequisites = &installconfig.Prerequisites{GPUOperator: &installconfig.Dependency{
					Mode:      "install",
					Type:      "real",
					OpenShift: &installconfig.DependencyOpenShift{InstallPlanApproval: "sometimes"},
				}}
			},
			wantSub: "invalid installPlanApproval",
		},
		{
			name: "installPlanApproval rejected for a helm-only prerequisite",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.Install.KubernetesDistribution = "openshift"
				cfg.ExternalConnection = nil
				cfg.Prerequisites = &installconfig.Prerequisites{Prometheus: &installconfig.Dependency{
					Mode:      "install",
					OpenShift: &installconfig.DependencyOpenShift{InstallPlanApproval: "manual"},
				}}
			},
			wantSub: "openshift.installPlanApproval only applies",
		},
		{
			name: "subscription config rejected on a non-openshift install",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.Features = &installconfig.Features{NIMOperator: &installconfig.Dependency{
					Mode:      "install",
					OpenShift: &installconfig.DependencyOpenShift{SubscriptionConfig: &installconfig.SubscriptionConfig{Env: map[string]string{"ENABLE_WEBHOOKS": "false"}}},
				}}
			},
			wantSub: "openshift.subscriptionConfig only applies",
		},
		{
			name: "control-plane version not in the compatibility matrix",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.ControlPlane.Version = "2.23.0" // not in the matrix; no --dev-mode
			},
			wantSub: "controlPlane.version",
		},
		{
			name: "cluster version not in the compatibility matrix",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.ControlPlane.Version = "2.26.0" // valid
				cfg.Cluster.Version = "2.23.0"      // not in the matrix; checked independently
			},
			wantSub: "cluster.version",
		},
		{
			name: "unknown install mode",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.Install.Mode = "everything"
			},
			wantSub: "unknown install mode",
		},
		{
			name: "control-plane-only with a cluster configured",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.Install.Mode = "runai_control_plane_only" // cluster stays enabled
			},
			wantSub: "control-plane-only install must not configure a cluster",
		},
		{
			name: "cluster-only with a control plane configured",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.Install.Mode = "runai_cluster_only" // control plane stays enabled
			},
			wantSub: "cluster-only install must not configure a control plane",
		},
		{
			name: "invalid workload routing value",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.Cluster.WorkloadRouting = "subdomain"
			},
			wantSub: "Must be 'path' or 'host'",
		},
		{
			name: "workload routing host contradicted by subdomainSupport override",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.Cluster.WorkloadRouting = "host"
				cfg.Cluster.CustomHelmValues = map[string]any{"clusterConfig.global.subdomainSupport": false}
			},
			wantSub: "they must agree",
		},
		{
			name: "workload routing path contradicted by subdomainSupport override",
			setup: func(t *testing.T, cfg *installconfig.Config, _ *installconfig.RunOptions) {
				cfg.Cluster.WorkloadRouting = "path"
				cfg.Cluster.CustomHelmValues = map[string]any{"clusterConfig.global.subdomainSupport": true}
			},
			wantSub: "they must agree",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := newInstallRunConfig()

			eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
			opts := baseRunOptions()
			opts.Output = spec.OutputFormatText
			opts.EventsFile = eventsPath

			tc.setup(t, cfg, &opts)

			assertValidationFails(t, cfg, opts, tc.wantSub)
		})
	}
}

// assertValidationFails runs the spec and asserts the run aborted at input
// validation: it returned an error, emitted a single input_validation_failed
// event whose findings contain wantSub, never started the pipeline, and
// produced no helm/k8s/API side effects.
func assertValidationFails(t *testing.T, cfg *installconfig.Config, opts installconfig.RunOptions, wantSub string) {
	t.Helper()
	helm := &harness.FakeHelmInstaller{}
	k8s := &harness.FakeK8sExecutor{}
	apiClient, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}
	term, _, _ := newTestTerminal()

	err := specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "input validation failed")

	assert.Empty(t, helm.Calls, "no helm install must be attempted")
	assert.Empty(t, k8s.CallsByVerb("apply"), "no k8s writes must be attempted")
	assert.False(t, apiClient.ClusterCreated, "no cluster must be registered")

	evs := readEvents(t, opts.EventsFile)
	ev := findValidationEvent(evs)
	require.NotNil(t, ev, "an input_validation_failed event must be emitted")
	assert.False(t, hasPipelineEvent(evs, "pipeline_started", ""), "validation must abort before the pipeline starts")
	assert.Contains(t, findingMessages(ev), wantSub)
}

// TestInputValidation_AirgapFiles checks the air-gap file-existence rules: the
// bundle must exist for a normal run, and the working directory must exist when
// preparation is skipped.
func TestInputValidation_AirgapFiles(t *testing.T) {
	airgapOpts := func(t *testing.T) installconfig.RunOptions {
		opts := baseRunOptions()
		opts.Output = spec.OutputFormatText
		opts.EventsFile = filepath.Join(t.TempDir(), "events.ndjson")
		return opts
	}

	t.Run("missing bundle", func(t *testing.T) {
		cfg := newInstallRunConfig()
		cfg.Install.Connectivity = "air_gapped"
		cfg.Airgap = &installconfig.Airgap{
			BundlePath: filepath.Join(t.TempDir(), "absent.tar.gz"),
			WorkingDir: t.TempDir(),
		}
		assertValidationFails(t, cfg, airgapOpts(t), "airgap.bundlePath")
	})

	t.Run("skipPrepare missing working dir", func(t *testing.T) {
		cfg := newInstallRunConfig()
		cfg.Install.Connectivity = "air_gapped"
		cfg.Airgap = &installconfig.Airgap{
			WorkingDir:  filepath.Join(t.TempDir(), "absent-dir"),
			SkipPrepare: true,
		}
		assertValidationFails(t, cfg, airgapOpts(t), "airgap.workingDir")
	})
}

// TestInputValidation_RegistryHost covers the registry-host rules: air-gapped
// installs always need a mirror host (even with a token), and basic-auth or
// unauthenticated registries need a host too. Token auth on a connected install
// is exempt (covered implicitly by every other passing test).
func TestInputValidation_RegistryHost(t *testing.T) {
	validationOpts := func(t *testing.T) installconfig.RunOptions {
		opts := baseRunOptions()
		opts.Output = spec.OutputFormatText
		opts.EventsFile = filepath.Join(t.TempDir(), "events.ndjson")
		return opts
	}

	t.Run("airgap token registry without host", func(t *testing.T) {
		cfg := newInstallRunConfig()
		cfg.Install.Connectivity = "air_gapped"
		cfg.Airgap = &installconfig.Airgap{BundlePath: harness.CreateTestTarball(t), WorkingDir: t.TempDir()}
		opts := validationOpts(t)
		opts.Registry = &spec.Registry{Type: spec.RegistryTypeDockerConfig, Token: "tok"}
		assertValidationFails(t, cfg, opts, "registry.host")
	})

	t.Run("airgap with no registry", func(t *testing.T) {
		cfg := newInstallRunConfig()
		cfg.Install.Connectivity = "air_gapped"
		cfg.Airgap = &installconfig.Airgap{BundlePath: harness.CreateTestTarball(t), WorkingDir: t.TempDir()}
		opts := validationOpts(t)
		opts.Registry = nil
		assertValidationFails(t, cfg, opts, "registry.host")
	})

	t.Run("basic auth without host", func(t *testing.T) {
		cfg := newInstallRunConfig()
		opts := validationOpts(t)
		opts.Registry = &spec.Registry{Username: "admin", Password: "Harbor12345"}
		assertValidationFails(t, cfg, opts, "registry.host")
	})

	t.Run("unauthenticated registry without host", func(t *testing.T) {
		cfg := newInstallRunConfig()
		opts := validationOpts(t)
		opts.Registry = &spec.Registry{}
		assertValidationFails(t, cfg, opts, "registry.host")
	})

	t.Run("airgap with host passes", func(t *testing.T) {
		eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
		cfg, opts := airgapConfig(t, harness.CreateTestTarball(t), eventsPath)
		assertInstallSucceeds(t, cfg, opts, eventsPath)
	})

	// Token auth is rejected for air-gapped installs: the token never reaches
	// setup.sh, so a token-authenticated mirror would silently fail to push.
	// The mirror must use username/password instead.
	t.Run("airgap with token and host is rejected", func(t *testing.T) {
		cfg := newInstallRunConfig()
		cfg.Install.Connectivity = "air_gapped"
		cfg.Airgap = &installconfig.Airgap{BundlePath: harness.CreateTestTarball(t), WorkingDir: t.TempDir()}
		opts := validationOpts(t)
		opts.Registry = &spec.Registry{Type: spec.RegistryTypeDockerConfig, Token: "dG9r", Host: "registry.local:5000"}
		assertValidationFails(t, cfg, opts, "not a token")
	})

	// An unauthenticated mirror (host only, no credentials) is allowed for
	// air-gapped installs.
	t.Run("airgap unauthenticated with host passes", func(t *testing.T) {
		eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
		cfg, opts := airgapConfig(t, harness.CreateTestTarball(t), eventsPath)
		opts.Registry = &spec.Registry{Host: "registry.local:5000"}
		assertInstallSucceeds(t, cfg, opts, eventsPath)
	})
}

// TestInputValidation_WildcardCertificate_Valid covers the two ways a wildcard
// domain can be satisfied: a dedicated wildcard certificate, and falling back
// to a cluster certificate whose SAN already covers the wildcard. Both must
// pass validation and let the install proceed.
func TestInputValidation_WildcardCertificate_Valid(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)

	t.Run("dedicated wildcard cert", func(t *testing.T) {
		eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
		cfg, opts := connectedConfig(t, eventsPath, false)
		wcCert, wcKey := harness.GenCert(t, []string{"*.inference.test.runai.local"}, future)
		cfg.Cluster.Inference = &installconfig.Inference{
			WildcardDomain: "*.inference.test.runai.local",
			WildcardTLS:    &installconfig.TLS{CertPath: wcCert, KeyPath: wcKey},
		}
		assertInstallSucceeds(t, cfg, opts, eventsPath)
	})

	t.Run("fallback to cluster cert with wildcard SAN", func(t *testing.T) {
		eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
		cfg, opts := connectedConfig(t, eventsPath, false)
		cert, key := harness.GenCert(t, []string{"test.runai.local", "*.inference.test.runai.local"}, future)
		cfg.Cluster.TLS = &installconfig.TLS{CertPath: cert, KeyPath: key}
		cfg.Cluster.Inference = &installconfig.Inference{WildcardDomain: "*.inference.test.runai.local"}
		assertInstallSucceeds(t, cfg, opts, eventsPath)
	})
}

func assertInstallSucceeds(t *testing.T, cfg *installconfig.Config, opts installconfig.RunOptions, eventsPath string) {
	t.Helper()
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	assert.Nil(t, findValidationEvent(evs), "no input validation failures expected")
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
}

func TestInputValidation_RootTLSFailures(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)
	cases := []struct {
		name    string
		setup   func(*installconfig.Config)
		wantSub string
	}{
		{
			name: "missing certificate",
			setup: func(cfg *installconfig.Config) {
				cfg.TLS = &installconfig.TLS{CertPath: filepath.Join(t.TempDir(), "missing.crt"), KeyPath: filepath.Join(t.TempDir(), "missing.key")}
			},
			wantSub: "Cannot read certificate",
		},
		{
			name: "expired certificate",
			setup: func(cfg *installconfig.Config) {
				cert, key := harness.GenCert(t, []string{"test.runai.local", "*.test.runai.local"}, time.Now().Add(-time.Hour))
				cfg.TLS = &installconfig.TLS{CertPath: cert, KeyPath: key}
			},
			wantSub: "expired",
		},
		{
			name: "mismatched key",
			setup: func(cfg *installconfig.Config) {
				cert, _ := harness.GenCert(t, []string{"test.runai.local", "*.test.runai.local"}, future)
				_, key := harness.GenCert(t, []string{"test.runai.local", "*.test.runai.local"}, future)
				cfg.TLS = &installconfig.TLS{CertPath: cert, KeyPath: key}
			},
			wantSub: "do not match",
		},
		{
			name: "certificate without key",
			setup: func(cfg *installconfig.Config) {
				cert, _ := harness.GenCert(t, []string{"test.runai.local", "*.test.runai.local"}, future)
				cfg.TLS = &installconfig.TLS{CertPath: cert}
			},
			wantSub: "key",
		},
		{
			name: "key without certificate",
			setup: func(cfg *installconfig.Config) {
				_, key := harness.GenCert(t, []string{"test.runai.local", "*.test.runai.local"}, future)
				cfg.TLS = &installconfig.TLS{KeyPath: key}
			},
			wantSub: "paths are required",
		},
		{
			name: "control plane hostname mismatch",
			setup: func(cfg *installconfig.Config) {
				cfg.Install.Mode = "runai_control_plane_only"
				cfg.Cluster = nil
				cert, key := harness.GenCert(t, []string{"other.example.com"}, future)
				cfg.TLS = &installconfig.TLS{CertPath: cert, KeyPath: key}
			},
			wantSub: "Hostname mismatch",
		},
		{
			name: "cluster hostname mismatch",
			setup: func(cfg *installconfig.Config) {
				cfg.Install.Mode = "runai_cluster_only"
				cfg.ControlPlane = nil
				cfg.Cluster.Domain = "cluster.example.com"
				cert, key := harness.GenCert(t, []string{"other.example.com", "*.other.example.com"}, future)
				cfg.TLS = &installconfig.TLS{CertPath: cert, KeyPath: key}
			},
			wantSub: "Hostname mismatch",
		},
		{
			name: "inference wildcard mismatch",
			setup: func(cfg *installconfig.Config) {
				cert, key := harness.GenCert(t, []string{"test.runai.local", "*.test.runai.local"}, future)
				cfg.TLS = &installconfig.TLS{CertPath: cert, KeyPath: key}
				cfg.Cluster.Inference = &installconfig.Inference{WildcardDomain: "*.inference.test.runai.local"}
			},
			wantSub: "inference wildcard domain",
		},
		{
			name: "missing CA",
			setup: func(cfg *installconfig.Config) {
				cert, key := harness.GenCert(t, []string{"test.runai.local", "*.test.runai.local"}, future)
				cfg.TLS = &installconfig.TLS{
					CertPath:    cert,
					KeyPath:     key,
					CAChainPath: filepath.Join(t.TempDir(), "missing-ca.crt"),
				}
			},
			wantSub: "Path not accessible",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := newInstallRunConfig()
			opts := baseRunOptions()
			opts.Output = spec.OutputFormatText
			opts.EventsFile = filepath.Join(t.TempDir(), "events.ndjson")
			tc.setup(cfg)
			assertValidationFails(t, cfg, opts, tc.wantSub)
		})
	}
}

func TestInputValidation_RootTLSChecksEffectiveOverridesOnly(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)
	cpCert, cpKey := harness.GenCert(t, []string{"test.runai.local"}, future)
	clusterCert, clusterKey := harness.GenCert(t, []string{"test.runai.local", "*.test.runai.local"}, future)
	inferenceCert, inferenceKey := harness.GenCert(t, []string{"*.inference.test.runai.local"}, future)
	cfg := newInstallRunConfig()
	cfg.TLS = &installconfig.TLS{
		CertPath:    filepath.Join(t.TempDir(), "unused.crt"),
		KeyPath:     filepath.Join(t.TempDir(), "unused.key"),
		CAChainPath: filepath.Join(t.TempDir(), "unused-ca.crt"),
	}
	cfg.ControlPlane.TLS = &installconfig.TLS{CertPath: cpCert, KeyPath: cpKey}
	cfg.Cluster.TLS = &installconfig.TLS{CertPath: clusterCert, KeyPath: clusterKey}
	cfg.Cluster.Inference = &installconfig.Inference{
		WildcardDomain: "*.inference.test.runai.local",
		WildcardTLS:    &installconfig.TLS{CertPath: inferenceCert, KeyPath: inferenceKey},
	}
	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.EventsFile = eventsPath

	assertInstallSucceeds(t, cfg, opts, eventsPath)
}

// TestInputValidation_ClusterDomainFallback makes the cert_check fallback
// explicit: when cluster.domain is unset, the cluster certificate is validated
// against the control-plane domain (mirroring the plan's own fallback). The
// shared config leaves cluster.Domain empty while the control-plane domain is
// validationDomain.
func TestInputValidation_ClusterDomainFallback(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)

	t.Run("cluster cert covering the control-plane domain passes", func(t *testing.T) {
		eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
		cfg, opts := connectedConfig(t, eventsPath, false)
		require.Empty(t, cfg.Cluster.Domain, "fallback only applies when cluster domain is unset")
		cert, key := harness.GenCert(t, []string{validationDomain}, future)
		cfg.Cluster.TLS = &installconfig.TLS{CertPath: cert, KeyPath: key}
		assertInstallSucceeds(t, cfg, opts, eventsPath)
	})

	t.Run("cluster cert not covering the control-plane domain fails", func(t *testing.T) {
		eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
		cfg, opts := connectedConfig(t, eventsPath, false)
		require.Empty(t, cfg.Cluster.Domain, "fallback only applies when cluster domain is unset")
		opts.Output = spec.OutputFormatText
		cert, key := harness.GenCert(t, []string{"cluster-only.example.com"}, future)
		cfg.Cluster.TLS = &installconfig.TLS{CertPath: cert, KeyPath: key}
		// The mismatch names validationDomain, proving the cluster cert was
		// verified against the control-plane domain rather than skipped.
		assertValidationFails(t, cfg, opts, validationDomain)
	})
}

// TestInputValidation_AggregatesMultipleFailures verifies that independent
// problems are all reported in a single event rather than failing on the first.
func TestInputValidation_AggregatesMultipleFailures(t *testing.T) {
	cpCert, cpKey := harness.GenCert(t, []string{"wrong-cp.example.com"}, time.Now().Add(24*time.Hour))
	clCert, clKey := harness.GenCert(t, []string{"wrong-cluster.example.com"}, time.Now().Add(24*time.Hour))

	cfg := newInstallRunConfig()
	cfg.ControlPlane.TLS = &installconfig.TLS{CertPath: cpCert, KeyPath: cpKey}
	cfg.Cluster.TLS = &installconfig.TLS{CertPath: clCert, KeyPath: clKey}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath

	helm := &harness.FakeHelmInstaller{}
	deps := newTestDeps(helm)
	term, _, _ := newTestTerminal()

	err := specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "2 problem(s)")

	ev := findValidationEvent(readEvents(t, eventsPath))
	require.NotNil(t, ev)
	msgs := findingMessages(ev)
	assert.Contains(t, msgs, "controlPlane.tls.certPath")
	assert.Contains(t, msgs, "cluster.tls.certPath")
	assert.Empty(t, helm.Calls)
}
