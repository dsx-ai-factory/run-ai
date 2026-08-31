/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/utils/helmclient"
	"github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	configinput "github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	"github.com/run-ai/runai/runai-installer/installer/internal/prereq/ids"
	k8s "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient"
	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const IngressStepName = "Installing NGINX Ingress Controller"

var (
	ingressNamespaces   = []string{"ingress-nginx", "nginx-ingress"}
	ingressServiceNames = []string{"ingress-nginx-controller", "nginx-ingress-controller"}
)

type NgnixIngressInstallStep struct {
	*step.RunAIInstallStepImpl
	k8sClient k8s.K8sClients
	version   string
}

func NewNgnixIngressInstallStep(k8sClient k8s.K8sClients, version string) step.InstallStep {
	n := &NgnixIngressInstallStep{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl(ids.PrerequisiteIdentifierIngress),
		k8sClient:            k8sClient,
		version:              version,
	}
	n.SetStatusEvaluator(n.EvaluateStatus)
	return n
}

func (n *NgnixIngressInstallStep) EvaluateStatus(runaiData *model.RunAIInstallData) {
	if n.ShouldSkipStatusEvaluation() {
		return
	}

	logging.Infof("[Prereq] Checking if NGINX Ingress is installed...")

	found := false
	for _, namespace := range ingressNamespaces {
		err := ValidateComponentNamespace(context.Background(), n.k8sClient.GetClient(), "nginx", "controller", namespace)
		if err == nil {
			found = true
			break
		}
	}

	if !found {
		elapsed := n.GetElapsedTime()
		timeout := n.GetTimeout()
		if elapsed > 0 {
			logging.Infof("[Prereq] NGINX Ingress not yet ready (elapsed: %v, timeout: %v)", elapsed, timeout)
		} else {
			logging.Infof("[Prereq] NGINX Ingress not found in cluster")
		}
		return
	}

	logging.Infof("[Prereq] NGINX Ingress is installed")
	n.SetStatus(model.Installed, nil)
}

func (n *NgnixIngressInstallStep) Execute(ctx context.Context, runaiData *model.RunAIInstallData) error {
	logging.Infof("[Prereq] Installing NGINX Ingress Controller v%s...", n.version)
	n.SetStatus(model.Installing, nil)
	values := n.getValuesForEnvironment(runaiData)
	logging.Infof("[Prereq] NGINX Ingress values: %v", values)
	err := helmclient.InstallOrUpgradeRelease(context.Background(),
		"https://kubernetes.github.io/ingress-nginx",
		n.version, "ingress-nginx", "ingress-nginx", "ingress-nginx", false, values)
	if err != nil {
		logging.Errorf("[Prereq] Failed to install NGINX Ingress: %v", err)
		n.SetStatus(model.Failed, err)
		return err
	}
	logging.Infof("[Prereq] NGINX Ingress Helm chart deployed, waiting for readiness")
	return nil
}

func (n *NgnixIngressInstallStep) PostInstall(runaiData *model.RunAIInstallData) {
	logging.Infof("[Prereq] Running post-install tasks for NGINX Ingress...")
	if err := n.getNodePortValues(runaiData); err != nil {
		logging.Errorf("[Prereq] Post-install failed for NGINX Ingress: %v", err)
		n.SetStatus(model.Failed, err)
	}
}

func (n *NgnixIngressInstallStep) getValuesForEnvironment(runaiData *model.RunAIInstallData) []string {
	httpsPort := runaiData.GetInputValue(inputs.InstallInputIngressControllerHTTPSPort)
	return n.getOnPremiseValues(httpsPort)
}

func (n *NgnixIngressInstallStep) getOnPremiseValues(httpsPort string) []string {
	values := []string{
		"controller.service.type=NodePort",
	}
	if httpsPort != "" {
		values = append(values, fmt.Sprintf("controller.service.nodePorts.https=%s", httpsPort))
	}
	return values
}

func (n *NgnixIngressInstallStep) getNodePortValues(runaiData *model.RunAIInstallData) error {
	ctx := context.Background()

	for i, namespace := range ingressNamespaces {
		serviceName := ingressServiceNames[i]

		svc := &v1.Service{}
		err := n.k8sClient.GetClient().Get(ctx, client.ObjectKey{Name: serviceName, Namespace: namespace}, svc)
		if err == nil {
			if svc.Spec.Type == v1.ServiceTypeNodePort {
				for _, port := range svc.Spec.Ports {
					if port.Name == "https" {
						runaiData.Inputs[configinput.InstallInputIngressControllerHTTPSPort] = fmt.Sprintf("%d", port.NodePort)
						return nil
					}
				}
			}
			return nil
		}
	}

	return fmt.Errorf("failed to find NGINX Ingress service in any known namespace: %v", ingressNamespaces)
}
