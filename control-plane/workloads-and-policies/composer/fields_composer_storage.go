/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	corev1 "k8s.io/api/core/v1"
)

const (
	NoneStorageClass = "runai-none"

	HomeVolumeName      = "user-data"
	RunaiDefaultHomeDir = "/home/runai-home"
	HomeEnvVarName      = "HOME"

	DshmVolumeName = "dshm"
	DshmMountPath  = "/dev/shm"
)

func (fc *FieldsComposer) ComposeHostPath(prefix string, src []workloadsApi.HostPathInstance,
	target any, setter FnMapToHostPathSetter, cookie any, mountPropagation *workloadsApi.HostPathMountPropagation,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if len(src) == 0 {
			return nil
		}

		entries := make(map[string]*workloadsApi.HostPathInstance)

		// we validate that all volumes have the same mountPropagation value, as this setting is workload-wide in the
		// cluster, rather than per volume
		var prevMountPropagation workloadsApi.HostPathMountPropagation

		for _, instance := range src {
			itemId := fc.ItemizedNumbering.Compose(prefix, instance.Name.Get(), nil, false)

			if convutils.BoolIsTrue(instance.Exclude.Get()) {
				entries[itemId] = nil
				continue
			} else {
				entries[itemId] = &instance
			}

			// set curr to None
			currMountPropagation := workloadsApi.HOSTPATHMOUNTPROPAGATION_NONE
			if convutils.HasValue(instance.MountPropagation) {
				currMountPropagation = *instance.MountPropagation.Get()
			}

			if prevMountPropagation != "" && prevMountPropagation != currMountPropagation {
				return workloadsApi.NewError(http.StatusBadRequest, "hostPath: all volumes must have the same mountPropagation value")
			}
			prevMountPropagation = currMountPropagation
		}

		*mountPropagation = prevMountPropagation

		return setter(entries, target, cookie)
	}
}

// ComposeMountPropagation is relevant for workloads that demand same mount propagation for all the volumes.
// it sets the workload-wide mount propagation to true, if all volumes are host-to-container
func (fc *FieldsComposer) ComposeMountPropagation(src *workloadsApi.HostPathMountPropagation,
	target any, setter FnBooleanSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src != nil && *src == workloadsApi.HOSTPATHMOUNTPROPAGATION_HOST_TO_CONTAINER {
			return setter(convutils.Ref(true), target, cookie)
		}

		return nil
	}
}

// ComposeHomeDirectory creates an emptyDir volume and mount for the home directory
// when createHomeDir is true
func (fc *FieldsComposer) ComposeHomeDirectory(createHomeDir *bool,
	target any, setter FnMapToEmptyDirSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if !convutils.BoolIsTrue(createHomeDir) {
			return nil
		}

		homeEmptyDir := workloadsApi.EmptyDirInstance{
			Name: *workloadsApi.NewNullableString(convutils.Ref(HomeVolumeName)),
			Path: *workloadsApi.NewNullableString(convutils.Ref(RunaiDefaultHomeDir)),
		}

		entries := map[string]*workloadsApi.EmptyDirInstance{
			HomeVolumeName: &homeEmptyDir,
		}

		return setter(entries, target, cookie)
	}
}

// ComposeHomeDirEnvVar injects the runai-managed home directory as the HOME env var
// when createHomeDir is requested. Intended to run AFTER the user environment variables
// have been written to target; the setter is responsible for not clobbering an existing
// env var of the same name (see EnvVarsToK8sEnv).
func (fc *FieldsComposer) ComposeHomeDirEnvVar(createHomeDir *bool,
	target any, setter FnEnvVarSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if !convutils.BoolIsTrue(createHomeDir) {
			return nil
		}

		return setter([]EnvVarInfo{{
			Name:      HomeEnvVarName,
			ValueType: EnvValueText,
			ValueFrom: RunaiDefaultHomeDir,
		}}, target, cookie)
	}
}

