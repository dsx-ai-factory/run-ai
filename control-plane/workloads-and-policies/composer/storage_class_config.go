/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	"fmt"
	"net/http"
	"regexp"

	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	k8sObjectsService "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client"
	k8sObjectsApi "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	"k8s.io/apimachinery/pkg/api/resource"
)

var onlyLetters = regexp.MustCompile(`[a-zA-Z]+$`)

func ApplyStorageClassConfig(scName string, claimInfo **workloadsApi.ClaimInfo, config *k8sObjectsApi.StorageClassConfiguration) *workloadsApi.Error {
	if config == nil || !convutils.HasValue(config.Customization) {
		return makeSureNoAddedAttrs(*claimInfo)
	}

	customization := config.Customization.Get()

	scNameForMessage := fmt.Sprintf("storage class %s", scName)
	if scName == NoneStorageClass {
		scNameForMessage = "static provisioning"
	}

	if *claimInfo != nil {
		err := applyAccessModeConfig(scNameForMessage, *claimInfo, customization.AccessMode.Get())
		if err != nil {
			return err
		}

		err = applyVolumeModeConfig(scNameForMessage, *claimInfo, customization.VolumeMode.Get())
		if err != nil {
			return err
		}

		err = applyClaimSizeConfig(scNameForMessage, *claimInfo, customization.ClaimSize.Get())
		if err != nil {
			return err
		}
	}

	if len(customization.Attributes) == 0 {
		return makeSureNoAddedAttrs(*claimInfo)
	}

	//
	//   load all the configured attributes to a map
	//
	configuredAttrs := make(map[string]*k8sObjectsApi.PvcAddedAttribute)

	for _, configuredAttr := range customization.Attributes {
		configuredAttrs[configuredAttr.Key] = &configuredAttr
	}

	//
	//  we have configured attributes, some of them may have defaults that we need to apply, or
	//  validation that
	//
	givenAttrKeys := datastructs.NewSet[string]()

	if *claimInfo == nil {
		*claimInfo = &workloadsApi.ClaimInfo{
			AddedAttrValues: make([]workloadsApi.PvcAddedAttrValue, 0),
		}
	} else if (*claimInfo).AddedAttrValues == nil {
		(*claimInfo).AddedAttrValues = make([]workloadsApi.PvcAddedAttrValue, 0)
	} else {
		for idx, attr := range (*claimInfo).AddedAttrValues {
			keyPrefix := fmt.Sprintf("attribute '%s': ", attr.Key)

			config, attrFound := configuredAttrs[attr.Key]
			if !attrFound {
				return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("attribute '%s' is not supported for %s", attr.Key, scNameForMessage))
			}

			if givenAttrKeys.Has(attr.Key) {
				return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("attribute '%s' is multiply defined", attr.Key))
			}

			givenAttrKeys.Add(attr.Key)

			if attr.Value == nil {
				if !convutils.HasValue(config.DefaultValue) {
					return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("value for attribute '%s' is not provided", attr.Key))
				}
				(*claimInfo).AddedAttrValues[idx].Value = config.DefaultValue.Get()
			} else {
				valueErr := k8sObjectsService.ValidatePvcAttValueVsConfig(config, keyPrefix, *attr.Value)
				if valueErr != nil {
					return workloadsApi.NewError(valueErr.Code, valueErr.Message)
				}
			}
		}
	}

	//
	//   handle required attributes that were not provided
	//
	for attrKey, attrConfig := range configuredAttrs {
		if givenAttrKeys.Has(attrKey) {
			continue
		}

		if convutils.BoolIsTrue(attrConfig.Required.Get()) {
			if !convutils.HasValue(attrConfig.DefaultValue) {
				return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("mandatory attribute '%s' is not provided", attrKey))
			}

			(*claimInfo).AddedAttrValues = append((*claimInfo).AddedAttrValues, workloadsApi.PvcAddedAttrValue{
				Key:   attrKey,
				Value: attrConfig.DefaultValue.Get(),
			})
		}
	}

	return nil
}

