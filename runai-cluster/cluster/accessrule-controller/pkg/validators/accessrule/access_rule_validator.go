/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package accessrule

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/common/utils/validator"
	"net/http"

	"github.com/run-ai/runai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/common"
	"github.com/run-ai/runai/runai-cluster/cluster/accessrule-controller/pkg/utils"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/validate-access-rule,mutating=false,failurePolicy=fail,sideEffects=None,groups=run.ai,resources=accessrules,verbs=create,update,delete,versions=*,name=vaccessrule.kb.io,admissionReviewVersions=v1
type AccessRuleValidator struct {
	serviceAccountValidator *validator.ServiceAccountValidator
}

func NewAccessRuleValidator(k8sClient client.Client, allowedServiceAccounts []string) *AccessRuleValidator {
	return &AccessRuleValidator{
		serviceAccountValidator: validator.NewServiceAccountValidator(k8sClient, allowedServiceAccounts, initValidator()),
	}
}

func initValidator() []validator.ValidatingFunction {
	// Initialize the service account validator with the required UIDs
	return []validator.ValidatingFunction{
		{Func: getRunaiAgentUID, Id: ""},
		{Func: getAccessRuleControllerUID, Id: ""},
		{Func: getRunaiOperatorUID, Id: ""},
	}
}

func (arv *AccessRuleValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	serviceAccountValidatorResponse, err := arv.serviceAccountValidator.Validate(ctx, req)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("AccessRuleValidator: failed validating request, error: <%s>", err.Error())
		return admission.Errored(http.StatusInternalServerError, err)
	}
	if !serviceAccountValidatorResponse {
		log.Ctx(ctx).Warn().Msgf("AccessRuleValidator: unauthorized user <%s/%s> tried to %v an access rule",
			req.UserInfo.UID, req.UserInfo.Username, req.Operation)
		return admission.ValidationResponse(false, "operation not allowed")
	}
	return admission.ValidationResponse(true, "")
}

func (arv *AccessRuleValidator) InjectDecoder(d *admission.Decoder) error {
	return nil
}

func getRunaiAgentUID(ctx context.Context, client client.Client) (string, error) {
	runaiAgentServiceAccount, err := utils.GetServiceAccount(
		ctx, client,
		common.RunaiAgentServiceAccountName, runaiv1.DefaultNamespace)
	if err != nil {
		return "", err
	}

	return string(runaiAgentServiceAccount.UID), nil
}

func getAccessRuleControllerUID(ctx context.Context, client client.Client) (string, error) {
	runaiAccessRuleControllerServiceAccount, err := utils.GetServiceAccount(
		ctx, client,
		common.RunaiAccessRuleControllerServiceAccountName, runaiv1.DefaultNamespace)
	if err != nil {
		return "", err
	}

	return string(runaiAccessRuleControllerServiceAccount.UID), nil
}

func getRunaiOperatorUID(ctx context.Context, client client.Client) (string, error) {
	runaiOperatorServiceAccount, err := utils.GetServiceAccount(
		ctx, client,
		common.RunaiOperatorServiceAccountName, runaiv1.DefaultNamespace)
	if err != nil {
		return "", err
	}

	return string(runaiOperatorServiceAccount.UID), nil
}