// ComposeLargeShm creates an emptyDir volume with Memory medium and mount at /dev/shm
// when largeShmRequest is true
func (fc *FieldsComposer) ComposeLargeShm(largeShmRequest *bool,
	target any, setter FnMapToEmptyDirSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if !convutils.BoolIsTrue(largeShmRequest) {
			return nil
		}

		dshmEmptyDir := workloadsApi.EmptyDirInstance{
			Name:   *workloadsApi.NewNullableString(convutils.Ref(DshmVolumeName)),
			Path:   *workloadsApi.NewNullableString(convutils.Ref(DshmMountPath)),
			Medium: *workloadsApi.NewNullableString(convutils.Ref("Memory")),
		}

		entries := map[string]*workloadsApi.EmptyDirInstance{
			DshmVolumeName: &dshmEmptyDir,
		}

		return setter(entries, target, cookie)
	}
}

func (fc *FieldsComposer) ComposeNfs(prefix string, src []workloadsApi.NfsInstance,
	target any, setter FnMapToNfsSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if len(src) == 0 {
			return nil
		}

		entries := make(map[string]*workloadsApi.NfsInstance)

		for _, instance := range src {
			itemId := fc.ItemizedNumbering.Compose(prefix, instance.Name.Get(), nil, false)

			if convutils.BoolIsTrue(instance.Exclude.Get()) {
				entries[itemId] = nil
				continue
			} else {
				entries[itemId] = &instance
			}
		}

		return setter(entries, target, cookie)
	}
}

const (
	defaultSecretDataKeyUser     = "User"
	defaultSecretDataKeyPassword = "Password"
)

// The missing image is transient — older cluster, or before the first cluster-sync
// after upgrade — so we surface a retryable service-unavailable instead
// of shipping a wrong default image. See the
// init-container injection design doc.
func ErrClusterValueMissing(field string) *workloadsApi.Error {
	return workloadsApi.NewError(http.StatusServiceUnavailable,
		fmt.Sprintf("cluster has not reported %q; it must sync its configuration before this workload can be submitted", field))
}

func ErrClusterValueInvalid(field string, err error) *workloadsApi.Error {
	return workloadsApi.NewError(http.StatusServiceUnavailable,
		fmt.Sprintf("cluster reported invalid %q: %v", field, err))
}

func (fc *FieldsComposer) ComposeGit(prefix string, src []workloadsApi.GitInstance,
	target any, setter FnGitSyncSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if len(src) == 0 {
			return nil
		}

		cfg := fc.ClusterStorageConfig
		if cfg == nil {
			cfg = DefaultClusterStorageConfig()
		}

		info := &GitSyncInfo{
			Items:  make(map[string]*GitSyncItem, len(src)),
			Config: cfg,
		}

		for _, instance := range src {
			itemId := fc.ItemizedNumbering.Compose(prefix, instance.Name.Get(), nil, false)

			if convutils.BoolIsTrue(instance.Exclude.Get()) {
				info.Items[itemId] = nil
				continue
			}

			if convutils.HasValue(instance.Revision) && convutils.HasValue(instance.Branch) {
				return workloadsApi.NewError(http.StatusBadRequest, "To point to your Git revision, provide only one of the following: ‘branch’ OR ‘revision’")
			}

			if convutils.HasValue(instance.PasswordSecret) && convutils.HasValue(instance.SecretRef) {
				return workloadsApi.NewError(http.StatusBadRequest, "git: passwordSecret is mutually exclusive with secretRef")
			}

			item := &GitSyncItem{
				Nick:       itemId,
				Repository: convutils.Deref(instance.Repository.Get(), ""),
				Branch:     convutils.Deref(instance.Branch.Get(), ""),
				Revision:   convutils.Deref(instance.Revision.Get(), ""),
				Directory:  convutils.Deref(instance.Path.Get(), ""),
			}

			if convutils.HasValue(instance.PasswordSecret) {
				item.SecretName = convutils.Deref(instance.PasswordSecret.Get(), "")
				item.Username = convutils.Deref(instance.SecretKeyOfUser.Get(), defaultSecretDataKeyUser)
				item.Password = convutils.Deref(instance.SecretKeyOfPassword.Get(), defaultSecretDataKeyPassword)
			}

			if convutils.HasValue(instance.SecretRef) {
				secretRef := instance.SecretRef.Get()

				if convutils.IsEmptyOrNull(secretRef.Name) {
					return workloadsApi.NewError(http.StatusBadRequest, "git: secret's name required")
				}

				if convutils.IsEmptyOrNull(EnumToString(secretRef.AuthenticationMethod)) {
					return workloadsApi.NewError(http.StatusBadRequest, "git: authenticatingMethod is required")
				}

				authMethod := *secretRef.AuthenticationMethod
				item.SecretName = convutils.Deref(secretRef.Name, "")

				switch authMethod {
				case workloadsApi.GITAUTHENTICATIONMETHOD_PASSWORD:
					// when the cluster only supports password auth, the method is implicit and omitted.
					if fc.Options.GitSyncWithSshKeySupported {
						item.AuthenticatingMethod = string(authMethod)
					}
					item.Username = convutils.Deref(secretRef.SecretKeyOfUser.Get(), defaultSecretDataKeyUser)
					item.Password = convutils.Deref(secretRef.SecretKeyOfPassword.Get(), defaultSecretDataKeyPassword)
				case workloadsApi.GITAUTHENTICATIONMETHOD_SSH_KEY:
					if !fc.Options.GitSyncWithSshKeySupported {
						return workloadsApi.NewError(http.StatusBadRequest, "git: authenticatingMethod is not supported in this cluster")
					}
					item.AuthenticatingMethod = string(authMethod)
				default:
					return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("git: unsupported authenticatingMethod '%s'", authMethod))
				}
			}

			info.Items[itemId] = item
		}

		return setter(info, target, cookie)
	}
}

