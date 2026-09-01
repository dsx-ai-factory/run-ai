/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package initca

import (
	"context"
	"log"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/run-ai/runai/runai-cluster/cluster/init-ca/pkg/webhooks"
	"github.com/run-ai/runai/runai-cluster/common/cert-utils/pkg/certificates"

	cmv1 "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned/typed/certmanager/v1"
)

const (
	namespace                             = "runai"
	admissionControllerSecretName         = "runai-tls-secret"
	anyworkloadControllerSecretName       = "runai-anyworkload-controller-tls-secret"
	podGroupAssignerSecretName            = "runai-pod-group-assigner-tls-secret"        /* #nosec */
	accessruleControllerSecretName        = "runai-accessrule-controller-tls-secret"     /* #nosec */
	agentSecretName                       = "runai-agent-tls-secret"                     /* #nosec */
	applicationControllerSecretName       = "application-controller-tls-secret"          /* #nosec */
	adviceRestartControllerSecretName     = "runai-advice-restart-controller-tls-secret" /* #nosec */
	admissionControllerServiceName        = "runai-admission-controller"
	anyworkloadControllerServiceName      = "anyworkload-controller"
	podGroupAssignerServiceName           = "pod-group-assigner"
	accessruleControllerServiceName       = "accessrule-controller"
	agentServiceName                      = "runai-agent"
	adviceRestartControllerServiceName    = "advice-restart-controller"
	admissionControllerDeploymentName     = "runai-admission-controller"
	applicationControllerServiceName      = "application-controller"
	adviceRestartControllerDeploymentName = adviceRestartControllerServiceName
	anyworkloadControllerDeploymentName   = anyworkloadControllerServiceName
	podGroupAssignerDeploymentName        = podGroupAssignerServiceName
	accessruleControllerDeploymentName    = accessruleControllerServiceName
	applicationControllerDeploymentName   = applicationControllerServiceName
	agentDeploymentName                   = agentServiceName
)

func InitCA(ctx context.Context, conf *rest.Config, webhooksPath string) error {
	log.Printf("Run::AI init-ca")

	kbClient := kubernetes.NewForConfigOrDie(conf)
	cmv1Client := cmv1.NewForConfigOrDie(conf)

	deployVersion, err := certificates.GenerateRandomString()
	if err != nil {
		log.Printf("Failed to generate random string: %v", err)
		return err
	}
	log.Printf("Deployment version: %s", deployVersion)

	// We used to use cert-manager to create the certificate,
	// so we need to delete the old certificate so it won't regenerate the webhook tls secret.
	// After all clusters are upgraded to 2.10 we can remove this code.
	{
		if err = certificates.DeleteOldCertificates(ctx, cmv1Client, namespace); err != nil {
			log.Printf("Failed to delete old certificate: %v", err)
			return err
		}
	}

	log.Printf("Signing certificates")
	certSigner := certificates.NewSigner(conf)
	webhookCertificates, err := signCertificates(
		ctx, kbClient, certSigner)
	if err != nil {
		return err
	}

	return webhooks.ApplyWebhooks(
		ctx, kbClient,
		webhooksPath, deployVersion,
		webhookCertificates,
	)
}

