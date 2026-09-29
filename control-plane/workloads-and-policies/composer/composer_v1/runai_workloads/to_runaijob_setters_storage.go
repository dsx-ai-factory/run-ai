/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runai_workloads

import (
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
)

func mapToVolumesField(src map[string]*workloadsApi.HostPathInstance, target any, _ any) *workloadsApi.Error {
	if len(src) == 0 {
		return nil
	}

	targetPtr, ok := target.(**field_types.VolumesField)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("mapToVolumesField: target type %T is not **field_types.VolumesField", target))
	}
	if *targetPtr == nil {
		*targetPtr = &field_types.VolumesField{}
	}
	if (*targetPtr).Value == nil {
		(*targetPtr).Value = &map[string]field_types.VolumeItemAndRules{}
	}

	targetMap := *((*targetPtr).Value)

	for _, key := range composer.SortedMapKeys(src) {
		instance := src[key]

		if instance == nil {
			continue
		}

		item := field_types.VolumeItemValue{
			SourcePath: instance.Path.Get(),
			TargetPath: convutils.EmptyIfNull(instance.MountPath.Get()),
		}

		if convutils.HasValue(instance.ReadOnly) {
			item.ReadOnly = convutils.Ref(convutils.Deref(instance.ReadOnly.Get(), true))
		}

		targetMap[key] = field_types.VolumeItemAndRules{
			Value: &item,
		}
	}

	return nil
}

func mapToNfsField(src map[string]*workloadsApi.NfsInstance, target any, _ any) *workloadsApi.Error {
	if len(src) == 0 {
		return nil
	}

	targetPtr, ok := target.(**field_types.NfsField)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("mapToNfsField: target type %T is not **field_types.NfsField", target))
	}
	if *targetPtr == nil {
		*targetPtr = &field_types.NfsField{}
	}
	if (*targetPtr).Value == nil {
		(*targetPtr).Value = &map[string]field_types.NfsItemAndRules{}
	}

	targetMap := *((*targetPtr).Value)

	for _, key := range composer.SortedMapKeys(src) {
		instance := src[key]

		if instance == nil {
			continue
		}

		item := field_types.NfsItemValue{
			SourcePath: convutils.EmptyIfNull(instance.Path.Get()),
			TargetPath: convutils.EmptyIfNull(instance.MountPath.Get()),
			NfsServer:  convutils.EmptyIfNull(instance.Server.Get()),
		}

		if convutils.HasValue(instance.ReadOnly) {
			item.ReadOnly = convutils.Ref(convutils.Deref(instance.ReadOnly.Get(), true))
		}

		targetMap[key] = field_types.NfsItemAndRules{
			Value: &item,
		}
	}

	return nil
}

func mapToGitSyncField(info *composer.GitSyncInfo, target any, _ any) *workloadsApi.Error {
	if info == nil || len(info.Items) == 0 {
		return nil
	}

	targetPtr, ok := target.(**field_types.GitSyncField)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("mapToGitSyncField: target type %T is not **field_types.GitSyncField", target))
	}
	if *targetPtr == nil {
		*targetPtr = &field_types.GitSyncField{}
	}
	if (*targetPtr).Value == nil {
		(*targetPtr).Value = &map[string]field_types.GitSyncItemAndRules{}
	}

	targetMap := *((*targetPtr).Value)

	for _, key := range composer.SortedMapKeys(info.Items) {
		item := info.Items[key]

		if item == nil {
			continue
		}

		value := field_types.GitSyncItemValue{
			Repository: convutils.Ref(item.Repository),
			Directory:  convutils.Ref(item.Directory),
		}

		if item.Branch != "" {
			value.Branch = convutils.Ref(item.Branch)
		}

		if item.Revision != "" {
			value.Revision = convutils.Ref(item.Revision)
		}

		if item.SecretName != "" {
			value.SecretName = convutils.Ref(item.SecretName)
		}

		if item.AuthenticatingMethod != "" {
			value.AuthenticatingMethod = convutils.Ref(item.AuthenticatingMethod)
		}

		if item.Username != "" {
			value.Username = convutils.Ref(item.Username)
		}

		if item.Password != "" {
			value.Password = convutils.Ref(item.Password)
		}

		targetMap[key] = field_types.GitSyncItemAndRules{
			Value: &value,
		}
	}

	return nil
}

