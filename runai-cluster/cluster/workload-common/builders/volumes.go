/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants/oauth2proxy"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/utils/pointer"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	dshmVolumeName = "dshm"
	dshmPath       = "/dev/shm"

	homeVolumeName = "user-data"

	s3mountPrefix = "s3"

	EnvVarPrefixCARunAi = "CUSTOM_CA"
	EnvVarPrefixCAGit   = "CUSTOM_CA_GIT"
	EnvVarPrefixCAS3    = "CUSTOM_CA_S3"
)
const pvcVolumeNameFormat = "pvc-volume-%d"

var HostToContainer = corev1.MountPropagationHostToContainer

// GetVolumes returns the list of volumes which the job requests. Note that it does not create
// any resource, this is the responsibility of the caller to this function
// (typically submit function of some sort)
func GetVolumes(ctx context.Context, storageFields *field_groups.StorageFields, containerFields *field_groups.ContainerFields, networkingFields *field_groups.NetworkingFields, k8sClient client.Client, namespace string,
	resourceBaseName string, nameSuffix string, isJupyter bool, isS3Supported bool, isOauth2ProxyRequired bool) ([]corev1.Volume, error) {

	var v []corev1.Volume

	pvcVolumes, err := GetPvcVolumes(ctx, resourceBaseName, nameSuffix, storageFields)
	if err != nil {
		return nil, err
	}

	dataVolumes, err := GetDataVolumes(ctx, storageFields, containerFields, k8sClient, namespace)
	if err != nil {
		return nil, err
	}

	hostPathVolumes := GetHostPathVolumes(ctx, storageFields)
	nfsVolumes := GetNfsVolumes(ctx, storageFields)
	s3Volumes := GetS3Volumes(ctx, storageFields)
	configMapVolumes := GetConfigMapVolumes(ctx, storageFields.ConfigMapVolumes)
	secretVolumes := GetSecretVolumes(ctx, storageFields.SecretVolumes)
	emptyDirVolumes, err := GetEmptyDirVolumes(storageFields.EmptyDirVolumes)
	if err != nil {
		return nil, err
	}

	if !isS3Supported && len(s3Volumes) > 0 {
		return nil, fmt.Errorf("this workload type does not support s3 volumes")
	}

	v = append(v, hostPathVolumes...)
	v = append(v, nfsVolumes...)
	v = append(v, pvcVolumes...)
	v = append(v, dataVolumes...)
	v = append(v, s3Volumes...)
	v = append(v, configMapVolumes...)
	v = append(v, secretVolumes...)
	v = append(v, emptyDirVolumes...)

	//shm
	if storageFields.LargeShm.Deref() {
		v = append(v, corev1.Volume{
			Name: dshmVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					Medium: corev1.StorageMediumMemory,
				},
			},
		})
	}
	//home-dir
	if NeedToSetupRunaiHome(containerFields, isJupyter) {
		v = append(v, corev1.Volume{
			Name: homeVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}

	//git-sync
	gitVolumes := storageFields.GitSync.Deref()

	if len(gitVolumes) > 0 {
		for gitIdx, nick := range utils.SortedKeys(gitVolumes) {
			volume := gitVolumes[nick]
			if volume.Value == nil {
				continue
			}

			v = append(v, corev1.Volume{
				Name: fmt.Sprintf("%s-%d", GitSyncNamePrefix, gitIdx),
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			})

			if isGitSSHAuthenticationSelected(volume.Value) {
				v = append(v, corev1.Volume{
					Name: fmt.Sprintf("%s-%d", gitSyncSSHKeyNamePrefix, gitIdx),
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: ptr.Deref(volume.Value.SecretName, ""),
						},
					},
				})
			}
		}

		if config.IsCustomCAEnabled(EnvVarPrefixCAGit) {
			v = append(v, ComposeCAVolume(EnvVarPrefixCAGit))
		}
	}

	if len(s3Volumes) > 0 {
		if config.IsCustomCAEnabled(EnvVarPrefixCAS3) {
			v = append(v, ComposeCAVolume(EnvVarPrefixCAS3))
		}
	}

	// Create volumes for proxy authorization (auth volume + optional CA volume)
	if (config.ExternalAuthUrlEnabled() && len(networkingFields.ExposedUrls.Deref()) != 0) || isOauth2ProxyRequired {
		v = append(v, ComposeAuthProxyVolumes()...)
	}

	return v, nil
}

