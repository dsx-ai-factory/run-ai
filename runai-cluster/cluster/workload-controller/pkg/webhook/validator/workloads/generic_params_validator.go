/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	workloadbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/run-ai/runai/runai-cluster/common/utils"
	"k8s.io/apimachinery/pkg/api/resource"
	validationAPI "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/utils/ptr"
)

const (
	masterType     = "master "
	limitFiled     = " limit"
	specMemoryName = "CPU memory"
	specGpuMemory  = "GPU memory"
	CPUFieldName   = "CPU compute"

	requestOverLimitErrMsg = "%s limit (%s) must be equal to or higher than %s request (%s)"
)

type GenericParamsValidator struct{}

func NewGenericParamsValidator() *GenericParamsValidator {
	return &GenericParamsValidator{}
}

func (gpv *GenericParamsValidator) Validate(ctx context.Context, workload *jobs.JobWorkload) error {
	baseWorkload := ptr.Deref(workload.Spec.BaseWorkload, "")

	err := ValidateLabels(workload.Labels)
	if err != nil {
		return fmt.Errorf("failed to validate labels, error: %v", err)
	}

	err = ValidatePodAffinity(workload.Spec.RunaiSchedulerFields.PodAffinity, workload.Spec.RunaiSchedulerFields.PodAffinityTopology)
	if err != nil {
		return fmt.Errorf("failed to validate pod affinity, error: %v", err)
	}
	if workload.Spec.Usage == common.UsedAsTemplate && baseWorkload != "" {
		return fmt.Errorf("baseWorkload is only allowed for workloads with usage=Submit")
	}
	err = ValidateImage(workload.Spec.ContainerFields.Image.Deref())
	if err != nil {
		return fmt.Errorf("failed to validate image, error: %v", err)
	}

	err = ValidatePVCs(workload.Spec.StorageFields)
	if err != nil {
		return fmt.Errorf("invalid PVCs: %v", err)
	}

	err = ValidateEmptyDirs(workload.Spec.StorageFields.EmptyDirVolumes)
	if err != nil {
		return fmt.Errorf("invalid EmptyDir volumes: %v", err)
	}

	err = ValidateProbes(workload.Spec.ContainerFields)
	if err != nil {
		return fmt.Errorf("invalid probes: %v", err)
	}

	err = ValidateGitRevision(workload.Spec.GitSync.Deref())
	if err != nil {
		return fmt.Errorf("invalid git parameters, error: %w", err)
	}

	err = ValidateGitAuthenticatingMethod(workload.Spec.GitSync.Deref())
	if err != nil {
		return fmt.Errorf("invalid git parameters, error: %w", err)
	}

	err = ValidateAuthorizedUsersAndGroups(workload.Spec.ExposedUrls, &workload.Spec.InferenceFields.RoutingAndAuthorizationFields)
	if err != nil {
		return fmt.Errorf("invalid authorized users/groups: %v", err)
	}

	_, err = retrieveRestartPolicyValue(workload.Spec.RestartPolicy)
	if err != nil {
		return fmt.Errorf("invalid restart policy value: %v", err)
	}

	err = ValidateWorkloadContainerResourceValuesSpecAndMasterSpec(workload)
	if err != nil {
		return fmt.Errorf("invalid container resource, %v", err)
	}

	return nil
}

func ValidateImage(image string) error {
	if image == "" {
		return fmt.Errorf("image must be provided")
	}
	return nil
}

func ValidateLabels(Labels map[string]string) error {
	// The labels are valid if they are not empty and if they are not longer than 63 characters
	if Labels == nil {
		return nil
	}

	for key, value := range Labels {
		if len(key) == 0 || len(value) == 0 {
			return fmt.Errorf("label %s=%s is invalid, label key and value must not be empty", key, value)
		}
		if len(key) > 63 || len(value) > 63 {
			return fmt.Errorf("label %s=%s is invalid, label key and value must not be longer than 63 characters", key, value)
		}
	}
	return nil
}