func applyAccessModeConfig(scNameForMessage string, claimInfo *workloadsApi.ClaimInfo, config *k8sObjectsApi.AccessModeCustomization) *workloadsApi.Error {
	if config == nil {
		return nil
	}

	//
	//  apply default if needed
	//
	if !convutils.HasValue(claimInfo.AccessModes) && convutils.HasValue(config.Default) {
		claimInfo.AccessModes = *workloadsApi.NewNullablePvcAccessModes(&workloadsApi.PvcAccessModes{})

		accessModes := claimInfo.AccessModes.Get()
		defaultModes := config.Default.Get()

		if convutils.HasValue(config.Default.Get().ReadWriteOnce) {
			accessModes.ReadWriteOnce = *workloadsApi.NewNullableBool(defaultModes.ReadWriteOnce.Get())
		}

		if convutils.HasValue(config.Default.Get().ReadWriteMany) {
			accessModes.ReadWriteMany = *workloadsApi.NewNullableBool(defaultModes.ReadWriteMany.Get())
		}

		if convutils.HasValue(config.Default.Get().ReadOnlyMany) {
			accessModes.ReadOnlyMany = *workloadsApi.NewNullableBool(defaultModes.ReadOnlyMany.Get())
		}
	}

	//
	//   check required flag
	//
	if !convutils.HasValue(claimInfo.AccessModes) {
		if convutils.BoolIsTrue(config.Required.Get()) {
			return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("mandatory accessModes for %s is not provided", scNameForMessage))
		}
		return nil
	}

	//
	//   check supported values
	//
	if !convutils.HasValue(config.SupportedValues) || !convutils.HasValue(claimInfo.AccessModes) {
		return nil
	}

	accessModes := claimInfo.AccessModes.Get()
	supported := config.SupportedValues.Get()

	if convutils.BoolIsTrue(accessModes.ReadOnlyMany.Get()) && !convutils.Deref(supported.ReadOnlyMany.Get(), true) {
		return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("ReadOnlyMany is not supported for %s", scNameForMessage))
	}

	if convutils.BoolIsTrue(accessModes.ReadWriteMany.Get()) && !convutils.Deref(supported.ReadWriteMany.Get(), true) {
		return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("ReadWriteMany is not supported for %s", scNameForMessage))
	}

	if convutils.BoolIsTrue(accessModes.ReadWriteOnce.Get()) && !convutils.Deref(supported.ReadWriteOnce.Get(), true) {
		return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("readWriteOnce is not supported for %s", scNameForMessage))
	}

	return nil
}

func applyVolumeModeConfig(scNameForMessage string, claimInfo *workloadsApi.ClaimInfo, config *k8sObjectsApi.VolumeModeCustomization) *workloadsApi.Error {
	if config == nil {
		return nil
	}

	//
	//  apply default if needed
	//
	if !convutils.HasValue(claimInfo.VolumeMode) && convutils.HasValue(config.Default) {
		defaultMode := workloadsApi.PvcVolumeMode(convutils.Deref(config.Default.Get(), ""))
		claimInfo.VolumeMode = *workloadsApi.NewNullablePvcVolumeMode(&defaultMode)
	}

	//
	//   check required flag
	//
	if !convutils.HasValue(claimInfo.VolumeMode) {
		if convutils.BoolIsTrue(config.Required.Get()) {
			return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("mandatory volumeMode for %s is not provided", scNameForMessage))
		}
		return nil
	}

	//
	//   check supported values
	//
	if !convutils.HasValue(config.SupportedValues) || !convutils.HasValue(claimInfo.VolumeMode) {
		return nil
	}

	volumeMode := claimInfo.VolumeMode.Get()
	supported := config.SupportedValues.Get()

	if *volumeMode == workloadsApi.PVCVOLUMEMODE_FILESYSTEM && !convutils.Deref(supported.Filesystem.Get(), true) {
		return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("%s is not supported for %s", workloadsApi.PVCVOLUMEMODE_FILESYSTEM, scNameForMessage))
	}

	if *volumeMode == workloadsApi.PVCVOLUMEMODE_BLOCK && !convutils.Deref(supported.Block.Get(), true) {
		return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("%s is not supported for %s", workloadsApi.PVCVOLUMEMODE_BLOCK, scNameForMessage))
	}

	return nil
}

