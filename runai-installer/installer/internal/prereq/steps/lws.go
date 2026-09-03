/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/prereq/ids"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/helmclient"
	k8s "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient"
)

const LWSStepName = "Installing LeaderWorkerSet"

type LWSInstallStep struct {
	k8sClient  k8s.K8sClients
	crdChecker CRDChecker
	*step.RunAIInstallStepImpl
	version string
}

func NewLWSInstallStep(k8sClient k8s.K8sClients, version string) step.InstallStep {
	l := &LWSInstallStep{
		k8sClient:            k8sClient,
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl(ids.PrerequisiteIdentifierLWS),
		version:              version,
		crdChecker:           &DefaultCRDChecker{},
	}
	l.SetStatusEvaluator(l.EvaluateStatus)
	return l
}

func (l *LWSInstallStep) EvaluateStatus(runaiData *model.RunAIInstallData) {
	// Early return if already in terminal state (prevents race with timeout detection)
	if l.ShouldSkipStatusEvaluation() {
		return
	}

	timeout := l.GetTimeout()
	logging.Infof("[Prereq] Checking if LWS (LeaderWorkerSet) is installed...")

	// Check if the LWS CRD exists
	logging.Infof("[Prereq] Checking for LWS CRD 'leaderworkersets.leaderworkerset.x-k8s.io'...")
	crdInstalled := l.crdChecker.IsCRDInstalled(l.k8sClient.GetConfig(), "leaderworkersets.leaderworkerset.x-k8s.io")
	if !crdInstalled {
		elapsed := l.GetElapsedTime()
		if elapsed > 0 {
			logging.Infof("[Prereq] LWS CRD not yet ready (elapsed: %v, timeout: %v)", elapsed, timeout)
		} else {
			logging.Infof("[Prereq] LWS CRD not found")
		}
		return
	}
	logging.Infof("[Prereq] LWS CRD found")

	// Check if the lws-controller-manager deployment exists
	logging.Infof("[Prereq] Checking for LWS controller-manager deployment in namespace 'lws-system'...")
	err := ValidateComponentNamespace(context.Background(), l.k8sClient.GetClient(), "lws", "controller-manager", "lws-system")
	if err != nil {
		elapsed := l.GetElapsedTime()
		if elapsed > 0 {
			logging.Infof("[Prereq] LWS controller-manager not yet ready (elapsed: %v, timeout: %v)", elapsed, timeout)
		} else {
			logging.Infof("[Prereq] LWS controller-manager not found")
		}
		return
	}
	logging.Infof("[Prereq] LWS controller-manager found")

	logging.Infof("[Prereq] LWS is installed")
	l.SetStatus(model.Installed, nil)
}

func (l *LWSInstallStep) Execute(ctx context.Context, runaiData *model.RunAIInstallData) error {
	logging.Infof("[Prereq] Installing LWS (LeaderWorkerSet) v%s...", l.version)
	l.SetStatus(model.Installing, nil)

	err := helmclient.InstallOrUpgradeRelease(
		ctx,
		"oci://registry.k8s.io/lws/charts/lws",
		l.version,
		"lws",
		"lws",
		"lws-system",
		true,
		[]string{},
	)
	if err != nil {
		logging.Errorf("[Prereq] Failed to install LWS: %v", err)
		l.SetStatus(model.Failed, err)
		return err
	}

	logging.Infof("[Prereq] LWS Helm chart deployed, waiting for readiness")
	return nil
}