func ValidatePodAffinity(podAffinity *flds.BoolField, podAffinityTopology *flds.StringField) error {
	if podAffinity.Deref() {
		if podAffinityTopology.Deref() == "" {
			// not checking podAffinitySchedulingRule value because by default it's always set
			return fmt.Errorf("podAffinityTopology and podAffinitySchedulingRule can be used only when podAffinity is enabled")
		}
	} else {
		if podAffinityTopology.Deref() != "" {
			return fmt.Errorf("podAffinity must be enabled when using podAffinityTopology or podAffinitySchedulingRule")
		}
	}

	return nil
}

// validateCPUValue validates a request/limit pair for CPU with custom error messages and prefix.
func validateCPUValue(requestField, limitField *flds.NumberField, fieldLabel string, fieldPrefix string) error {
	if requestField.DerefStr() == "" && limitField.DerefStr() == "" {
		return nil
	}

	label := fieldPrefix + fieldLabel
	var err error
	var request, limit resource.Quantity
	if requestField.DerefStr() != "" {
		request, err = validatePositiveNumberFiled(requestField, label, request)
		if err != nil {
			return err
		}
	}
	if limitField.DerefStr() != "" {

		limit, err = validatePositiveNumberFiled(limitField, label+limitFiled, limit)
		if err != nil {
			return err
		}
	}

	if limitField.DerefStr() != "" && requestField.DerefStr() != "" &&
		request.Cmp(limit) > 0 {

		return fmt.Errorf(requestOverLimitErrMsg,
			label, limit.String(), label, request.String())
	}
	return nil
}

func validatePositiveNumberFiled(field *flds.NumberField, label string, requestQuantity resource.Quantity) (resource.Quantity, error) {
	q, err := resource.ParseQuantity(field.DerefStr())
	if err != nil {
		return resource.Quantity{}, fmt.Errorf("invalid %s (%s): %s", label, field.DerefStr(), err.Error())
	}
	requestQuantity = q
	if requestQuantity.Sign() < 0 {
		return resource.Quantity{}, fmt.Errorf("%s (%s) must be positive", label, field.DerefStr())
	}
	return requestQuantity, nil
}

// ValidateCPUResourceRequirements will check if any of the CPU compute resource Limits/Requests are of a valid value
func ValidateCPUResourceRequirements(workloadResources field_groups.ResourceAllocationFields, fieldPrefix string) error {
	return validateCPUValue(workloadResources.Cpu, workloadResources.CpuLimit, CPUFieldName, fieldPrefix)
}

func ValidatePVCs(storage field_groups.StorageFields) error {
	pvcs := storage.Pvcs
	if pvcs == nil {
		return nil
	}

	for _, pvc := range pvcs.Deref() {
		pvcValue := pvc.Value
		if pvcValue == nil {
			return fmt.Errorf("value is nil")
		}
		if ptr.Deref(pvcValue.ExistingPvc, false) {
			return nil
		}
		pvcClaimName := ptr.Deref(pvcValue.ClaimName, "")
		err := validatePVCName(pvcClaimName)
		if err != nil {
			return fmt.Errorf("PVC \"%s\" claim name is invalid. %v", pvcClaimName, err.Error())
		}
		if pvc.Value.Size == nil {
			return fmt.Errorf("size is empty")
		}
		pvcSize := ptr.Deref(pvcValue.Size, "")
		withoutPrefixValue, err := resource.ParseQuantity(pvcSize)
		if err != nil {
			return fmt.Errorf("%s size is invalid. %v", pvcSize, err.Error())
		}
		if withoutPrefixValue.Value() <= 0 {
			return fmt.Errorf("%s size is invalid", withoutPrefixValue.String())
		}
	}
	return nil
}

