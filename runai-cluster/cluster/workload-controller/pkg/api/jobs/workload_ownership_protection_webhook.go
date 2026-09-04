/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package jobs

import (
	"context"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	authv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	workloadproject "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/project"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/common"
	pkgcommon "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/webhook"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	authenticationv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/fields"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/validate-workload-ownership-protection,mutating=false,failurePolicy=fail,sideEffects=None,groups=run.ai,resources=trainingworkloads;interactiveworkloads;deploymentworkloads;inferenceworkloads,verbs=delete;update,versions=v2alpha1,name=runai-workload-ownership-protection-validation.kb.io,admissionReviewVersions=v1

const (
	RunaiAgentServiceAccountName = operands.RunaiAgentName

	// GVK of the kai.resources Project, read as unstructured to avoid a typed kai SDK
	// (and thus KAI-scheduler) dependency in this package — see departmentOfNamespace.
	kaiProjectGroup   = "kai.resources"
	kaiProjectVersion = "v1alpha1"
	kaiProjectKind    = "Project"
)

// Note: WorkloadWrapper interface and WrapperCreator type are defined in workload_ownership_webhook.go
// We're reusing the same abstractions for consistency across both webhooks

type WorkloadOwnershipProtectionValidator struct {
	client.Client
	workloadControllerUID         string
	runtimeRulesServiceAccountUID string
	agentId                       string
	createWrapper                 WrapperCreator
}

type ValidationRequestInfo struct {
	WorkloadNamespace string
	WorkloadOwnerUID  string
	RequesterUID      string
	RequesterGroups   []string
}

func NewWorkloadOwnershipProtectionValidator(k8sClient client.Client, createWrapper WrapperCreator, workloadControllerUID string) *WorkloadOwnershipProtectionValidator {
	return &WorkloadOwnershipProtectionValidator{
		Client:                k8sClient,
		createWrapper:         createWrapper,
		workloadControllerUID: workloadControllerUID,
	}
}

func (wopv *WorkloadOwnershipProtectionValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	wrapper := wopv.createWrapper()

	logger := logf.FromContext(ctx)
	err := wrapper.Unmarshal(req.OldObject.Raw)
	if err != nil {
		logger.Error(err, "Failed to unmarshal workload",
			"operation", req.Operation, "kind", req.Kind, "namespace", req.Namespace, "name", req.Name)
		return admission.Errored(http.StatusBadRequest, err)
	}

	if wrapper.GetUsage() != common.UsedForSubmit {
		// if it's a workload for templates - don't validate UserId
		return admission.ValidationResponse(true, "")
	}

	logger.V(5).Info("WorkloadOwnershipProtectionValidator: handling validation of workload",
		"namespace", wrapper.GetNamespace(), "name", wrapper.GetName(), "requestOperation", req.Operation)

	allowed, reason, err := wopv.validateWebhookRequest(ctx, wrapper, req.UserInfo)
	if err != nil {
		logger.Error(err, "WorkloadOwnershipProtectionValidator: failed validating request")
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.ValidationResponse(allowed, reason)
}

func (wopv *WorkloadOwnershipProtectionValidator) InjectDecoder(d *admission.Decoder) error {
	return nil
}

func (wopv *WorkloadOwnershipProtectionValidator) validateWebhookRequest(
	ctx context.Context,
	wrapper WorkloadWrapper,
	admissionRequestUserInfo authenticationv1.UserInfo) (bool, string, error) {
	workloadOwnerUID := wrapper.GetOwnerUID()
	if workloadOwnerUID == "" {
		return true, "", nil
	}

	runtimeRulesServiceAccountUID, err := wopv.getCachedRuntimeRulesServiceAccountUID(ctx)
	if err != nil {
		return false, "", err
	}

	requesterUID := webhook.GetRequesterUserId(admissionRequestUserInfo)
	requesterGroups := admissionRequestUserInfo.Groups

	// some requests are being made by internal system components (workload-controller,
	// runtime-rules). we want to allow them before having to perform other api calls.
	if requesterUID == wopv.workloadControllerUID ||
		requesterUID == runtimeRulesServiceAccountUID {
		return true, "", nil
	}

	validationRequestInfo := ValidationRequestInfo{
		WorkloadNamespace: wrapper.GetNamespace(),
		WorkloadOwnerUID:  workloadOwnerUID,
		RequesterUID:      requesterUID,
		RequesterGroups:   requesterGroups,
	}
	return wopv.validateRequestInner(ctx, validationRequestInfo)
}

func (wopv *WorkloadOwnershipProtectionValidator) ValidateResearcherServiceRequest(
	ctx context.Context,
	validateRequestInfo ValidationRequestInfo) (bool, string, error) {
	return wopv.validateRequestInner(ctx, validateRequestInfo)
}

func (wopv *WorkloadOwnershipProtectionValidator) validateRequestInner(
	ctx context.Context,
	validationRequestInfo ValidationRequestInfo) (bool, string, error) {
	requesterUID := validationRequestInfo.RequesterUID

	if requesterUID == validationRequestInfo.WorkloadOwnerUID {
		return true, "", nil
	}

	departmentName, err := wopv.departmentOfNamespace(ctx, validationRequestInfo.WorkloadNamespace)
	if err != nil {
		return false, "", err
	}

	requesterGroupsSet := utils.NewSet[string](validationRequestInfo.RequesterGroups...)

	doesHaveProjectCreationPermissionsInRelevantScope, err :=
		wopv.doesUserHaveProjectCreationPermissionsInRelevantScope(ctx, departmentName,
			requesterUID, validationRequestInfo.RequesterGroups, requesterGroupsSet)
	if err != nil {
		log.Error(err, "Failed validating whether user has workload crud permissions in higher scope")
		return false, "", err
	}
	if doesHaveProjectCreationPermissionsInRelevantScope {
		log.V(6).Info(
			"WorkloadOwnershipProtectionValidator: authorizing request for user with project creation permissions in a relevant scope",
			"requesterUID", requesterUID, "workloadOwnerUID", validationRequestInfo.WorkloadOwnerUID)
		return true, "", nil
	}

	userIsAgent, err := wopv.isAgent(ctx, requesterUID)
	if err != nil {
		log.Error(err, "WorkloadOwnershipProtectionValidator: Failed to get if user is agent")
		return false, "", err
	}
	if userIsAgent {
		log.V(6).Info("WorkloadOwnershipProtectionValidator: authorizing request for agent",
			"requesterUID", requesterUID, "workloadOwnerUID", validationRequestInfo.WorkloadOwnerUID)
		return true, "", nil
	}

	userPermittedForNamespace, err := wopv.userPermittedForNamespace(ctx, requesterUID)
	if err != nil {
		log.Error(err, "WorkloadOwnershipProtectionValidator: Failed to get if user permitted for namespace")
		return false, "", err
	}
	if userPermittedForNamespace {
		log.V(6).Info("WorkloadOwnershipProtectionValidator: authorizing request for non-runai authenticated user",
			"requesterUID", requesterUID, "workloadOwnerUID", validationRequestInfo.WorkloadOwnerUID, "requesterGroups", validationRequestInfo.RequesterGroups)
		return true, "", nil
	}

	log.V(6).Info("WorkloadOwnershipProtectionValidator: operation not allowed",
		"requesterUID", requesterUID, "workloadOwnerUID", validationRequestInfo.WorkloadOwnerUID)
	return false, "operation not allowed", nil
}

func (wopv *WorkloadOwnershipProtectionValidator) isAgent(ctx context.Context, requesterUID string) (bool, error) {
	if wopv.agentId == "" {
		agentUID, err := wopv.getRunaiAgentUID(ctx)
		if err != nil {
			return false, err
		}
		wopv.agentId = agentUID
	}
	return wopv.agentId == requesterUID, nil
}

func (wopv *WorkloadOwnershipProtectionValidator) getRunaiAgentUID(ctx context.Context) (string, error) {
	serviceAccount := &corev1.ServiceAccount{}
	serviceAccountKey := types.NamespacedName{
		Name:      RunaiAgentServiceAccountName,
		Namespace: runaiv1.DefaultNamespace,
	}
	err := wopv.Client.Get(ctx, serviceAccountKey, serviceAccount)
	if err != nil {
		return "", err
	}
	return string(serviceAccount.UID), nil
}

// this feature is designed to prevent other runai-users from deleting/updating their fellow's workloads.
// if it's kubeadmin, or just a k8s user that is not authenticated with runai,
// but has general RBAC to preform those operations - we will allow it.
func (wopv *WorkloadOwnershipProtectionValidator) userPermittedForNamespace(ctx context.Context, requesterUID string) (bool, error) {
	sar := &authv1.SubjectAccessReview{
		Spec: authv1.SubjectAccessReviewSpec{
			User: requesterUID,
			ResourceAttributes: &authv1.ResourceAttributes{
				Verb:     "create",
				Resource: "namespaces",
			},
		},
	}
	err := wopv.Client.Create(ctx, sar, &client.CreateOptions{DryRun: []string{"All"}})
	if err != nil {
		return false, err
	}

	return sar.Status.Allowed, nil
}

// doesUserHaveProjectCreationPermissionsInRelevantScope - lists all relevant access rules using the field selector
// (the manager will cache accessrules that are department or cluster scoped and have project creation permissions);
// and validates that at least one of their scope is relevant for the current request
// (if department scoped - that it is indeed the current project's department scope).
// departmentOfNamespace returns the department (parent) name of the project that owns the
// namespace, read from the kai.resources project's spec.parent.
//
// The kai project is read as unstructured on purpose: this package is imported by
// control-plane services (e.g. workloads) that pin an older controller-runtime, and the
// typed kai SDK transitively pulls in KAI-scheduler, which fails to compile there. Reading
// unstructured keeps that dependency out of this package.
func (wopv *WorkloadOwnershipProtectionValidator) departmentOfNamespace(ctx context.Context, namespace string) (string, error) {
	projectName, err := workloadproject.ProjectNameOfNamespace(ctx, wopv.Client, namespace)
	if err != nil {
		return "", err
	}
	project := &unstructured.Unstructured{}
	project.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   kaiProjectGroup,
		Version: kaiProjectVersion,
		Kind:    kaiProjectKind,
	})
	if err := wopv.Get(ctx, client.ObjectKey{Name: projectName}, project); err != nil {
		return "", err
	}
	parent, _, err := unstructured.NestedString(project.Object, "spec", "parent")
	if err != nil {
		return "", err
	}
	return parent, nil
}