// GetPvcVolumes returns list of PVCs which the job requests. Note that it does not create
// the PVC themselves, this is the responsibility of the caller to this function
// (typically submit function of some sort) and is implemented in createPvcs
func GetPvcVolumes(_ context.Context, resourceBaseName string, nameSuffix string, storageFields *field_groups.StorageFields) ([]corev1.Volume, error) {

	pvcs := storageFields.Pvcs.Deref()

	result := make([]corev1.Volume, 0, len(pvcs))

	idx := 0
	// Required to assign a consistent volume name for each claimName.
	// When a single PVC is mounted to multiple paths in the same pod,
	// Kubernetes requires all volumeMounts to reference the same volume name.
	// Otherwise, workload creation fails due to a known k8s bug (kubernetes/kubernetes#127004).
	claimNameToPvcVolumeName := map[string]string{}

	for _, nick := range utils.SortedKeys(pvcs) {

		pvc := pvcs[nick].Value

		// if Ephemeral is set to true, we don't create a PVC for it
		if pvc == nil || pointer.BoolDeref(pvc.Ephemeral, false) {
			continue
		}

		claimName := GetPvcClaimName(pvc, resourceBaseName, nameSuffix, idx)

		if pointer.BoolDeref(pvc.ExistingPvc, false) {
			if ptr.Deref(pvc.ClaimName, "") == "" {
				return nil, fmt.Errorf("%s: claimName is mandatory when binding to existing PVC", nick)
			}
			if ptr.Deref(pvc.Size, "") != "" {
				return nil, fmt.Errorf("%s: size is redundant when binding to existing PVC", nick)
			}
		} else {
			if ptr.Deref(pvc.Size, "") == "" {
				return nil, fmt.Errorf("%s: size is mandatory for creating a PVC", nick)
			}
		}

		if _, exists := claimNameToPvcVolumeName[claimName]; !exists {
			claimNameToPvcVolumeName[claimName] = fmt.Sprintf(pvcVolumeNameFormat, idx)
			result = append(result, corev1.Volume{
				Name: claimNameToPvcVolumeName[claimName],
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: claimName,
						ReadOnly:  pointer.BoolDeref(pvc.ReadOnly, false),
					},
				},
			})
			idx += 1
		}
	}

	return result, nil
}

// GetDataVolumes returns list of DataVolumes which the job requests.
func GetDataVolumes(ctx context.Context, storageFields *field_groups.StorageFields, containerFields *field_groups.ContainerFields, k8sClient client.Client, namespace string) ([]corev1.Volume, error) {

	dataVolumes := storageFields.DataVolumes.Deref()

	if len(dataVolumes) == 0 {
		return nil, nil
	}

	result := make([]corev1.Volume, 0, len(dataVolumes))

	idx := 0

	for _, nick := range utils.SortedKeys(dataVolumes) {

		dataVolume := dataVolumes[nick].Value
		if dataVolume == nil {
			continue
		}

		dataVolumeId := ptr.Deref(dataVolume.Id, "")
		if dataVolumeId == "" {
			return nil, fmt.Errorf("DataVolume id is missing")
		}

		dataVolumeMountPath := ptr.Deref(dataVolume.MountPath, "")
		if dataVolumeMountPath == "" {
			return nil, fmt.Errorf("DataVolume mountPath is missing")
		}

		claim, err := getDataVolumeNamespacePvc(ctx, dataVolumeId, namespace, k8sClient)
		if err != nil {
			return nil, err
		}

		result = append(result, corev1.Volume{
			Name: fmt.Sprintf("data-volume-%d", idx),
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: claim.Name,
				},
			},
		})

		idx += 1
	}

	return result, nil
}