func validatePVCName(pvcName string) error {
	if len(pvcName) == 0 {
		return nil
	}
	// it returns a list of errors, if the list is empty the name is valid
	violation := validationAPI.IsDNS1123Subdomain(pvcName)
	if len(violation) > 0 {
		// return the first violation
		return fmt.Errorf("%s", violation[0])
	}
	return nil
}

func ValidateEmptyDirs(emptyDirVolumes *flds.EmptyDirVolumeField) error {
	_, err := workloadbuilders.GetEmptyDirVolumes(emptyDirVolumes)
	return err
}

func ValidateProbes(containerFields field_groups.ContainerFields) error {
	readinessProbe := containerFields.ReadinessProbe
	if readinessProbe == nil {
		return nil
	}

	readinessProbeSpec := readinessProbe.Deref()

	return workloadbuilders.ValidateProbe(&readinessProbeSpec, "readinessProbe", true)
}

func ValidateAuthorizedUsersAndGroups(exposedUrls *flds.ExposedUrlsField, routingAndAuth *field_groups.RoutingAndAuthorizationFields) error {
	if exposedUrls != nil && exposedUrls.Value != nil {
		for _, exposedUrl := range ptr.Deref(exposedUrls.Value, map[string]field_types.ExposedUrlItemAndRules{}) {
			if exposedUrl.Value.AuthorizedGroups != nil {
				if exposedUrl.Value.AuthorizedUsers != nil {
					return fmt.Errorf("authorizedUsers and authorizedGroups are mutually exclusive")
				}
				_, err := workloadbuilders.SanitizeAuthorizedGroups(*exposedUrl.Value.AuthorizedGroups)
				if err != nil {
					return fmt.Errorf("invalid authorizedGroups format")
				}
			}
		}
	}

	if routingAndAuth != nil {
		if routingAndAuth.ServingPortAuthorizationType.Deref() != constants.ServingPortAuthorizationTypeAuthorizedUsersOrGroups {
			if routingAndAuth.ServingPortAuthorizedUsers.Deref() != "" {
				return fmt.Errorf("servingPortAuthorizedUsers is only allowed when servingPortAuthorizationType is set to authorizedUsersOrGroups")
			}

			if routingAndAuth.ServingPortAuthorizedGroups.Deref() != "" {
				return fmt.Errorf("servingPortAuthorizedGroups is only allowed when servingPortAuthorizationType is set to authorizedUsersOrGroups")
			}
		} else {
			if routingAndAuth.ServingPortAuthorizedUsers.Deref() == "" && routingAndAuth.ServingPortAuthorizedGroups.Deref() == "" {
				return fmt.Errorf("servingPortAuthorizedUsers or servingPortAuthorizedGroups must be set when servingPortAuthorizationType is set to authorizedUsersOrGroups")
			}

			if routingAndAuth.ServingPortAuthorizedUsers.Deref() != "" && routingAndAuth.ServingPortAuthorizedGroups.Deref() != "" {
				return fmt.Errorf("servingPortAuthorizedUsers and servingPortAuthorizedGroups are mutually exclusive")
			}

			if routingAndAuth.ServingPortAuthorizedGroups.Deref() != "" {
				_, err := workloadbuilders.SanitizeAuthorizedGroups(routingAndAuth.ServingPortAuthorizedGroups.Deref())
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func ValidateWorkloadContainerResourceValues(resourceFields *field_groups.ResourceAllocationFields, fieldPrefix string) error {
	// Validate memory (request, limit)
	err := validateMemoryValue(resourceFields.Memory, resourceFields.MemoryLimit, specMemoryName, fieldPrefix)
	if err != nil {
		return err
	}

	// Validate GPU memory (request, limit)
	err = validateMemoryValue(resourceFields.GpuMemory, resourceFields.GpuMemoryLimit, specGpuMemory, fieldPrefix)
	if err != nil {
		return err
	}

	// Validate CPU resources
	err = ValidateCPUResourceRequirements(*resourceFields, fieldPrefix)
	if err != nil {
		return err
	}

	return nil
}

func ValidateWorkloadContainerResourceValuesSpecAndMasterSpec(workload *jobs.JobWorkload) error {
	// Validate spec resource values
	err := ValidateWorkloadContainerResourceValues(&workload.Spec.ResourceAllocationFields, "")
	if err != nil {
		return err
	}

	// Validate masterSpec resource values if it exists
	if workload.MasterSpec != nil {
		err = ValidateWorkloadContainerResourceValues(&workload.MasterSpec.ResourceAllocationFields, masterType)
		if err != nil {
			return err
		}
	}

	return nil
}

// validateMemoryValue validates a request/limit pair for memory or GPU memory with custom error messages.
// If fieldPrefix is not empty, it is prepended to the error messages (e.g., "Master CPU memory ...").
func validateMemoryValue(requestField, limitField *flds.QuantityField, fieldLabel string, fieldPrefix string) error {
	if requestField.Deref() == "" && limitField.Deref() == "" {
		return nil
	}

	var request, limit resource.Quantity
	var err error
	label := fieldPrefix + fieldLabel
	if requestField.Deref() != "" {
		request, err = validatePositiveQuantityField(requestField, label, request)
		if err != nil {
			return err
		}
	}
	if limitField.Deref() != "" {
		limit, err = validatePositiveQuantityField(limitField, label+limitFiled, limit)
		if err != nil {
			return err
		}
	}

	if limitField.Deref() != "" && requestField.Deref() != "" &&
		request.Cmp(limit) > 0 {
		return fmt.Errorf(requestOverLimitErrMsg,
			label, limit.String(), label, request.String())
	}
	return nil
}

func validatePositiveQuantityField(field *flds.QuantityField, label string, request resource.Quantity) (resource.Quantity, error) {
	q, err := resource.ParseQuantity(field.Deref())
	if err != nil {
		return resource.Quantity{}, fmt.Errorf("invalid %s (%s): %s", label, field.Deref(), err.Error())
	}
	request = q
	if request.Sign() < 0 {
		return resource.Quantity{}, fmt.Errorf("invalid %s (%s): must be positive", label, field.Deref())
	}
	return request, nil
}

// ValidateGitRevision will check for ambiguity between revision and branch params
func ValidateGitRevision(workloadGitSyncItems map[string]flds.GitSyncItemAndRules) error {
	if len(workloadGitSyncItems) == 0 {
		return nil
	}

	for _, nick := range utils.SortedKeys(workloadGitSyncItems) {
		if workloadGitSyncItems[nick].Value == nil {
			continue
		}

		params := *workloadGitSyncItems[nick].Value

		branchExists := ptr.Deref(params.Branch, "") != ""
		revisionExists := ptr.Deref(params.Revision, "") != ""

		if revisionExists && branchExists {
			return fmt.Errorf("to point your git revision please provide only one of the following: 'branch' or 'revision'")
		}

	}

	return nil
}

func ValidateGitAuthenticatingMethod(workloadGitSyncItems map[string]flds.GitSyncItemAndRules) error {
	if len(workloadGitSyncItems) == 0 {
		return nil
	}

	for _, nick := range utils.SortedKeys(workloadGitSyncItems) {
		if workloadGitSyncItems[nick].Value == nil {
			continue
		}
		params := *workloadGitSyncItems[nick].Value
		method := ptr.Deref(params.AuthenticatingMethod, "")

		if method != "" && method != constants.GitAuthenticatingMethodPassword && method != constants.GitAuthenticatingMethodSSHKey {
			return fmt.Errorf("authenticating method must be one of the following: '%s' or '%s'", constants.GitAuthenticatingMethodPassword, constants.GitAuthenticatingMethodSSHKey)
		}

		if method == constants.GitAuthenticatingMethodSSHKey && params.SecretName == nil {
			return fmt.Errorf("when authenticating with ssh, a secret to the SSH key must be provided")
		}
	}

	return nil
}
