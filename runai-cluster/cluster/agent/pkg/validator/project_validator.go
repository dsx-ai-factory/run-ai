/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validator

import (
	"context"
	"github.com/rs/zerolog/log"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/utils/validator"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// ProjectValidator validates project and department resources.
//
// +kubebuilder:webhook:path=/validate-project,mutating=false,failurePolicy=fail,sideEffects=None,groups=run.ai;scheduling.run.ai,resources=projects;departments,verbs=create;update;delete,versions=*,name=vproject-validator.kb.io,admissionReviewVersions=v1
//
//nolint:lll // kubebuilder annotation
type ProjectValidator struct {
	garbageCollectorReader  client.Reader
	serviceAccountValidator *validator.ServiceAccountValidator
}

// NewProjectValidator uses an uncached reader for the collector lookup so it only needs get permission.
func NewProjectValidator(k8sClient client.Client, garbageCollectorReader client.Reader, allowedServiceAccounts []string) *ProjectValidator {
	pv := &ProjectValidator{
		garbageCollectorReader: garbageCollectorReader,
	}
	pv.serviceAccountValidator = validator.NewServiceAccountValidator(k8sClient, allowedServiceAccounts, pv.initValidator())
	return pv
}

func (pv *ProjectValidator) initValidator() []validator.ValidatingFunction {
	// Initialize the service account validator with the required UIDs
	return []validator.ValidatingFunction{
		{Func: pv.getRunaiAgentUID, Id: ""},
		{Func: pv.getRunaiClusterSyncUID, Id: ""},
		{Func: pv.getRunaiOperatorUID, Id: ""},
		{Func: pv.getProjectControllerUID, Id: ""},
	}
}

func (pv *ProjectValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	// Orphan deletion completes through a garbage-collector UPDATE, after the
	// agent's DELETE has already been authorized.
	allowedCleanup, err := pv.allowOrphanFinalizerRemoval(ctx, req)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	if allowedCleanup {
		return admission.Allowed("")
	}
	serviceAccountValidatorResponse, err := pv.serviceAccountValidator.Validate(ctx, req)
	if err != nil {
		log.Error().Msgf("ProjectValidator: failed validating request, err: %v", err)
		return admission.Errored(http.StatusInternalServerError, err)
	}
	if !serviceAccountValidatorResponse {
		log.Info().Msgf("ProjectValidator: unauthorized user UserUID %s UserName %s Operation %s", req.UserInfo.UID, req.UserInfo.Username, req.Operation)
		return admission.ValidationResponse(false, "operation not allowed")
	}
	return admission.ValidationResponse(true, "")
}

func (pv *ProjectValidator) getRunaiAgentUID(ctx context.Context, k8sClient client.Client) (string, error) {
	runaiAgentServiceAccount, err := GetServiceAccount(
		ctx, k8sClient,
		constants.RunaiAgentName, runaiv1.DefaultNamespace)
	if err != nil {
		return "", err
	}

	return string(runaiAgentServiceAccount.UID), nil
}

func (pv *ProjectValidator) getRunaiClusterSyncUID(ctx context.Context, k8sClient client.Client) (string, error) {
	runaiClusterSyncServiceAccount, err := GetServiceAccount(
		ctx, k8sClient,
		constants.RunaiClusterSyncName, runaiv1.DefaultNamespace)
	if err != nil {
		return "", err
	}

	return string(runaiClusterSyncServiceAccount.UID), nil
}

// getProjectControllerUID resolves KRM's project-controller, which replaced the
// Run:ai-built one; its service account is named without the runai- prefix.
func (pv *ProjectValidator) getProjectControllerUID(ctx context.Context, k8sClient client.Client) (string, error) {
	projectControllerServiceAccount, err := GetServiceAccount(
		ctx, k8sClient,
		constants.KRMProjectControllerName, runaiv1.DefaultNamespace)
	if err != nil {
		return "", err
	}

	return string(projectControllerServiceAccount.UID), nil
}

func (pv *ProjectValidator) getRunaiOperatorUID(ctx context.Context, k8sClient client.Client) (string, error) {
	runaiOperatorServiceAccount, err := GetServiceAccount(
		ctx, k8sClient,
		constants.RunaiOperatorName, runaiv1.DefaultNamespace)
	if err != nil {
		return "", err
	}

	return string(runaiOperatorServiceAccount.UID), nil
}