// getDataVolumeNamespacePvc returns the PVC of the DataVolume in the given namespace
func getDataVolumeNamespacePvc(ctx context.Context, dataVolumeId string, namespace string, k8sClient client.Client) (*corev1.PersistentVolumeClaim, error) {
	logger := log.FromContext(ctx)

	listOptions := client.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			v1alpha1.DataVolumeIdLabel: dataVolumeId,
		}),
		Namespace: namespace,
	}

	var claims corev1.PersistentVolumeClaimList
	err := k8sClient.List(ctx, &claims, &listOptions)
	if err != nil {
		return &corev1.PersistentVolumeClaim{}, err
	}
	// expect that claims will have only one item
	if len(claims.Items) == 0 {
		logger.Info("pvc from data volume not found", "dataVolumeId", dataVolumeId, "namespace", namespace)
		return nil, fmt.Errorf("DataVolume %s not found in namespace %s", dataVolumeId, namespace)
	}
	if len(claims.Items) > 1 {
		logger.Info("more than one pvc from data volume found", "dataVolumeId", dataVolumeId, "namespace", namespace)
		return nil, fmt.Errorf("DataVolume %s has more than one claim", dataVolumeId)
	}
	if claims.Items[0].DeletionTimestamp != nil {
		logger.Info("pvc from data volume is terminating", "dataVolumeId", dataVolumeId, "namespace", namespace)
		return nil, fmt.Errorf("DataVolume %s is terminating", dataVolumeId)
	}

	claim := claims.Items[0]
	return &claim, nil
}

// GetHostPathVolumes returns the list of Volumes which the job requests. Note that it does not create
// the Volume themselves, this is the responsibility of the caller to this function
// (typically submit function of some sort) and is implemented in createVolumes
func GetHostPathVolumes(_ context.Context, storageFields *field_groups.StorageFields) []corev1.Volume {

	volumes := storageFields.Volumes.Deref()

	result := make([]corev1.Volume, 0, len(volumes))

	idx := 0

	for _, nick := range utils.SortedKeys(volumes) {

		vol := volumes[nick].Value
		if vol == nil {
			continue
		}

		if ptr.Deref(vol.NfsServer, "") != "" {
			// Leaving the NFS handling for backward compatibility
			// but in general - this should be removed, as nfs now has its own field
			result = append(result, corev1.Volume{
				Name: fmt.Sprintf("host-path-volume-%d", idx),
				VolumeSource: corev1.VolumeSource{
					NFS: &corev1.NFSVolumeSource{
						Server: ptr.Deref(vol.NfsServer, ""),
						Path:   ptr.Deref(vol.SourcePath, ""),
					},
				},
			})
		} else {
			result = append(result, corev1.Volume{
				Name: fmt.Sprintf("host-path-volume-%d", idx),
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: ptr.Deref(vol.SourcePath, ""),
					},
				},
			})
		}

		idx += 1
	}
	return result
}

// GetNfsVolumes returns the list of nfs volumes which the job requests.
func GetNfsVolumes(_ context.Context, storageFields *field_groups.StorageFields) []corev1.Volume {

	nfsVolumes := storageFields.Nfs.Deref()

	result := make([]corev1.Volume, 0, len(nfsVolumes))

	idx := 0

	for _, nick := range utils.SortedKeys(nfsVolumes) {

		vol := nfsVolumes[nick].Value
		if vol == nil {
			continue
		}

		result = append(result, corev1.Volume{
			Name: fmt.Sprintf("nfs-volume-%d", idx),
			VolumeSource: corev1.VolumeSource{
				NFS: &corev1.NFSVolumeSource{
					Server: ptr.Deref(vol.NfsServer, ""),
					Path:   ptr.Deref(vol.SourcePath, ""),
				},
			},
		})

		idx += 1
	}
	return result
}