func (fc *FieldsComposer) ComposePvc(prefix string, src []workloadsApi.PvcInstance,
	target any, setter FnMapToPvcSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if len(src) == 0 {
			return nil
		}

		entries := make(map[string]*workloadsApi.PvcInstance)

		for _, instance := range src {
			itemId := fc.ItemizedNumbering.Compose(prefix, instance.Name.Get(), nil, false)

			if convutils.BoolIsTrue(instance.Exclude.Get()) {
				entries[itemId] = nil
				continue
			}

			if fc.WorkloadType == string(workloadsApi.WORKLOADTYPE_INFERENCE) {
				// for inference workloads:
				// ephemeral must not be set 'true' as it is not applicable for inference workloads
				// existingPvc must be set 'true' only for cluster version >= 2.20, as workload volumes were not supported by the controller before that

				if convutils.HasValue(instance.Ephemeral) && *instance.Ephemeral.Get() {
					return workloadsApi.NewError(http.StatusBadRequest, "pvc: Ephemeral must not be set 'true' for inference workloads")
				}
			}

			existingPvc := convutils.BoolIsTrue(instance.ExistingPvc.Get())

			if !convutils.HasValue(instance.ClaimName) {
				if existingPvc {
					if !fc.Options.IgnoreMissingValues {
						return workloadsApi.NewError(http.StatusBadRequest, "claimName is mandatory for existing PVC")
					}
				} else {
					// NOTE!
					// this name is random, so if needed for other composers, you'll have to return the chosen name
					// from this function so it can be used consistently by the other composers
					instance.ClaimName = *workloadsApi.NewNullableString(convutils.Ref(fmt.Sprintf("%s-%s-%s", fc.WorkloadName,
						workloadsApi.SCOPETYPE_PROJECT, strings.ToLower(testutils.RandString(5)))))
				}
			}

			if !existingPvc {
				if !convutils.HasValue(instance.ClaimInfo) && !fc.Options.IgnoreMissingValues {
					return workloadsApi.NewError(http.StatusBadRequest, "claimInfo is mandatory for non-existing PVC")
				}

				if fc.StorageClassConfiguration != nil {
					storageClassName := convutils.Deref(instance.ClaimInfo.Get().StorageClass.Get(), "")
					if storageClassName == "" {
						storageClassName = NoneStorageClass
					}

					storageClassConfig, configDefined := fc.StorageClassConfiguration[storageClassName]

					if configDefined {
						if !convutils.HasValue(instance.ClaimInfo) {
							instance.ClaimInfo = *workloadsApi.NewNullableClaimInfo(&workloadsApi.ClaimInfo{})
						}

						var claimInfo *workloadsApi.ClaimInfo
						if convutils.HasValue(instance.ClaimInfo) {
							claimInfo = instance.ClaimInfo.Get()
						}

						configErr := ApplyStorageClassConfig(storageClassName, &claimInfo, &storageClassConfig)
						if configErr != nil {
							return configErr
						}

						// when applying storage class configuration attributes can change so we have to
						// update the claim info with the updated value
						if claimInfo != nil {
							instance.ClaimInfo.Set(claimInfo)
						}
					}
				}

				if convutils.HasValue(instance.ClaimInfo) && !convutils.HasValue(instance.ClaimInfo.Get().Size) && !fc.Options.IgnoreMissingValues {
					return workloadsApi.NewError(http.StatusBadRequest, "pvcSize is mandatory for non-existing PVC")
				}
			}

			entries[itemId] = &instance
		}

		return setter(entries, target, cookie)
	}
}