func mapToPvcField(src map[string]*workloadsApi.PvcInstance, target any, _ any) *workloadsApi.Error {
	if len(src) == 0 {
		return nil
	}

	targetPtr, ok := target.(**field_types.PvcsField)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("mapToPvcField: target type %T is not **field_types.PvcsField", target))
	}
	if *targetPtr == nil {
		*targetPtr = &field_types.PvcsField{}
	}
	if (*targetPtr).Value == nil {
		(*targetPtr).Value = &map[string]field_types.PvcsItemAndRules{}
	}

	targetMap := *((*targetPtr).Value)

	for _, key := range composer.SortedMapKeys(src) {
		instance := src[key]

		if instance == nil {
			continue
		}

		item := field_types.PvcsItemValue{
			Path:      convutils.EmptyIfNull(instance.Path.Get()),
			Ephemeral: instance.Ephemeral.Get(),
		}

		if convutils.HasValue(instance.ClaimName) {
			item.ClaimName = instance.ClaimName.Get()
		}

		if convutils.HasValue(instance.ReadOnly) {
			item.ReadOnly = convutils.Ref(convutils.Deref(instance.ReadOnly.Get(), true))
		}

		item.ExistingPvc = instance.ExistingPvc.Get()

		if !convutils.BoolIsTrue(item.ExistingPvc) && convutils.HasValue(instance.ClaimInfo) {
			claimInfo := instance.ClaimInfo.Get()

			item.StorageClass = claimInfo.StorageClass.Get()

			item.Size = claimInfo.Size.Get()

			if convutils.HasValue(claimInfo.AccessModes) {
				item.ReadOnlyMany = claimInfo.AccessModes.Get().ReadOnlyMany.Get()
				item.ReadWriteOnce = claimInfo.AccessModes.Get().ReadWriteOnce.Get()
				item.ReadWriteMany = claimInfo.AccessModes.Get().ReadWriteMany.Get()
			}

			if convutils.HasValue(claimInfo.VolumeMode) {
				item.VolumeMode = convutils.Ref(string(*claimInfo.VolumeMode.Get()))
			}
		}

		targetMap[key] = field_types.PvcsItemAndRules{
			Value: &item,
		}
	}

	return nil
}

func mapToS3Field(info *composer.S3Info, target any, _ any) *workloadsApi.Error {
	if info == nil || len(info.Items) == 0 {
		return nil
	}

	targetPtr, ok := target.(**field_types.S3bucketField)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("mapToS3Field: target type %T is not **field_types.S3bucketField", target))
	}
	if *targetPtr == nil {
		*targetPtr = &field_types.S3bucketField{}
	}
	if (*targetPtr).Value == nil {
		(*targetPtr).Value = &map[string]field_types.S3bucketItemAndRules{}
	}

	targetMap := *((*targetPtr).Value)

	for _, key := range composer.SortedMapKeys(info.Items) {
		item := info.Items[key]

		if item == nil {
			continue
		}

		s3item := field_types.S3bucketItemValue{
			Bucket: convutils.Ref(item.Bucket),
			Path:   convutils.Ref(item.Path),
		}

		if item.Url != "" {
			s3item.Url = convutils.Ref(item.Url)
		}

		if item.SecretName != "" {
			s3item.SecretName = convutils.Ref(item.SecretName)
			s3item.Key = convutils.Ref(item.AccessKeyIdKey)
			s3item.Secret = convutils.Ref(item.SecretAccessKeyKey)

			if item.SessionToken != "" {
				s3item.SessionToken = convutils.Ref(item.SessionToken)
			}
		}

		targetMap[key] = field_types.S3bucketItemAndRules{
			Value: &s3item,
		}
	}

	return nil
}