// IdentifyPvcsVolumeMode iterates the pvcs and sets the VolumeMode attribute for existing pvcs,
// as it is not necessarily defined in the Workload spec
func IdentifyPvcsVolumeMode(ctx context.Context, pvcs *flds.PvcsField, namespace string, k8sClient client.Client) error {
	if pvcs == nil {
		return nil
	}

	for i, pvc := range pvcs.Deref() {
		if pointer.BoolDeref(pvc.Deleted, false) {
			continue
		}
		if pointer.BoolDeref(pvc.Value.ExistingPvc, false) {
			// existing pvc, make sure claimName is set
			claimName := ptr.Deref(pvc.Value.ClaimName, "")
			if claimName == "" {
				return fmt.Errorf("claimName is mandatory when binding to existing PVC")
			}
			// existing pvc, get its VolumeMode
			claimKey := client.ObjectKey{
				Name:      claimName,
				Namespace: namespace,
			}

			var claim corev1.PersistentVolumeClaim
			err := k8sClient.Get(ctx, claimKey, &claim)
			if err != nil {
				return err
			}

			// copy the VolumeMode from the claim to our list
			if claim.Spec.VolumeMode != nil {
				volumeMode := string(*claim.Spec.VolumeMode)
				(*pvcs.Value)[i].Value.VolumeMode = &volumeMode
			}
		}
	}

	return nil
}

