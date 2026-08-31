/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-openapi/jsonpointer"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/common/utils/validator"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/webhook/validator/workloads"
	"gomodules.xyz/jsonpatch/v2"
	admissionV1 "k8s.io/api/admission/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/pkg/logging"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	namePath             = "/metadata/name"
	namespacePath        = "/metadata/namespace"
	metadataPrefix       = "/metadata"
	statusPrefix         = "/status"
	activeFieldPrefix    = "/spec/active"
	managedByPath        = "/spec/managedBy"
	kaiAnnotationPrefix  = "/metadata/annotations/kai.scheduler"
	metadataLabelsPrefix = "/metadata/labels/"
	validOk              = "Ok"

	managedByFieldDeniedMessage = "The managedBy field is reserved for system controllers and cannot be modified manually."
)

var (
	immutableMetadataLabels = []string{mnnvl.FeatureLabelKey}
	// escapedRestartRequestUIDLabel is the JSON pointer escaped form of RestartRequestUIDLabel for patch path checks
	restartRequestUIDPath = metadataLabelsPrefix + jsonpointer.Escape(workloads.RestartRequestUIDLabel)
)

type WorkloadValidatorHandler struct {
	k8sClient                      client.Client
	creationValidators             []workloads.WorkloadValidator
	updateValidators               []workloads.WorkloadValidator
	escapedImmutableMetadataLabels []string
	allowedServiceAccounts         []string
}

func NewWorkloadParamsValidator(k8sClient client.Client, allowedServiceAccounts []string, creationValidators ...workloads.WorkloadValidator) *WorkloadValidatorHandler {
	var workloadValidators []workloads.WorkloadValidator
	workloadValidators = append(workloadValidators, creationValidators...)

	// in json patches, labels cannot have '/' chars, so they are escaped
	// https://www.rfc-editor.org/rfc/rfc6901#section-3
	var escapedImmutableMetadataLabels []string
	for _, label := range immutableMetadataLabels {
		escapedImmutableMetadataLabels = append(escapedImmutableMetadataLabels, jsonpointer.Escape(label))
	}

	return &WorkloadValidatorHandler{
		k8sClient:                      k8sClient,
		creationValidators:             workloadValidators,
		updateValidators:               []workloads.WorkloadValidator{},
		escapedImmutableMetadataLabels: escapedImmutableMetadataLabels,
		allowedServiceAccounts:         allowedServiceAccounts,
	}
}

func (wv *WorkloadValidatorHandler) RegisterUpdateValidators(validators ...workloads.WorkloadValidator) {
	wv.updateValidators = validators
}

func (wv *WorkloadValidatorHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	switch req.Operation {
	case admissionV1.Create:
		return wv.HandleCreate(ctx, req.Object, req.RequestKind.Kind, req.UserInfo.Username)
	case admissionV1.Update:
		return wv.HandleUpdate(ctx, req.OldObject, req.Object, req.RequestKind.Kind, req.UserInfo.Username)
	default:
		// in the case of delete, we don't need to do anything
		return admission.Allowed(validOk)
	}
}

func (wv *WorkloadValidatorHandler) HandleCreate(ctx context.Context, newObject runtime.RawExtension, workloadKind string, username string) admission.Response {
	log := logging.FromContext(ctx)

	workload := jobs.JobWorkload{}
	err := json.Unmarshal(newObject.Raw, &workload)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	err = wv.enforceOwnership(workloadKind, username, workload.Labels, wv.allowedServiceAccounts)
	if err != nil {
		return admission.Errored(http.StatusUnauthorized, err)
	}

	if workload.Spec.ManagedBy != "" && !validator.AllowedServiceAccounts(username, wv.allowedServiceAccounts) {
		return admission.Denied(managedByFieldDeniedMessage)
	}

	// check if there is service with resource name with the same name, if so, deny
	err = wv.workloadResourceLookupFound(ctx, &workload)
	if err != nil {
		log.Error(err, "admission failed due resource lookup result")
		return admission.Denied(err.Error())
	}

	for _, creationValidator := range wv.creationValidators {
		err = creationValidator.Validate(ctx, &workload)
		if err != nil {
			return admission.Denied(err.Error())
		}
	}

	return admission.Allowed("Ok")
}

func (wv *WorkloadValidatorHandler) InjectDecoder(_ *admission.Decoder) error {
	return nil
}