func signCertificates(ctx context.Context, client kubernetes.Interface, certSigner certificates.CASigner) (
	webhookCertificates *webhooks.WebhookCertificates,
	err error,
) {
	webhookCertificates = &webhooks.WebhookCertificates{}

	if webhookCertificates.AdmissionControllerCertificateDetails, err = certSigner.Sign(
		ctx, admissionControllerServiceName, admissionControllerSecretName, namespace, true,
	); err != nil {
		log.Printf("Failed to sign certificate: %v", err)
		return webhookCertificates, err
	}
	log.Printf("Signed certificate for admission controller")

	if webhookCertificates.AnyworkloadControllerCertificateDetails, err = certSigner.Sign(
		ctx, anyworkloadControllerServiceName, anyworkloadControllerSecretName, namespace, true,
	); err != nil {
		log.Printf("Failed to sign certificate for anyworkload controller: %v", err)
		return webhookCertificates, err
	}
	log.Printf("Signed certificate for anyworkload controller")

	if webhookCertificates.PodGroupAssignerCertificateDetails, err = certSigner.Sign(
		ctx, podGroupAssignerServiceName, podGroupAssignerSecretName, namespace, true,
	); err != nil {
		log.Printf("Failed to sign certificate for pod-group-assigner: %v", err)
		return webhookCertificates, err
	}
	log.Printf("Signed certificate for pod-group-assigner")

	if webhookCertificates.AccessruleControllerCertificateDetails, err = certSigner.Sign(
		ctx, accessruleControllerServiceName, accessruleControllerSecretName, namespace, true,
	); err != nil {
		log.Printf("Failed to sign certificate for accessrule-controller: %v", err)
		return webhookCertificates, err
	}
	log.Printf("Signed certificate for accessrule-controller")

	if webhookCertificates.ApplicationControllerCertificateDetails, err = certSigner.Sign(
		ctx, applicationControllerServiceName, applicationControllerSecretName, namespace, true,
	); err != nil {
		log.Printf("Failed to sign certificate for application-controller: %v", err)
		return webhookCertificates, err
	}
	log.Printf("Signed certificate for application-controller")

	if webhookCertificates.AdviceRestartControllerCertificateDetails, err = certSigner.Sign(
		ctx, adviceRestartControllerServiceName, adviceRestartControllerSecretName, namespace, true,
	); err != nil {
		log.Printf("Failed to sign certificate for advice-restart-controller: %v", err)
		return webhookCertificates, err
	}
	log.Printf("Signed certificate for advice-restart-controller")

	if webhookCertificates.AgentCertificateDetails, err = certSigner.Sign(
		ctx, agentServiceName, agentSecretName, namespace, true,
	); err != nil {
		log.Printf("Failed to sign certificate for agent: %v", err)
		return webhookCertificates, err
	}
	log.Printf("Signed certificate for agent")

	log.Printf("Restarting admission controller")
	if err = rolloutRestartDeployment(ctx, client, admissionControllerDeploymentName, namespace); err != nil {
		log.Printf("Failed to restart deployment: %v", err)
		return webhookCertificates, err
	}

	log.Printf("Restarting anyworkload-controller")
	if err = rolloutRestartDeployment(ctx, client, anyworkloadControllerDeploymentName, namespace); err != nil {
		log.Printf("Failed to restart deployment: %v", err)
		return webhookCertificates, err
	}

	log.Printf("Restarting pod-group-assigner")
	if err = rolloutRestartDeployment(ctx, client, podGroupAssignerDeploymentName, namespace); err != nil {
		log.Printf("Failed to restart deployment: %v", err)
		return webhookCertificates, err
	}

	log.Printf("Restarting accessrule-controller")
	if err = rolloutRestartDeployment(ctx, client, accessruleControllerDeploymentName, namespace); err != nil {
		log.Printf("Failed to restart deployment: %v", err)
		return webhookCertificates, err
	}

	log.Printf("Restarting application-controller")
	if err = rolloutRestartDeployment(ctx, client, applicationControllerDeploymentName, namespace); err != nil {
		log.Printf("Failed to restart deployment: %v", err)
		return webhookCertificates, err
	}

	log.Printf("Restarting advice-restart-controller")
	if err = rolloutRestartDeployment(ctx, client, adviceRestartControllerDeploymentName, namespace); err != nil {
		log.Printf("Failed to restart deployment: %v", err)
		return webhookCertificates, err
	}

	log.Printf("Restarting agent")
	if err = rolloutRestartDeployment(ctx, client, agentDeploymentName, namespace); err != nil {
		log.Printf("Failed to restart deployment: %v", err)
		return webhookCertificates, err
	}

	return webhookCertificates, nil
}