func GetVolumeMounts(ctx context.Context, storageFields *field_groups.StorageFields, containerFields *field_groups.ContainerFields, isJupyter bool, isEphemeralPvcSupported bool) ([]corev1.VolumeMount, error) {

	var v []corev1.VolumeMount

	volumes := storageFields.Volumes.Deref()
	nfsVolumes := storageFields.Nfs.Deref()
	pvcs := storageFields.Pvcs.Deref()
	dataVolumes := storageFields.DataVolumes.Deref()
	gitSyncs := storageFields.GitSync.Deref()

	//volume
	idx := 0

	for _, nick := range utils.SortedKeys(volumes) {

		volume := volumes[nick].Value
		if volume == nil {
			continue
		}

		vm := corev1.VolumeMount{
			Name:      fmt.Sprintf("host-path-volume-%d", idx),
			MountPath: ptr.Deref(volume.TargetPath, ""),
			ReadOnly:  pointer.BoolDeref(volume.ReadOnly, false),
		}

		if ptr.Deref(volume.NfsServer, "") != "" && storageFields.MountPropagation.Deref() {
			vm.MountPropagation = &HostToContainer
		}

		v = append(v, vm)

		idx += 1
	}

	//nfs
	idx = 0

	for _, nick := range utils.SortedKeys(nfsVolumes) {

		nfs := nfsVolumes[nick].Value
		if nfs == nil {
			continue
		}

		vm := corev1.VolumeMount{
			Name:      fmt.Sprintf("nfs-volume-%d", idx),
			MountPath: ptr.Deref(nfs.TargetPath, ""),
			ReadOnly:  pointer.BoolDeref(nfs.ReadOnly, false),
		}

		if storageFields.MountPropagation.Deref() {
			vm.MountPropagation = &HostToContainer
		}

		v = append(v, vm)

		idx += 1
	}

	//pvc
	idx = 0
	// Required to assign a consistent volume name for each claimName.
	// When a single PVC is mounted to multiple paths in the same pod,
	// Kubernetes requires all volumeMounts to reference the same volume name.
	// Otherwise, workload creation fails due to a known k8s bug (kubernetes/kubernetes#127004).
	claimNameToVolumeName := map[string]string{}

	for _, nick := range utils.SortedKeys(pvcs) {

		pvc := pvcs[nick].Value
		if pvc == nil {
			continue
		}

		// if Ephemeral is set to true, we don't create a PVC for it
		if pointer.BoolDeref(pvc.Ephemeral, false) {
			if !isEphemeralPvcSupported {
				return nil, fmt.Errorf("this workload type doesn't support ephemeral PVC")
			}

			v = append(v, corev1.VolumeMount{
				Name:      *pvc.ClaimName,
				MountPath: ptr.Deref(pvc.Path, ""),
				ReadOnly:  pointer.BoolDeref(pvc.ReadOnly, false),
			})
			continue
		}

		if pvcs[nick].Value != nil &&
			(pvcs[nick].Value.VolumeMode == nil ||
				*(pvcs[nick].Value.VolumeMode) == string(corev1.PersistentVolumeFilesystem)) {

			claimName := ptr.Deref(pvc.ClaimName, "")
			volumeName, found := claimNameToVolumeName[claimName]
			if !found {
				volumeName = fmt.Sprintf(pvcVolumeNameFormat, idx)
				idx++
				if claimName != "" {
					claimNameToVolumeName[claimName] = volumeName
				}
			}

			v = append(v, corev1.VolumeMount{
				Name:      volumeName,
				MountPath: ptr.Deref(pvc.Path, ""),
				ReadOnly:  pointer.BoolDeref(pvc.ReadOnly, false),
			})
		}
	}

	// dataVolume
	idx = 0

	for _, nick := range utils.SortedKeys(dataVolumes) {

		dataVolume := dataVolumes[nick].Value
		if dataVolume == nil {
			continue
		}

		v = append(v, corev1.VolumeMount{
			Name:      fmt.Sprintf("data-volume-%d", idx),
			MountPath: ptr.Deref(dataVolume.MountPath, ""),
			ReadOnly:  true,
		})

		idx += 1
	}

	// shm
	if storageFields.LargeShm.Deref() {
		v = append(v, corev1.VolumeMount{
			Name:      dshmVolumeName,
			MountPath: dshmPath,
		})
	}

	//home-dir
	if NeedToSetupRunaiHome(containerFields, isJupyter) {
		vm := corev1.VolumeMount{
			Name:      homeVolumeName,
			MountPath: constants.RunaiDefaultHomeDir,
		}

		if storageFields.MountPropagation.Deref() {
			vm.MountPropagation = &HostToContainer
		}

		v = append(v, vm)
	}

	//git-sync
	for idx, nick := range utils.SortedKeys(gitSyncs) {
		gitSyncParams := gitSyncs[nick].Value
		if gitSyncParams != nil {
			v = append(v, corev1.VolumeMount{
				Name:      fmt.Sprintf("%s-%d", GitSyncNamePrefix, idx),
				MountPath: ptr.Deref(gitSyncParams.Directory, fmt.Sprintf("%s-%s", GitSyncPathPrefix, nick)),
			})
		}
	}

	v = append(v, GetS3VolumeMounts(ctx, storageFields)...)

	// ConfigMap Volumes
	configMapVolumes := storageFields.ConfigMapVolumes.Deref()
	for i, nick := range utils.SortedKeys(configMapVolumes) {
		v = append(v, configMapVolumes[nick].Value.VolumeMount(flds.GenerateConfigMapVolumeName(i)))
	}

	// Secret Volumes
	secretVolumes := storageFields.SecretVolumes.Deref()
	for i, nick := range utils.SortedKeys(secretVolumes) {
		v = append(v, secretVolumes[nick].Value.VolumeMount(flds.GenerateSecretVolumeName(i)))
	}

	// EmptyDir Volumes
	emptyDirVolumes := storageFields.EmptyDirVolumes.Deref()
	idx = 0
	for _, nick := range utils.SortedKeys(emptyDirVolumes) {
		item := emptyDirVolumes[nick].Value
		if item == nil {
			continue
		}
		volName := fmt.Sprintf("emptydir-volume-%d", idx)
		mountPath := ""
		if item.Path != nil {
			mountPath = *item.Path
		}
		v = append(v, corev1.VolumeMount{
			Name:      volName,
			MountPath: mountPath,
		})
		idx++
	}

	return v, nil
}

