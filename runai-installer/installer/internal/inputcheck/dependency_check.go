/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inputcheck

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
)

func init() {
	checks = append(checks, gpuTypeCheck, dynamoRequiresGrove, knativeRequiresInference, manifestFeaturesRejectHelmValues, manifestFeaturesRejectVersion, distributionCustomization, installPlanApprovalCheck)
}

// distributionCustomization enforces that each prerequisite/feature is customized
// through the surface that matches how it actually installs on the target
// distribution: customHelmValues for Helm installs, openshift.customResource for
// OperatorHub (OLM) installs that apply a managed custom resource. A value set
// for the wrong surface is silently ignored at install time, so it is surfaced
// here instead.
//
// The set of dependencies whose install method differs by distribution lives in
// one place (below), keyed off the spec's single-source-of-truth predicates, so
// adding the next OLM-on-OpenShift operator is a one-line change kept consistent
// with the planner.
func distributionCustomization(s *spec.Spec) []Finding {
	type entry struct {
		field    string
		dep      *spec.Dependency
		usesOLM  bool
		olmHasCR bool
	}

	var entries []entry
	if p := s.Prerequisites; p != nil {
		entries = append(entries,
			entry{"prerequisites.prometheus", p.Prometheus, false, false},
			entry{"prerequisites.gpuOperator", p.GPUOperator, s.GPUOperatorUsesOLM(), true},
			entry{"prerequisites.networkOperator", p.NetworkOperator, s.NetworkOperatorUsesOLM(), true},
		)
	}
	if f := s.Features; f != nil {
		// The NIM operator installs from Helm on every distribution (a
		// lightweight controller that runs under restricted-v2), so it is
		// configured with customHelmValues — the OpenShift OLM surfaces
		// (customResource, subscriptionConfig) do not apply.
		entries = append(entries,
			entry{"features.nimOperator", f.NIMOperator, false, false},
			// LWS installs via OperatorHub on OpenShift (the Red Hat Leader
			// Worker Set operator with a LeaderWorkerSetOperator CR), so its
			// OpenShift surfaces apply there; off OpenShift it is the Helm chart.
			entry{"features.lws", f.LWS, s.LWSUsesOLM(), true},
			// Knative installs via OperatorHub on OpenShift (the OpenShift
			// Serverless operator with a KnativeServing CR), so its OpenShift
			// surfaces apply there; off OpenShift it is the Helm flow.
			entry{"features.knative", f.Knative, s.KnativeUsesOLM(), true},
		)
	}

	var findings []Finding

	for _, e := range entries {
		if e.dep == nil {
			continue
		}
		if e.usesOLM && len(e.dep.CustomHelmValues) > 0 {
			findings = append(findings, Finding{
				Check:   "openshift-customization",
				Field:   e.field + ".customHelmValues",
				Message: "custom Helm values are not supported on OpenShift (the operator installs via OperatorHub/OLM)",
			})
		}
		if len(e.dep.OpenShiftCR) > 0 && (!e.usesOLM || !e.olmHasCR) {
			findings = append(findings, Finding{
				Check:   "openshift-customization",
				Field:   e.field + ".openshift.customResource",
				Message: "openshift.customResource only applies when the operator installs via OpenShift OperatorHub/OLM with a managed custom resource",
			})
		}
		if len(e.dep.OpenShiftSubscriptionEnv) > 0 && !e.usesOLM {
			findings = append(findings, Finding{
				Check:   "openshift-customization",
				Field:   e.field + ".openshift.subscriptionConfig",
				Message: "openshift.subscriptionConfig only applies when the operator installs via OpenShift OperatorHub/OLM",
			})
		}
		if e.dep.OpenShiftInstallPlanApproval != "" && !e.usesOLM {
			findings = append(findings, Finding{
				Check:   "openshift-customization",
				Field:   e.field + ".openshift.installPlanApproval",
				Message: "openshift.installPlanApproval only applies when the operator installs via OpenShift OperatorHub/OLM",
			})
		}
	}

	return findings
}

// installPlanApprovalCheck rejects an openshift.installPlanApproval that is
// neither "manual" nor "automatic". An unknown value would otherwise be treated
// as the default (manual) by the plan, masking a configuration typo.
func installPlanApprovalCheck(s *spec.Spec) []Finding {
	type entry struct {
		field string
		dep   *spec.Dependency
	}

	var entries []entry
	if p := s.Prerequisites; p != nil {
		entries = append(entries,
			entry{"prerequisites.gpuOperator", p.GPUOperator},
			entry{"prerequisites.networkOperator", p.NetworkOperator},
		)
	}
	if f := s.Features; f != nil {
		entries = append(entries, entry{"features.knative", f.Knative})
	}

	var findings []Finding
	for _, e := range entries {
		if e.dep == nil {
			continue
		}
		switch e.dep.OpenShiftInstallPlanApproval {
		case "", spec.OLMApprovalManual, spec.OLMApprovalAutomatic:
		default:
			findings = append(findings, Finding{
				Check:   "install-plan-approval",
				Field:   e.field + ".openshift.installPlanApproval",
				Message: fmt.Sprintf("invalid installPlanApproval %q: must be %q or %q", e.dep.OpenShiftInstallPlanApproval, spec.OLMApprovalManual, spec.OLMApprovalAutomatic),
			})
		}
	}

	return findings
}