func (fc *FieldsComposer) ComposeS3(prefix string, src []workloadsApi.S3Instance,
	uidGid UidGid, target any, setter FnS3Setter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if len(src) == 0 {
			return nil
		}

		cfg := fc.ClusterStorageConfig
		if cfg == nil {
			cfg = DefaultClusterStorageConfig()
		}

		info := &S3Info{
			Items:  make(map[string]*S3Item, len(src)),
			Config: cfg,
		}

		if uidGid.Uid != nil && *uidGid.Uid != 0 {
			info.RunAsUid = convutils.Ref(*uidGid.Uid)
		}
		if uidGid.Gid != nil && *uidGid.Gid != 0 {
			info.RunAsGid = convutils.Ref(*uidGid.Gid)
		}

		for _, instance := range src {
			itemId := fc.ItemizedNumbering.Compose(prefix, instance.Name.Get(), nil, false)

			if convutils.BoolIsTrue(instance.Exclude.Get()) {
				info.Items[itemId] = nil
				continue
			}

			if err := ValidateUrlSyntax(instance.Url.Get(), false); err != nil {
				return err
			}

			if err := ValidatePathSyntax(instance.Path.Get()); err != nil {
				return err
			}

			item := &S3Item{
				Bucket: convutils.Deref(instance.Bucket.Get(), ""),
				Path:   convutils.Deref(instance.Path.Get(), ""),
				Url:    convutils.Deref(instance.Url.Get(), ""),
			}

			if convutils.HasValue(instance.AccessKeySecret) {
				item.SecretName = convutils.Deref(instance.AccessKeySecret.Get(), "")
				item.AccessKeyIdKey = convutils.Deref(instance.SecretKeyOfAccessKeyId.Get(), S3DefaultAccessKeyIdKey)
				item.SecretAccessKeyKey = convutils.Deref(instance.SecretKeyOfSecretKey.Get(), S3DefaultSecretAccessKeyKey)
				item.SessionToken = convutils.Deref(instance.AwsSessionToken.Get(), "")
			}

			info.Items[itemId] = item
		}

		return setter(info, target, cookie)
	}
}

func (fc *FieldsComposer) ComposeConfigMap(prefix string, src []workloadsApi.ConfigMapInstance,
	target any, setter FnMapToConfigMapSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if len(src) == 0 {
			return nil
		}

		entries := make(map[string]*workloadsApi.ConfigMapInstance)

		for _, instance := range src {
			itemId := fc.ItemizedNumbering.Compose(prefix, instance.Name.Get(), nil, false)

			if convutils.BoolIsTrue(instance.Exclude.Get()) {
				entries[itemId] = nil
				continue
			}

			if !instance.ConfigMap.IsSet() {
				return workloadsApi.NewError(http.StatusBadRequest, "configMap is a required field")
			}

			if convutils.IsStringHasValue(instance.DefaultMode.Get()) {
				if !fc.Options.VolumeDefaultModeSupported {
					return workloadsApi.NewError(http.StatusBadRequest, "storage.configMap.defaultMode is not supported in this cluster")
				}
			}

			entries[itemId] = &instance
		}

		return setter(entries, target, cookie)
	}
}