func GetVolumeDevices(_ context.Context, storageFields *field_groups.StorageFields) []corev1.VolumeDevice {

	var v []corev1.VolumeDevice

	pvcs := storageFields.Pvcs.Deref()

	//pvc
	idx := 0

	for _, nick := range utils.SortedKeys(pvcs) {

		pvc := pvcs[nick].Value
		if pvc == nil {
			continue
		}

		// if Ephemeral is set to true, we don't create a PVC for it
		if pointer.BoolDeref(pvc.Ephemeral, false) {
			continue
		}

		if pvcs[nick].Value != nil &&
			pvcs[nick].Value.VolumeMode != nil &&
			*(pvcs[nick].Value.VolumeMode) == string(corev1.PersistentVolumeBlock) {
			v = append(v, corev1.VolumeDevice{
				Name:       fmt.Sprintf("pvc-volume-%d", idx),
				DevicePath: ptr.Deref(pvc.Path, ""),
			})
		}

		idx += 1
	}

	return v
}

func CreateEphemeralVolumes(pvcs *flds.PvcsField, volume *[]corev1.Volume, workloadName string) error {
	if pvcs == nil {
		return nil
	}

	pvcsMap := pvcs.Deref()

	idx := 0
	for _, nick := range utils.SortedKeys(pvcsMap) {
		pvc := pvcsMap[nick].Value
		if pvc == nil {
			continue
		}
		// if Ephemeral is set to false, we don't want to create an ephemeral volume
		if pvc == nil || !pointer.BoolDeref(pvc.Ephemeral, false) {
			continue
		}

		claimName := ptr.Deref(pvc.ClaimName, "")
		if claimName == "" {
			claimName = fmt.Sprintf("ephemeral-%s-%d", workloadName, idx)
		}

		if pointer.BoolDeref(pvc.ExistingPvc, false) {
			return fmt.Errorf("cannot use existing PVC %s as ephemeral", claimName)
		}

		pvcSize := ptr.Deref(pvc.Size, "")
		pvcValue, err := resource.ParseQuantity(pvcSize)

		if err != nil || pvcValue.Value() <= 0 {
			return fmt.Errorf("invalid size for creating a PVC %s: %s", claimName, pvcSize)
		}

		ephemeralVolume := corev1.Volume{
			Name: claimName,
			VolumeSource: corev1.VolumeSource{
				Ephemeral: &corev1.EphemeralVolumeSource{
					VolumeClaimTemplate: &corev1.PersistentVolumeClaimTemplate{
						Spec: corev1.PersistentVolumeClaimSpec{
							Resources: corev1.VolumeResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceStorage: resource.MustParse(ptr.Deref(pvc.Size, "")),
								},
							},
						},
					},
				},
			},
		}

		if ptr.Deref(pvc.StorageClass, "") != "" {
			ephemeralVolume.VolumeSource.Ephemeral.VolumeClaimTemplate.Spec.StorageClassName = pvc.StorageClass
		}

		accessModes := make([]corev1.PersistentVolumeAccessMode, 0)
		if pointer.BoolDeref(pvc.ReadOnlyMany, false) {
			accessModes = append(accessModes, corev1.ReadOnlyMany)
		}
		if pointer.BoolDeref(pvc.ReadWriteMany, false) {
			accessModes = append(accessModes, corev1.ReadWriteMany)
		}

		if len(accessModes) == 0 && pointer.BoolDeref(pvc.ReadWriteOnce, true) {
			// if no access mode was specified, we default to ReadWriteOnce
			accessModes = append(accessModes, corev1.ReadWriteOnce)
		} else if len(accessModes) != 0 && pointer.BoolDeref(pvc.ReadWriteOnce, false) {
			// if ReadWriteOnce was specified, we add it to the list of access modes
			accessModes = append(accessModes, corev1.ReadWriteOnce)
		}

		if len(accessModes) != 0 {
			ephemeralVolume.VolumeSource.Ephemeral.VolumeClaimTemplate.Spec.AccessModes = accessModes
		} else {
			return fmt.Errorf("no access mode was specified for PVC %s", claimName)
		}

		*volume = append(*volume, ephemeralVolume)

		idx += 1
	}

	return nil
}