func (wv *WorkloadValidatorHandler) HandleUpdate(ctx context.Context, oldObject runtime.RawExtension, newObject runtime.RawExtension,
	workloadKind string, username string) admission.Response {

	log := logf.FromContext(ctx)

	//
	//   first, let's do the validations on the new workload. this may be sufficient so we won't need
	//   to unmarshal the old object and save time
	//
	newWorkload := jobs.JobWorkload{}

	err := json.Unmarshal(newObject.Raw, &newWorkload)
	if err != nil {
		log.Error(err, "cannot unmarshal new object")
		return admission.Errored(http.StatusBadRequest, err)
	}

	// if the new workload is marked with label of run.ai/workload we allow it only
	// to system service account
	err = wv.enforceOwnership(workloadKind, username, newWorkload.Labels, wv.allowedServiceAccounts)
	if err != nil {
		return admission.Errored(http.StatusUnauthorized, err)
	}

	isTemplate := isUsageTemplate(&newWorkload)
	isAllowedServiceAccounts := validator.AllowedServiceAccounts(username, wv.allowedServiceAccounts)

	// check if we need the old workload. we need it in the following cases
	// - if the user is not system service account, we need it to validate ownership based on the labels
	//   of the old workload
	// - if the new workload is a template, we need to check if the old object is also a template. if both
	//   are templates, no need to perform some of the validations that we do if old/new workload is not a template
	// - if DistributedWorkload, we need it to validate managedBy and restart-request-uid label rules
	needToCheckOldWorkload := !isAllowedServiceAccounts || isTemplate || workloadKind == constants.KindDistributedWorkload

	var oldWorkload jobs.JobWorkload
	if needToCheckOldWorkload {
		// arriving here, the information from the new workload is not sufficient for us
		// and we have to check also the old object
		err = json.Unmarshal(oldObject.Raw, &oldWorkload)
		if err != nil {
			log.Error(err, "cannot unmarshal old object")
			return admission.Errored(http.StatusBadRequest, err)
		}

		// and also the ownership on the labels of the old workload
		err = wv.enforceOwnership(workloadKind, username, oldWorkload.Labels, wv.allowedServiceAccounts)
		if err != nil {
			return admission.Errored(http.StatusUnauthorized, err)
		}

		isTemplate = isTemplate && isUsageTemplate(&oldWorkload)

		if !isAllowedServiceAccounts && oldWorkload.Spec.ManagedBy != newWorkload.Spec.ManagedBy {
			return admission.Denied(managedByFieldDeniedMessage)
		}
	}

	if isTemplate {
		// support workload template is going to be deprecated soon
		// for now, we allow all changes to workload template
		return admission.Allowed(validOk)
	}

	// DistributedWorkload: validate managedBy and restart-request-uid label rules
	if workloadKind == constants.KindDistributedWorkload {
		if err = workloads.ValidateDistributedWorkloadRestartManagedBy(&oldWorkload, &newWorkload); err != nil {
			return admission.Denied(err.Error())
		}
	}

	// When a controller has taken over management (e.g. for restart), apply create-level validation
	// instead of strict immutability, allowing spec fields to be updated.
	if newWorkload.Spec.ManagedBy != "" {
		if !isAllowedServiceAccounts {
			return admission.Denied(managedByFieldDeniedMessage)
		}
		for _, v := range wv.creationValidators {
			if err = v.Validate(ctx, &newWorkload); err != nil {
				return admission.Denied(err.Error())
			}
		}
		return admission.Allowed(validOk)
	}

	patches, err := jsonpatch.CreatePatch(oldObject.Raw, newObject.Raw)
	if err != nil {
		log.Error(err, "cannot compare old and new objects")
		return admission.Errored(http.StatusServiceUnavailable, err)
	}
	for _, change := range patches {
		if !wv.onlyActiveOrMetadataChange(change, workloadKind) {
			return admission.Denied(fmt.Sprintf("cannot update %s, field is immutable", change.Path))
		}
	}

	for _, updateValidator := range wv.updateValidators {
		err = updateValidator.Validate(ctx, &newWorkload)
		if err != nil {
			return admission.Denied(err.Error())
		}
	}

	return admission.Allowed(validOk)
}

func (wv *WorkloadValidatorHandler) workloadResourceLookupFound(ctx context.Context, workload *jobs.JobWorkload) error {
	labels := map[string]string{constants.LabelRelease: workload.Name}
	serviceList := &v1.ServiceList{}

	err := wv.k8sClient.List(ctx, serviceList,
		client.InNamespace(workload.Namespace), client.MatchingLabels(labels))
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if len(serviceList.Items) != 0 {
		return fmt.Errorf("A workload with the name %s already exists. Please choose a different name for your workload.", workload.Name)
	}

	return nil
}

func (wv *WorkloadValidatorHandler) onlyActiveOrMetadataChange(change jsonpatch.Operation, workloadKind string) bool {
	// DistributedWorkload: allow spec.managedBy and restart-request-uid label (validation already enforced above)
	if workloadKind == constants.KindDistributedWorkload {
		if change.Path == managedByPath || change.Path == restartRequestUIDPath {
			return true
		}
	}

	// Prevent changes of topology annotations
	if strings.HasPrefix(change.Path, kaiAnnotationPrefix) {
		return false
	}

	// do not allow changes to name or namespace

	if strings.HasPrefix(change.Path, namePath) || strings.HasPrefix(change.Path, namespacePath) {
		return false
	}
	// allow changes to all other metadata fields

	if strings.HasPrefix(change.Path, metadataPrefix) {
		if wv.isImmutableMetadataLabelChange(change) {
			return false
		}

		return true
	}
	// allow changes to status

	if strings.HasPrefix(change.Path, statusPrefix) {
		return true
	}
	// if the change is not in spec/active, deny

	return strings.HasPrefix(change.Path, activeFieldPrefix)
}

func (wv *WorkloadValidatorHandler) isImmutableMetadataLabelChange(change jsonpatch.Operation) bool {
	for _, label := range wv.escapedImmutableMetadataLabels {
		if strings.HasSuffix(change.Path, label) {
			return true
		}
	}

	return false
}

func (wv *WorkloadValidatorHandler) enforceOwnership(workloadKind string, requesterUsername string, labels map[string]string, allowedServiceAccounts []string) error {
	//
	//   we apply service account ownership protection only for run-ai workloads
	//
	_, runaiWorkload := labels[constants.WorkloadLabel]
	if !runaiWorkload {
		return nil
	}

	switch workloadKind {
	case constants.KindTrainingWorkload, constants.KindInteractiveWorkload, constants.KindDistributedWorkload:
		if !validator.AllowedServiceAccounts(requesterUsername, allowedServiceAccounts) {
			return fmt.Errorf("Username %s is not authorized to perform this action on %s", requesterUsername, workloadKind)
		}
	}

	return nil
}

func isUsageTemplate(workload *jobs.JobWorkload) bool {
	return workload != nil && workload.Spec.Usage == common.UsedAsTemplate
}