func (wopv *WorkloadOwnershipProtectionValidator) doesUserHaveProjectCreationPermissionsInRelevantScope(
	ctx context.Context,
	departmentName string,
	requesterUID string,
	requesterGroups []string,
	requesterGroupsSet utils.Set[string]) (bool, error) {
	subjectIds := []string{requesterUID}
	subjectIds = append(subjectIds, requesterGroups...)

	for _, subjectId := range subjectIds {
		privilegedWorkloadCrudPermissionsFieldSelector := fields.OneTermEqualSelector(
			pkgcommon.PrivilegedWorkloadCrudPermissionsField, subjectId)
		accessRules, err := utils.ListAccessRules(ctx, wopv.Client, &privilegedWorkloadCrudPermissionsFieldSelector)
		if err != nil {
			// already logged; maybe there was a problem with the field, let's continue..
			continue
		}

		for _, accessRule := range accessRules {
			if doesAccessRuleContainScopeForRequest(accessRule, departmentName, requesterUID, requesterGroupsSet) {
				return true, nil
			}
		}
	}

	return false, nil
}

func (wopv *WorkloadOwnershipProtectionValidator) getCachedWorkloadControllerUID(ctx context.Context) (string, error) {
	if wopv.workloadControllerUID != "" {
		return wopv.workloadControllerUID, nil
	}

	workloadControllerUID, err := webhook.GetWorkloadControllerUID(ctx, wopv.Client)
	if err != nil {
		logger := logf.FromContext(ctx)
		logger.Error(err, "Failed getting workload-controller service account")
		return "", err
	}

	wopv.workloadControllerUID = workloadControllerUID
	return wopv.workloadControllerUID, nil
}