func GetS3VolumeMounts(_ context.Context, storageFields *field_groups.StorageFields) []corev1.VolumeMount {

	s3buckets := storageFields.S3buckets.Deref()

	var result []corev1.VolumeMount

	mountHostToContainer := corev1.MountPropagationHostToContainer

	for s3idx, nick := range utils.SortedKeys(s3buckets) {

		s3bucket := s3buckets[nick].Value
		if s3bucket == nil {
			continue
		}

		result = append(result, corev1.VolumeMount{
			Name:             fmt.Sprintf("%s-%d", s3mountPrefix, s3idx),
			MountPath:        ptr.Deref(s3bucket.Path, ""),
			MountPropagation: &mountHostToContainer,
		})
	}

	return result
}

// GetConfigMapVolumes returns a slice of volume from ConfigMap fields
func GetConfigMapVolumes(_ context.Context, configMapsVolumes *flds.ConfigMapVolumeField) []corev1.Volume {
	items := configMapsVolumes.Deref()
	volumes := make([]corev1.Volume, 0, len(items))
	for i, nick := range utils.SortedKeys(items) {
		volumes = append(volumes, items[nick].Value.Volume(flds.GenerateConfigMapVolumeName(i)))
	}
	return volumes
}

// GetSecretVolumes returns a slice of volume from Secret fields
func GetSecretVolumes(_ context.Context, secretVolumes *flds.SecretVolumeField) []corev1.Volume {
	items := secretVolumes.Deref()
	volumes := make([]corev1.Volume, 0, len(items))
	for i, nick := range utils.SortedKeys(items) {
		volumes = append(volumes, items[nick].Value.Volume(flds.GenerateSecretVolumeName(i)))
	}
	return volumes
}

func GetS3Volumes(_ context.Context, storageFields *field_groups.StorageFields) []corev1.Volume {

	s3buckets := storageFields.S3buckets.Deref()

	result := make([]corev1.Volume, 0, len(s3buckets))

	for s3idx, nick := range utils.SortedKeys(s3buckets) {

		s3bucket := s3buckets[nick].Value
		if s3bucket == nil {
			continue
		}

		result = append(result, corev1.Volume{
			Name: fmt.Sprintf("%s-%d", s3mountPrefix, s3idx),
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})

	}
	return result
}

// NeedToSetupRunaiHome checks if it is necessary to mount and use runai home directory
// The conditions are
//  1. CreateHomeDir is true, or it is unset and RunAsUser is true. RunAsUser provides the
//     default only; an explicit CreateHomeDir value of false disables the home directory.
//  2. Job is not of type jupyter
//  3. HOME is not set or is set to the runai default
func NeedToSetupRunaiHome(containerFields *field_groups.ContainerFields, isJupyter bool) bool {
	if containerFields == nil || isJupyter {
		return false
	}

	shouldCreateHomeDir := containerFields.CreateHomeDir.Deref()
	// RunAsUser historically implies a managed home directory.
	// Fall back to RunAsUser only when CreateHomeDir is unset;
	// an explicit true or false value takes precedence.
	if !containerFields.CreateHomeDir.IsValueInitialized() {
		shouldCreateHomeDir = containerFields.RunAsUser.Deref()
	}

	if shouldCreateHomeDir {
		for envName, envValue := range containerFields.Environment.Deref() {
			if envName == constants.HomeDirEnvName {
				// need to setup runai HOME if the value of HOME is not the default runai directory
				return ptr.Deref(envValue.Value, "") == constants.RunaiDefaultHomeDir
			}
		}
		return true // user did not provide HOME environment
	}

	return false
}