// gpuTypeCheck rejects a GPU operator type that is neither "real" nor "fake".
// An unknown type would otherwise be treated as real by the plan, masking a
// configuration typo (e.g. "simulated").
func gpuTypeCheck(s *spec.Spec) []Finding {
	p := s.Prerequisites
	if p == nil || p.GPUOperator == nil {
		return nil
	}
	switch p.GPUType {
	case spec.GPUTypeReal, spec.GPUTypeFake:
		return nil
	}
	return []Finding{{
		Check:   "gpu-type",
		Field:   "prerequisites.gpuOperator.type",
		Message: fmt.Sprintf("invalid GPU operator type %q: must be %q or %q", p.GPUType, spec.GPUTypeReal, spec.GPUTypeFake),
	}}
}

// dynamoRequiresGrove rejects enabling Dynamo without Grove. The Dynamo step
// turns on Grove integration (global.grove.enabled=true) so the operator can
// orchestrate multinode deployments via PodCliqueSets; without the Grove
// operator (and its CRDs) installed, those reconciles fail. KAI is not checked
// here: the Run:ai cluster always provides the scheduler Grove needs.
func dynamoRequiresGrove(s *spec.Spec) []Finding {
	f := s.Features
	if f == nil || f.Skip || !f.Dynamo.IsInstall() {
		return nil
	}
	if f.Grove.IsInstall() {
		return nil
	}
	return []Finding{{
		Check:   "dynamo-requires-grove",
		Field:   "features.dynamo",
		Message: "Dynamo requires Grove: set features.grove.mode=install (the Run:ai cluster already provides the KAI scheduler Grove needs)",
	}}
}

// knativeRequiresInference rejects enabling Knative Serving without the inputs
// its KnativeServing CR and front-door are built from. A wildcard domain
// (cluster.inference.wildcardDomain) is always required (it is the front-door's
// host). A wildcard certificate is required only for the in-cluster Ingress
// front-door, which terminates TLS itself with the secret. Two front-doors do
// not need it: under Gateway API the Gateway listener terminates TLS, and on
// OpenShift serving is exposed through OpenShift Routes (the cluster/router
// cert), so neither path consumes a spec-provided wildcard certificate. The
// certificate is resolved like the install step does (cluster.inference.wildcardTls,
// falling back to cluster.tls); its contents are validated separately by
// certificateCheck when one is provided.
func knativeRequiresInference(s *spec.Spec) []Finding {
	f := s.Features
	if f == nil || f.Skip || !f.Knative.IsInstall() {
		return nil
	}
	c := s.Cluster
	if c == nil || c.Inference == nil || c.Inference.WildcardDomain == "" {
		return []Finding{{
			Check:   "knative-requires-inference",
			Field:   "features.knative",
			Message: "Knative requires an inference wildcard domain: set cluster.inference.wildcardDomain",
		}}
	}
	if !isGatewayAPI(s) && !s.IsOpenShift() && !knativeWildcardCertAvailable(s) {
		return []Finding{{
			Check:   "knative-requires-certificate",
			Field:   "features.knative",
			Message: "Knative requires a wildcard certificate: set cluster.inference.wildcardTls, or a cluster certificate (cluster.tls) whose SAN covers the wildcard domain",
		}}
	}
	return nil
}

// knativeWildcardCertAvailable reports whether a certificate is configured to
// serve the inference wildcard domain. It mirrors the install step's resolution
// (the dedicated wildcard certificate when set, otherwise the cluster
// certificate it falls back to).
func knativeWildcardCertAvailable(s *spec.Spec) bool {
	if s.SelfSignedTLSEnabled() {
		return true
	}
	tls := s.InferenceTLS()
	return tls != nil && tls.CertPath != ""
}

// manifestFeaturesRejectHelmValues rejects customHelmValues on the manifest-based
// features (Kubeflow training operator, MPI operator). They are installed from
// kustomize/raw manifests, not Helm, so the values would be silently ignored —
// a misconfiguration we surface rather than swallow.
func manifestFeaturesRejectHelmValues(s *spec.Spec) []Finding {
	f := s.Features
	if f == nil {
		return nil
	}
	var findings []Finding
	if d := f.Kubeflow; d != nil && len(d.CustomHelmValues) > 0 {
		findings = append(findings, Finding{
			Check:   "kubeflow-no-helm-values",
			Field:   "features.kubeflow.customHelmValues",
			Message: "custom Helm values are not supported for the Kubeflow training operator (it is installed from manifests, not Helm)",
		})
	}
	if d := f.MPIOperator; d != nil && len(d.CustomHelmValues) > 0 {
		findings = append(findings, Finding{
			Check:   "mpi-operator-no-helm-values",
			Field:   "features.mpiOperator.customHelmValues",
			Message: "custom Helm values are not supported for the MPI operator (it is installed from manifests, not Helm)",
		})
	}
	return findings
}

// manifestFeaturesRejectVersion rejects a custom version on the manifest-based
// features (Kubeflow training operator, MPI operator). They are pinned to the
// version in the compatibility matrix and installed from version-specific
// upstream manifest URLs, so an arbitrary override is not supported: the matrix
// lock is always used.
func manifestFeaturesRejectVersion(s *spec.Spec) []Finding {
	f := s.Features
	if f == nil {
		return nil
	}
	var findings []Finding
	if d := f.Kubeflow; d != nil && d.Version != "" {
		findings = append(findings, Finding{
			Check:   "kubeflow-no-custom-version",
			Field:   "features.kubeflow.version",
			Message: "custom version is not supported for the Kubeflow training operator (it is pinned to the compatibility matrix version)",
		})
	}
	if d := f.MPIOperator; d != nil && d.Version != "" {
		findings = append(findings, Finding{
			Check:   "mpi-operator-no-custom-version",
			Field:   "features.mpiOperator.version",
			Message: "custom version is not supported for the MPI operator (it is pinned to the compatibility matrix version)",
		})
	}
	return findings
}
