/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validator

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/utils/validator"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SecretPvcValidator validates Secret and PVC resources labeled with run.ai/asset to ensure only Run:AI-managed identities can mutate them.
//
// +kubebuilder:webhook:path=/validate-secrets-pvc,mutating=false,failurePolicy=ignore,sideEffects=None,groups=*,resources=secrets;persistentvolumeclaims,verbs=create;update;delete,versions=*,name=vsecret-pvc-validator.kb.io,admissionReviewVersions=v1
//
//nolint:lll // kubebuilder annotation
type SecretPvcValidator struct {
	serviceAccountValidator *validator.ServiceAccountValidator
}

func NewSecretPvcValidator(k8sClient client.Client, allowedServiceAccounts []string) *SecretPvcValidator {
	return &SecretPvcValidator{
		serviceAccountValidator: validator.NewServiceAccountValidator(k8sClient, allowedServiceAccounts, []validator.ValidatingFunction{}),
	}
}

func (v *SecretPvcValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	ctx = log.Logger.WithContext(ctx)
	reqKind := req.Kind.Kind
	reqUsername := req.UserInfo.Username
	log.Info().Msgf("executing validation handler - start, username: %s, kind: %s", reqUsername, reqKind)

	serviceAccountValidatorResponse, err := v.serviceAccountValidator.Validate(ctx, req)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("SecretPvcValidator: failed validating request, error: <%s>", err.Error())
		return admission.Errored(http.StatusInternalServerError, err)
	}
	if !serviceAccountValidatorResponse {
		reason := fmt.Sprintf("username %s is not permitted to do any action on %s with the run.ai/asset label", reqUsername, reqKind)
		log.Info().Msgf("executing validation handler - failed, reason: %s", reason)
		return admission.ValidationResponse(false, reason)
	}
	log.Info().Msgf("executing validation handler - allowed, username: %s, kind: %s", reqUsername, reqKind)
	return admission.ValidationResponse(true, "")
}