func mapToConfigMapField(src map[string]*workloadsApi.ConfigMapInstance, target any, _ any) *workloadsApi.Error {
	if len(src) == 0 {
		return nil
	}

	targetPtr, ok := target.(**field_types.ConfigMapVolumeField)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("mapToConfigMapField: target type %T is not **field_types.ConfigMapVolumeField", target))
	}
	if *targetPtr == nil {
		*targetPtr = &field_types.ConfigMapVolumeField{}
	}
	if (*targetPtr).Value == nil {
		(*targetPtr).Value = &map[string]field_types.ConfigMapVolumeItemAndRules{}
	}

	targetMap := *((*targetPtr).Value)

	for _, key := range composer.SortedMapKeys(src) {
		instance := src[key]

		if instance == nil {
			continue
		}

		item := field_types.ConfigMapVolumeItemValue{
			Name:      instance.ConfigMap.Get(),
			MountPath: convutils.EmptyIfNull(instance.MountPath.Get()),
		}

		if convutils.IsStringHasValue(instance.SubPath.Get()) {
			item.SubPath = instance.SubPath.Get()
		}

		if convutils.IsStringHasValue(instance.DefaultMode.Get()) {
			item.DefaultMode = instance.DefaultMode.Get()
		}

		targetMap[key] = field_types.ConfigMapVolumeItemAndRules{
			Value: &item,
		}
	}

	return nil
}

func mapToEmptyDirField(src map[string]*workloadsApi.EmptyDirInstance, target any, _ any) *workloadsApi.Error {
	if len(src) == 0 {
		return nil
	}

	targetPtr, ok := target.(**field_types.EmptyDirVolumeField)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("mapToEmptyDirField: target type %T is not **field_types.EmptyDirVolumeField", target))
	}
	if *targetPtr == nil {
		*targetPtr = &field_types.EmptyDirVolumeField{}
	}
	if (*targetPtr).Value == nil {
		(*targetPtr).Value = &map[string]field_types.EmptyDirVolumeItemAndRules{}
	}

	targetMap := *((*targetPtr).Value)

	for _, key := range composer.SortedMapKeys(src) {
		instance := src[key]

		if instance == nil {
			continue
		}

		item := field_types.EmptyDirVolumeItemValue{
			Path: convutils.EmptyIfNull(instance.Path.Get()),
		}

		if convutils.HasValue(instance.Medium) {
			item.Medium = instance.Medium.Get()
		}

		if convutils.HasValue(instance.SizeLimit) {
			item.SizeLimit = instance.SizeLimit.Get()
		}

		targetMap[key] = field_types.EmptyDirVolumeItemAndRules{
			Value: &item,
		}
	}

	return nil
}

func mapToDataVolumeField(src map[string]*workloadsApi.DataVolumeInstance, target any, _ any) *workloadsApi.Error {
	if len(src) == 0 {
		return nil
	}

	targetPtr, ok := target.(**field_types.DataVolumeField)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("mapToDataVolumeField: target type %T is not **field_types.DataVolumeField", target))
	}
	if *targetPtr == nil {
		*targetPtr = &field_types.DataVolumeField{}
	}
	if (*targetPtr).Value == nil {
		(*targetPtr).Value = &map[string]field_types.DataVolumeItemAndRules{}
	}

	targetMap := *((*targetPtr).Value)

	for _, key := range composer.SortedMapKeys(src) {
		instance := src[key]

		if instance == nil {
			continue
		}

		item := field_types.DataVolumeItemValue{
			Id:        convutils.EmptyIfNull(instance.Id.Get()),
			MountPath: convutils.EmptyIfNull(instance.MountPath.Get()),
		}

		targetMap[key] = field_types.DataVolumeItemAndRules{
			Value: &item,
		}
	}

	return nil
}

func mapToSecretVolumeField(src map[string]*workloadsApi.SecretInstance, target any, _ any) *workloadsApi.Error {
	if len(src) == 0 {
		return nil
	}

	targetPtr, ok := target.(**field_types.SecretVolumeField)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("mapToSecretVolumeField: target type %T is not **field_types.SecretVolumeField", target))
	}
	if *targetPtr == nil {
		*targetPtr = &field_types.SecretVolumeField{}
	}
	if (*targetPtr).Value == nil {
		(*targetPtr).Value = &map[string]field_types.SecretVolumeItemAndRules{}
	}

	targetMap := *((*targetPtr).Value)

	for _, key := range composer.SortedMapKeys(src) {
		instance := src[key]

		if instance == nil {
			continue
		}

		item := field_types.SecretVolumeItemValue{
			Name:      instance.Secret.Get(),
			MountPath: convutils.EmptyIfNull(instance.MountPath.Get()),
		}

		if convutils.IsStringHasValue(instance.DefaultMode.Get()) {
			item.DefaultMode = instance.DefaultMode.Get()
		}

		targetMap[key] = field_types.SecretVolumeItemAndRules{
			Value: &item,
		}
	}

	return nil
}