func (fc *FieldsComposer) ComposeEmptyDir(prefix string, src []workloadsApi.EmptyDirInstance,
	target any, setter FnMapToEmptyDirSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if len(src) == 0 {
			return nil
		}

		entries := make(map[string]*workloadsApi.EmptyDirInstance)

		for _, instance := range src {
			itemId := fc.ItemizedNumbering.Compose(prefix, instance.Name.Get(), nil, false)

			if convutils.BoolIsTrue(instance.Exclude.Get()) {
				entries[itemId] = nil
				continue
			}

			if !fc.Options.EmptyDirAsVolumeSupported {
				return workloadsApi.NewError(http.StatusBadRequest, "storage.emptyDir is not supported in this cluster")
			}

			if !convutils.HasValue(instance.Path) && !fc.Options.IgnoreMissingValues {
				return workloadsApi.NewError(http.StatusBadRequest, "path is mandatory for emptyDir")
			}

			if convutils.HasValue(instance.Medium) && *instance.Medium.Get() != string(corev1.StorageMediumMemory) {
				return workloadsApi.NewError(http.StatusBadRequest, "emptyDir medium must be either empty or 'Memory'")
			}

			entries[itemId] = &instance
		}

		return setter(entries, target, cookie)
	}
}

func (fc *FieldsComposer) ComposeDataVolume(prefix string, src []workloadsApi.DataVolumeInstance,
	target any, setter FnMapToDataVolumeSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if len(src) == 0 {
			return nil
		}

		entries := make(map[string]*workloadsApi.DataVolumeInstance)

		for _, instance := range src {
			itemId := fc.ItemizedNumbering.Compose(prefix, instance.Id.Get(), nil, false)

			if convutils.BoolIsTrue(instance.Exclude.Get()) {
				entries[itemId] = nil
				continue
			}

			if !convutils.HasValue(instance.Id) && !fc.Options.IgnoreMissingValues {
				return workloadsApi.NewError(http.StatusBadRequest, "dataVolume: id is mandatory")
			}

			if !convutils.HasValue(instance.MountPath) && !fc.Options.IgnoreMissingValues {
				return workloadsApi.NewError(http.StatusBadRequest, "dataVolume: mountPath is mandatory")
			}

			entries[itemId] = &instance
		}

		return setter(entries, target, cookie)
	}
}

func (fc *FieldsComposer) ComposeSecretVolume(prefix string, src []workloadsApi.SecretInstance,
	target any, setter FnMapToSecretVolumeSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if len(src) == 0 {
			return nil
		}

		entries := make(map[string]*workloadsApi.SecretInstance)

		for _, instance := range src {
			itemId := fc.ItemizedNumbering.Compose(prefix, instance.Name.Get(), nil, false)

			if convutils.BoolIsTrue(instance.Exclude.Get()) {
				entries[itemId] = nil
				continue
			}

			if convutils.IsStringHasValue(instance.DefaultMode.Get()) {
				if !fc.Options.VolumeDefaultModeSupported {
					return workloadsApi.NewError(http.StatusBadRequest, "storage.secret.defaultMode is not supported in this cluster")
				}
			}

			entries[itemId] = &instance
		}

		return setter(entries, target, cookie)
	}
}

var filePathPattern = "^(/)?([^/\\0]+(/)?)+$"
var filePathRegexp = regexp.MustCompile(filePathPattern)

func ValidatePathSyntax(path *string) *workloadsApi.Error {
	if path == nil {
		return nil
	}

	if !filePathRegexp.MatchString(*path) {
		return workloadsApi.NewError(http.StatusBadRequest, "maformed file path")
	}

	return nil
}
