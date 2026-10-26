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
	certManagerOperatorNamespace = "cert-manager-operator"
	certManagerOperatorPackage   = "openshift-cert-manager-operator"
	certManagerOperatorSentinel  = "cert-manager-operator-controller-manager"

	// LWSOperatorNamespace is where the Red Hat Leader Worker Set operator runs
	// its LWS controller; the Run:ai cluster must be pointed at it.
	LWSOperatorNamespace    = "openshift-lws-operator"
	lwsOperatorPackage      = "leader-worker-set"
	lwsOperatorSentinel     = "openshift-lws-operator"
	lwsOperatorCRD          = "leaderworkersetoperators.operator.openshift.io"
	lwsOperatorCRAPIVersion = "operator.openshift.io/v1"
	lwsOperatorCRKind       = "LeaderWorkerSetOperator"
	lwsOperatorCRName       = "cluster"
)

// LWSOpenShiftParams configures the OpenShift LWS flow. Channels and the
// LeaderWorkerSetOperator override are resolved by the planner.
type LWSOpenShiftParams struct {
	Channel string
	// CertManagerChannel is the channel of the cert-manager Operator for Red Hat
	// OpenShift, which the Leader Worker Set operator requires for its webhook
	// certificates.
	CertManagerChannel  string
	InstallPlanApproval string
	CRSpec              map[string]any
	SubscriptionEnv     map[string]string
	Catalog             actions.OLMCatalog
}

// NewLWSOpenShift installs LWS on OpenShift via the Red Hat Leader Worker Set
// operator (OperatorHub): the cert-manager operator it depends on, the operator
// itself, and its LeaderWorkerSetOperator CR built from the operator's CSV
// default with any user override merged on top.
func NewLWSOpenShift(p LWSOpenShiftParams, deps orchestrator2.Deps) []orchestrator2.Step {
	certManager := NewOLMOperator(OLMOperator{
		StepName:            "install-cert-manager-operator",
		Namespace:           certManagerOperatorNamespace,
		Package:             certManagerOperatorPackage,
		Channel:             p.CertManagerChannel,
		InstallPlanApproval: p.InstallPlanApproval,
		Catalog:             p.Catalog,
		Sentinel:            deploymentSentinel(certManagerOperatorSentinel, certManagerOperatorNamespace),
		AllNamespaces:       true,
	}, deps)

	operator := NewOLMOperator(OLMOperator{
		StepName:            "install-lws-operator",
		Namespace:           LWSOperatorNamespace,
		NamespaceLabels:     map[string]string{clusterMonitoringLabel: "true"},
		Package:             lwsOperatorPackage,
		Channel:             p.Channel,
		InstallPlanApproval: p.InstallPlanApproval,
		Catalog:             p.Catalog,
		Sentinel:            deploymentSentinel(lwsOperatorSentinel, LWSOperatorNamespace),
		SubscriptionEnv:     p.SubscriptionEnv,
	}, deps)

	instance := NewOperatorDefaultCR(OperatorCR{
		StepName:         "create-leader-worker-set-operator",
		CRD:              lwsOperatorCRD,
		CRAPIVersion:     lwsOperatorCRAPIVersion,
		CRKind:           lwsOperatorCRKind,
		CRName:           lwsOperatorCRName,
		CRNamespace:      LWSOperatorNamespace,
		CSVNamespace:     LWSOperatorNamespace,
		SubscriptionName: lwsOperatorPackage,
		Override:         p.CRSpec,
	}, deps)

	return []orchestrator2.Step{certManager, operator, instance}
}
