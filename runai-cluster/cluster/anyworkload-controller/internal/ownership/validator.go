/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package ownership

import (
	"context"
	"net/http"

	runaiv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	authorizationv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	runaiNamespace = "runai"
	queueLabel     = "runai/queue"
)

var privilegedRoles = map[string]struct{}{
	"system-administrator":     {},
	"department-administrator": {},
	"editor":                   {},
}

// Validator protects workload updates and deletions from unauthorized users.
type Validator struct {
	client               client.Client
	serviceAccountReader client.Reader
}

// NewValidator returns a workload ownership validator.
func NewValidator(k8sClient client.Client, serviceAccountReaders ...client.Reader) *Validator {
	serviceAccountReader := client.Reader(k8sClient)
	if len(serviceAccountReaders) != 0 && serviceAccountReaders[0] != nil {
		serviceAccountReader = serviceAccountReaders[0]
	}
	return &Validator{client: k8sClient, serviceAccountReader: serviceAccountReader}
}

func (v *Validator) Handle(ctx context.Context, request admission.Request) admission.Response {
	if request.Kind.Kind == "WorkloadRunner" {
		return v.handleWorkloadRunner(ctx, request)
	}
	workload, err := decodeLegacyWorkload(request.OldObject.Raw)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if workload.usage() != submitUsage || workload.ownerUID() == "" {
		return admission.Allowed("")
	}
	if request.Operation == admissionv1.Update {
		desired, err := decodeLegacyWorkload(request.Object.Raw)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		if desired.ownerUID() != workload.ownerUID() {
			return admission.Denied("workload ownership cannot be changed")
		}
	}

	allowed, err := v.isAllowed(ctx, workload.namespace(), workload.ownerUID(), request.UserInfo)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	if allowed {
		return admission.Allowed("")
	}
	return admission.Denied("operation not allowed")
}

func (v *Validator) handleWorkloadRunner(ctx context.Context, request admission.Request) admission.Response {
	var storedOwner string
	var namespace = request.Namespace
	var stored *workloadRunner
	if request.Operation != admissionv1.Create {
		var err error
		stored, err = decodeWorkloadRunner(request.OldObject.Raw)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		storedOwner = stored.ownerUID()
		namespace = stored.namespace()
	}
	if request.Operation == admissionv1.Update && storedOwner != "" {
		desired, err := decodeWorkloadRunner(request.Object.Raw)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		if desired.ownerUID() != storedOwner {
			return admission.Denied("workload ownership cannot be changed")
		}
	}

	if storedOwner != "" {
		allowed, err := v.isAllowed(ctx, namespace, storedOwner, request.UserInfo)
		if err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}
		if allowed {
			return admission.Allowed("")
		}
		return admission.Denied("operation not allowed")
	}
	if stored != nil && stored.controlPlaneManaged() {
		agent, err := v.isServiceAccount(ctx, "runai-agent", requesterUserID(request.UserInfo))
		if err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}
		if agent {
			return admission.Allowed("")
		}
		return admission.Denied("control-plane-managed workload ownership has not been initialized")
	}

	if request.Operation == admissionv1.Delete {
		return admission.Allowed("")
	}
	desired, err := decodeWorkloadRunner(request.Object.Raw)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if desired.ownerUID() == "" {
		return admission.Allowed("")
	}
	agent, err := v.isServiceAccount(ctx, "runai-agent", requesterUserID(request.UserInfo))
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	if agent {
		return admission.Allowed("")
	}
	return admission.Denied("operation not allowed")
}

func (v *Validator) isAllowed(ctx context.Context, namespace, ownerUID string, userInfo authenticationv1.UserInfo) (bool, error) {
	requesterUID := requesterUserID(userInfo)
	if requesterUID == ownerUID {
		return true, nil
	}

	internal, err := v.isInternalServiceAccount(ctx, requesterUID)
	if err != nil || internal {
		return internal, err
	}

	department, err := v.departmentOfNamespace(ctx, namespace)
	if err != nil {
		return false, err
	}
	privileged, err := v.hasPrivilegedAccessRule(ctx, department, requesterUID, userInfo.Groups)
	if err != nil || privileged {
		return privileged, err
	}

	agent, err := v.isServiceAccount(ctx, "runai-agent", requesterUID)
	if err != nil || agent {
		return agent, err
	}
	return v.canCreateNamespaces(ctx, userInfo)
}

