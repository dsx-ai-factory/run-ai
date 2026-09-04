/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/actions"
)

const (
	openshiftMarketplaceNamespace = "openshift-marketplace"
	certifiedOperatorsSource      = "certified-operators"
	redhatOperatorsSource         = "redhat-operators"

	nfdNamespace    = "openshift-nfd"
	nfdPackage      = "nfd"
	nfdSentinel     = "nfd-controller-manager"
	nfdCRD          = "nodefeaturediscoveries.nfd.openshift.io"
	nfdInstanceName = "nfd-instance"
	nfdCRAPIVersion = "nfd.openshift.io/v1"
	nfdCRKind       = "NodeFeatureDiscovery"
)

// CertifiedOperatorsCatalog is the OperatorHub catalog NVIDIA's certified
// operators (GPU, network, NIM) are published to.
func CertifiedOperatorsCatalog() actions.OLMCatalog {
	return actions.OLMCatalog{Source: certifiedOperatorsSource, SourceNamespace: openshiftMarketplaceNamespace}
}

// RedHatOperatorsCatalog is the OperatorHub catalog the Red Hat operators (NFD,
// SR-IOV, cert-manager, LWS) are published to.
func RedHatOperatorsCatalog() actions.OLMCatalog {
	return actions.OLMCatalog{Source: redhatOperatorsSource, SourceNamespace: openshiftMarketplaceNamespace}
}

// startingCSV builds an OLM Subscription startingCSV from a package name and a
// resolved version (e.g. "gpu-operator-certified" + "v25.3.0" ->
// "gpu-operator-certified.v25.3.0"), pinning the initial installed version. An
// empty version yields an empty startingCSV, letting OLM track the channel head.
func startingCSV(pkg, version string) string {
	if version == "" {
		return ""
	}

	return pkg + "." + version
}

// newNFDSteps installs the Node Feature Discovery operator and its
// NodeFeatureDiscovery CR. NFD labels nodes with hardware features (PCI devices,
// kernel, etc.) that the GPU and network operators select on, so both flows
// reuse it; the steps' Detect makes a second install idempotent when both the
// GPU and network operators are enabled.
func newNFDSteps(channel, installPlanApproval string, catalog actions.OLMCatalog, deps orchestrator2.Deps) []orchestrator2.Step {
	operator := NewOLMOperator(OLMOperator{
		StepName:            "install-nfd-operator",
		Namespace:           nfdNamespace,
		Package:             nfdPackage,
		Channel:             channel,
		InstallPlanApproval: installPlanApproval,
		Catalog:             catalog,
		Sentinel:            deploymentSentinel(nfdSentinel, nfdNamespace),
	}, deps)

	instance := NewOperatorDefaultCR(OperatorCR{
		StepName:         "create-node-feature-discovery",
		CRD:              nfdCRD,
		CRAPIVersion:     nfdCRAPIVersion,
		CRKind:           nfdCRKind,
		CRName:           nfdInstanceName,
		CRNamespace:      nfdNamespace,
		CSVNamespace:     nfdNamespace,
		SubscriptionName: nfdPackage,
	}, deps)

	return []orchestrator2.Step{operator, instance}
}