func applyClaimSizeConfig(_ string, claimInfo *workloadsApi.ClaimInfo, config *k8sObjectsApi.ClaimSizeCustomization) *workloadsApi.Error {
	if config == nil {
		return nil
	}

	//
	//  apply default if needed
	//
	if !convutils.HasValue(claimInfo.Size) && convutils.HasValue(config.Default) {
		defaultSize := convutils.Deref(config.Default.Get(), "")
		claimInfo.Size = *workloadsApi.NewNullableString(&defaultSize)
	}

	//
	//   no required flag for size. it is always mandatory and this is validated in the caller
	//
	if !convutils.HasValue(claimInfo.Size) {
		return nil
	}

	var supportedUnits datastructs.Set[string]

	if len(config.SupportedUnits) != 0 {
		supportedUnits = datastructs.NewSet(config.SupportedUnits...)
	}

	var sizeErr *workloadsApi.Error

	value := matchPvcSizeToUnit(&sizeErr, "pvcSize", claimInfo.Size.Get(), supportedUnits)
	minSize := matchPvcSizeToUnit(&sizeErr, "min", config.Min.Get(), supportedUnits)
	maxSize := matchPvcSizeToUnit(&sizeErr, "max", config.Max.Get(), supportedUnits)
	stepSize := matchPvcSizeToUnit(&sizeErr, "step", config.Step.Get(), supportedUnits)

	if sizeErr != nil {
		return sizeErr
	}

	if value == 0 {
		return workloadsApi.NewError(http.StatusBadRequest, "claim size should be a positive value")
	}

	if minSize > 0 && value < minSize {
		return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("claim size should be no less than %s", *config.Min.Get()))
	}

	if maxSize > 0 && value > maxSize {
		return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("claim size should be no greater than %s", *config.Max.Get()))
	}

	if stepSize > 0 {
		//
		//  if min value is not specified take 0 for the calculation, so step 3 will be valid for values 3, 6, 9, ...
		//
		var realMinSize = minSize
		if minSize < 0 {
			realMinSize = 0
		}

		diffFromMin := int64(value - realMinSize)
		if diffFromMin%int64(stepSize) != 0 {
			if minSize > 0 {
				return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("claim size is not aligned with %s steps from the minimal value %s", *config.Step.Get(), *config.Min.Get()))
			} else {
				return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("claim size is not aligned with steps of %s", *config.Step.Get()))
			}
		}
	}

	return nil
}

// this function is more comfortable to use when there is a chain of calls to match multiple values
func matchPvcSizeToUnit(err **workloadsApi.Error, description string, value *string, supportedValues datastructs.Set[string]) float64 {
	if err == nil || *err != nil {
		return -1
	}

	realSize, matchErr := MatchPvcSizeToUnit(description, value, supportedValues)

	if matchErr != nil {
		*err = matchErr
	}

	return realSize
}

func MatchPvcSizeToUnit(description string, value *string, supportedValues datastructs.Set[string]) (float64, *workloadsApi.Error) {
	if value == nil {
		return -1, nil
	}

	//
	//  note that the units are MB, MiB, etc. but the value is given without the B
	//
	if supportedValues != nil {
		unit := onlyLetters.FindString(*value)
		if !supportedValues.Has(unit + "B") {
			return -1, workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("%s: unit suffix '%s' is not supported", description, unit))
		}
	}

	realSize, convErr := resource.ParseQuantity(*value)
	if convErr != nil {
		return -1, workloadsApi.NewError(http.StatusBadRequest, convErr.Error())
	}

	return realSize.AsApproximateFloat64(), nil
}

func makeSureNoAddedAttrs(claimInfo *workloadsApi.ClaimInfo) *workloadsApi.Error {
	if claimInfo == nil {
		return nil
	}

	if len(claimInfo.AddedAttrValues) != 0 {
		return workloadsApi.NewError(http.StatusBadRequest, "addedAttrValues are not permitted for PVC of this storage class")
	}

	return nil
}