func requesterUserID(userInfo authenticationv1.UserInfo) string {
	if userInfo.UID != "" {
		return userInfo.UID
	}
	return userInfo.Username
}

func (v *Validator) isInternalServiceAccount(ctx context.Context, requesterUID string) (bool, error) {
	for _, name := range []string{"workload-controller", "inference-workload-controller", "runtime-rules"} {
		matches, err := v.isServiceAccount(ctx, name, requesterUID)
		if err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return false, err
		}
		if matches {
			return true, nil
		}
	}
	return false, nil
}

func (v *Validator) isServiceAccount(ctx context.Context, name, requesterUID string) (bool, error) {
	serviceAccount := &corev1.ServiceAccount{}
	if err := v.serviceAccountReader.Get(ctx, types.NamespacedName{Name: name, Namespace: runaiNamespace}, serviceAccount); err != nil {
		return false, err
	}
	return string(serviceAccount.UID) == requesterUID, nil
}

func (v *Validator) departmentOfNamespace(ctx context.Context, namespace string) (string, error) {
	namespaceObject := &corev1.Namespace{}
	if err := v.client.Get(ctx, types.NamespacedName{Name: namespace}, namespaceObject); err != nil {
		return "", err
	}
	projectName := namespaceObject.Labels[queueLabel]
	if projectName == "" {
		return "", nil
	}
	project := &unstructured.Unstructured{}
	project.SetGroupVersionKind(schema.GroupVersionKind{Group: "kai.resources", Version: "v1alpha1", Kind: "Project"})
	if err := v.client.Get(ctx, types.NamespacedName{Name: projectName}, project); err != nil {
		return "", err
	}
	department, _, err := unstructured.NestedString(project.Object, "spec", "parent")
	return department, err
}

func (v *Validator) hasPrivilegedAccessRule(ctx context.Context, department, requesterUID string, groups []string) (bool, error) {
	accessRules := &runaiv1alpha1.AccessRuleList{}
	if err := v.client.List(ctx, accessRules); err != nil {
		return false, err
	}
	requesterGroups := make(map[string]struct{}, len(groups))
	for _, group := range groups {
		requesterGroups[group] = struct{}{}
	}
	for index := range accessRules.Items {
		accessRule := &accessRules.Items[index]
		if _, privileged := privilegedRoles[accessRule.Spec.RoleName]; !privileged {
			continue
		}
		if !accessRuleAppliesToRequester(accessRule, requesterUID, requesterGroups) {
			continue
		}
		if accessRule.Spec.ScopeType == runaiv1alpha1.ClusterScope ||
			(accessRule.Spec.ScopeType == runaiv1alpha1.DepartmentScope && accessRule.Spec.ScopeName == department) {
			return true, nil
		}
	}
	return false, nil
}

func accessRuleAppliesToRequester(accessRule *runaiv1alpha1.AccessRule, requesterUID string, groups map[string]struct{}) bool {
	if accessRule.Spec.SubjectType == runaiv1alpha1.UserSubject {
		return accessRule.Spec.SubjectId == requesterUID
	}
	if accessRule.Spec.SubjectType != runaiv1alpha1.GroupSubject {
		return false
	}
	_, exists := groups[accessRule.Spec.SubjectId]
	return exists
}

func (v *Validator) canCreateNamespaces(ctx context.Context, userInfo authenticationv1.UserInfo) (bool, error) {
	review := &authorizationv1.SubjectAccessReview{
		Spec: authorizationv1.SubjectAccessReviewSpec{
			User:   userInfo.Username,
			UID:    userInfo.UID,
			Groups: userInfo.Groups,
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Verb: "create", Resource: "namespaces",
			},
		},
	}
	if err := v.client.Create(ctx, review, &client.CreateOptions{DryRun: []string{"All"}}); err != nil {
		return false, err
	}
	return review.Status.Allowed, nil
}
