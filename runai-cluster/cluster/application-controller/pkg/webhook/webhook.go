/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package webhook

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/utils/validator"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/application-controller/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func RegisterAccessRuleWebhook(mgr manager.Manager) {
	configOpts := config.Config()

	serviceAccountValidator := NewServiceAccountValidator(mgr.GetClient(), configOpts.AllowedServiceAccounts)

	webhookServer := webhook.NewServer(webhook.Options{
		Port:    configOpts.WebhookPort,
		CertDir: configOpts.CertDir,
	})

	webhookServer.Register("/validate-application", &webhook.Admission{
		Handler: serviceAccountValidator,
	})

	err := mgr.Add(webhookServer)
	if err != nil {
		log.Error().Msgf("unable to setup validation webhook")
	}
}

type ServiceAccountValidator struct {
	serviceAccountValidator *validator.ServiceAccountValidator
}

func NewServiceAccountValidator(k8sClient client.Client, allowedServiceAccounts []string) *ServiceAccountValidator {
	return &ServiceAccountValidator{
		serviceAccountValidator: validator.NewServiceAccountValidator(k8sClient, allowedServiceAccounts, []validator.ValidatingFunction{}),
	}
}

func (arv *ServiceAccountValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	ctx = log.Logger.WithContext(ctx)
	reqKind := req.Kind.Kind
	reqUsername := req.UserInfo.Username
	log.Info().Msgf("executing validation handler - start, username: %s, kind: %s", req.UserInfo.Username, reqKind)

	serviceAccountValidatorResponse, err := arv.serviceAccountValidator.Validate(ctx, req)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("ServiceAccountValidator: failed validating request, error: <%s>", err.Error())
		return admission.Errored(http.StatusInternalServerError, err)
	}
	if !serviceAccountValidatorResponse {
		reason := fmt.Sprintf("username %s is not permitted to do manage %s", reqUsername, reqKind)
		log.Info().Msgf("executing validation handler - failed, reason: %s", reason)
		return admission.ValidationResponse(false, reason)
	}

	log.Info().Msgf("executing validation handler - allowed, username: %s, kind: %s", reqUsername, reqKind)
	return admission.ValidationResponse(true, "")
}