// GetEmptyDirVolumes returns a slice of corev1.Volume for each entry in EmptyDirVolumes, or an error if validation fails
func GetEmptyDirVolumes(emptyDirVolumes *flds.EmptyDirVolumeField) ([]corev1.Volume, error) {
	if emptyDirVolumes == nil {
		return nil, nil
	}
	items := emptyDirVolumes.Deref()
	var result []corev1.Volume
	idx := 0
	for _, nick := range utils.SortedKeys(items) {
		item := items[nick].Value
		if item == nil {
			continue
		}
		if item.Path == nil || *item.Path == "" {
			return nil, fmt.Errorf("%s: path is mandatory for emptyDir volume", nick)
		}
		medium := ""
		if item.Medium != nil {
			medium = *item.Medium
		}
		if medium != "" && medium != string(corev1.StorageMediumMemory) {
			return nil, fmt.Errorf("%s: medium must be either empty or '%s' for emptyDir volume", nick, corev1.StorageMediumMemory)
		}
		var sizeLimit resource.Quantity
		var sizeLimitPtr *resource.Quantity
		if item.SizeLimit != nil && *item.SizeLimit != "" {
			parsed, err := resource.ParseQuantity(*item.SizeLimit)
			if err != nil {
				return nil, fmt.Errorf("%s: invalid sizeLimit for emptyDir volume: %s", nick, *item.SizeLimit)
			}
			sizeLimit = parsed
			sizeLimitPtr = &sizeLimit
		}
		result = append(result, corev1.Volume{
			Name: fmt.Sprintf("emptydir-volume-%d", idx),
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					Medium:    corev1.StorageMedium(medium),
					SizeLimit: sizeLimitPtr,
				},
			},
		})
		idx++
	}
	return result, nil
}

func ComposeCAVolume(envVarPrefix string) corev1.Volume {
	return corev1.Volume{
		Name: config.GetCustomCASecret(envVarPrefix),
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				DefaultMode: utils.Ref(int32(420)),
				Items: []corev1.KeyToPath{
					{Key: config.GetCustomCAKey(envVarPrefix), Path: config.GetCustomCAKey(envVarPrefix)},
				},
				Optional:   utils.Ref(true),
				SecretName: config.GetCustomCASecret(envVarPrefix),
			},
		},
	}
}

func ComposeCAVolumeMount(envVarPrefix string) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      config.GetCustomCASecret(envVarPrefix),
		MountPath: fmt.Sprintf("/etc/ssl/certs/%s", config.GetCustomCAKey(envVarPrefix)),
		SubPath:   config.GetCustomCAKey(envVarPrefix),
	}
}

// ComposeAuthVolume creates the emptyDir volume for storing authorized users file
func ComposeAuthVolume() corev1.Volume {
	return corev1.Volume{
		Name: constants.ExternalUrlAuthVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
}

// ComposeAuthProxyVolumes creates all volumes needed for OAuth2 proxy authentication
// Includes the auth volume and the configured templates and custom CA volumes.
func ComposeAuthProxyVolumes() []corev1.Volume {
	volumes := []corev1.Volume{ComposeAuthVolume()}
	if config.OAuth2ProxyTemplatesEnabled() {
		volumes = append(volumes, corev1.Volume{
			Name: oauth2proxy.VolumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: oauth2proxy.ConfigMapName},
					Items:                []corev1.KeyToPath{{Key: oauth2proxy.TemplateKey, Path: oauth2proxy.TemplateKey}},
					Optional:             utils.Ref(true),
				},
			},
		})
	}

	if config.IsCustomCAEnabled(EnvVarPrefixCARunAi) {
		volumes = append(volumes, ComposeCAVolume(EnvVarPrefixCARunAi))
	}

	return volumes
}