func (wopv *WorkloadOwnershipProtectionValidator) getCachedRuntimeRulesServiceAccountUID(ctx context.Context) (string, error) {
	if wopv.runtimeRulesServiceAccountUID != "" {
		return wopv.runtimeRulesServiceAccountUID, nil
	}

	runtimeRulesServiceAccountUID, err := webhook.GetRuntimeRulesUID(ctx, wopv.Client)
	if err != nil {
		logger := logf.FromContext(ctx)
		logger.Error(err, "Failed getting runtime-rules service account")
		return "", err
	}

	wopv.runtimeRulesServiceAccountUID = runtimeRulesServiceAccountUID
	return wopv.runtimeRulesServiceAccountUID, nil
}

func doesAccessRuleContainScopeForRequest(
	accessRule v1alpha1.AccessRule,
	departmentName string,
	requesterUID string,
	requesterGroupsSet utils.Set[string]) bool {
	// just to be sure
	if !doesAccessRuleEnforceCurrentRequester(&accessRule, requesterUID, requesterGroupsSet) {
		return false
	}

	if accessRule.Spec.ScopeType == v1alpha1.ClusterScope {
		return true
	}
	if accessRule.Spec.ScopeType == v1alpha1.DepartmentScope && accessRule.Spec.ScopeName == departmentName {
		return true
	}
	return false
}

func doesAccessRuleEnforceCurrentRequester(
	accessRule *v1alpha1.AccessRule,
	requesterUID string,
	requesterGroupsSet utils.Set[string]) bool {
	if accessRule.Spec.SubjectId == requesterUID && accessRule.Spec.SubjectType == v1alpha1.UserSubject {
		return true
	}

	if !requesterGroupsSet.Has(accessRule.Spec.SubjectId) {
		return false
	}

	return accessRule.Spec.SubjectType == v1alpha1.GroupSubject
}
