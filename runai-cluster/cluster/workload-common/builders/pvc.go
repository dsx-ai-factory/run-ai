/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	"fmt"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	controllerutils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	commonutils "github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type PvcBuilder struct {
	k8sClient client.Client
}

func NewPvcBuilder(k8sClient client.Client) *PvcBuilder {
	return &PvcBuilder{
		k8sClient: k8sClient,
	}
}

// GetPreview returns a list of resources (name/kind only) that will be created on build
func (pb *PvcBuilder) GetPreview(ctx context.Context, objectMeta *metav1.ObjectMeta, workloadBaseName string, owner client.Object,
	pvcs *flds.PvcsField, nameSuffix string) ([]Resource, error) {
	var result []Resource

	idx := 0
	// this index is required in order to keep the claim names consistent with the ones used in the workload volumes.
	// it is incremented in the same cases as when creating volumes. so, if an index is skipped here as the PVC is
	// an existing one, it will also be skipped when creating the volume (even that there a volume is created even for existing PVCs).
	idxForPvcClaimName := 0

	for _, nick := range commonutils.SortedKeys(pvcs.Deref()) {
		pvc := pvcs.Deref()[nick].Value

		skip, promoteIndex := pb.skipPvcBuilding(pvc)
		if skip {
			if promoteIndex {
				idxForPvcClaimName += 1
			}
			continue
		}

		claimName := GetPvcClaimName(pvc, workloadBaseName, nameSuffix, idxForPvcClaimName)

		err := pb.IsValidPvc(ctx, objectMeta.Namespace, claimName, owner, objectMeta.Name)
		if err != nil {
			return nil, fmt.Errorf("invalid PVC %s: %v", claimName, err)
		}

		result = append(result, Resource{
			Kind: metav1.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    constants.KindPvc,
			},
			Name:      claimName,
			Namespace: objectMeta.Namespace,
		})

		idx += 1
		idxForPvcClaimName += 1
	}
	return result, nil
}

// Build creates the resources
func (pb *PvcBuilder) Build(ctx context.Context, jobType string, objectMeta *metav1.ObjectMeta, workloadBaseName string,
	owner client.Object, pvcs *flds.PvcsField, nameSuffix string) ([]corev1.PersistentVolumeClaim, error) {
	var result []corev1.PersistentVolumeClaim
	logger := log.FromContext(ctx)

	idx := 0
	// this index is required in order to keep the claim names consistent with the ones used in the workload volumes.
	// it is incremented in the same cases as when creating volumes. so, if an index is skipped here as the PVC is
	// an existing one, it will also be skipped when creating the volume (even that there a volume is created even for existing PVCs).
	idxForPvcClaimName := 0

	for _, nick := range commonutils.SortedKeys(pvcs.Deref()) {
		pvc := pvcs.Deref()[nick].Value

		skip, promoteIndex := pb.skipPvcBuilding(pvc)
		if skip {
			if promoteIndex {
				idxForPvcClaimName += 1
			}
			continue
		}

		claimName := GetPvcClaimName(pvc, workloadBaseName, nameSuffix, idxForPvcClaimName)

		err := pb.IsValidPvc(ctx, objectMeta.Namespace, claimName, owner, objectMeta.Name)
		if err != nil {
			return nil, fmt.Errorf("invalid PVC %s: %v", claimName, err)
		}

		logger.Info("PVC does not exist, need to create it", "claim", claimName)

		pvcSize := ptr.Deref(pvc.Size, "")
		pvcValue, err := resource.ParseQuantity(pvcSize)
		if err != nil || pvcValue.Value() <= 0 {
			return nil, fmt.Errorf("invalid size for creating a PVC %s: %s", claimName, pvcSize)
		}

		pvcClaim := corev1.PersistentVolumeClaim{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       constants.KindPvc,
			},
			ObjectMeta: metav1.ObjectMeta{
				CreationTimestamp: metav1.Time{Time: time.Now()},
				Name:              claimName,
				Namespace:         objectMeta.Namespace,
				Labels:            objectMeta.Labels,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: pvcValue,
					},
				},
			},
		}

		if pvcClaim.ObjectMeta.Labels == nil {
			pvcClaim.ObjectMeta.Labels = make(map[string]string)
		}
		pvcClaim.ObjectMeta.Labels[constants.LabelCreatedBy] = jobType

		if ptr.Deref(pvc.StorageClass, "") != "" {
			pvcClaim.Spec.StorageClassName = pvc.StorageClass
			if ptr.Deref(pvc.StorageClass, "") == constants.EmptyStorageClass {
				pvcClaim.Spec.StorageClassName = ptr.To("")
			}
		}

		accessModes := make([]corev1.PersistentVolumeAccessMode, 0)
		if ptr.Deref(pvc.ReadOnlyMany, false) {
			accessModes = append(accessModes, corev1.ReadOnlyMany)
		}
		if ptr.Deref(pvc.ReadWriteMany, false) {
			accessModes = append(accessModes, corev1.ReadWriteMany)
		}

		if len(accessModes) == 0 && ptr.Deref(pvc.ReadWriteOnce, true) {
			// if no access mode was specified, we default to ReadWriteOnce
			accessModes = append(accessModes, corev1.ReadWriteOnce)
		} else if len(accessModes) != 0 && ptr.Deref(pvc.ReadWriteOnce, false) {
			// if ReadWriteOnce was specified, we add it to the list of access modes
			accessModes = append(accessModes, corev1.ReadWriteOnce)
		}

		if len(accessModes) != 0 {
			pvcClaim.Spec.AccessModes = accessModes
		}

		if ptr.Deref(pvc.VolumeMode, "") != "" {
			volumeMode := corev1.PersistentVolumeMode(*pvc.VolumeMode)
			pvcClaim.Spec.VolumeMode = &volumeMode
		}
		if owner != nil {
			err = controllerutils.SetControllerReference(owner, &pvcClaim, pb.k8sClient.Scheme())
			if err != nil {
				return nil, fmt.Errorf("failed to set controller reference for PVC %s: %v", claimName, err)
			}
		}

		result = append(result, pvcClaim)

		idx += 1
		idxForPvcClaimName += 1
	}

	return result, nil
}

// GetPvcClaimName returns the claim name for the PVC
// it should be either "pvc-workloadBaseName-nameSuffix" or "pvc-workloadBaseName-idx-nameSuffix"
func GetPvcClaimName(pvc *flds.PvcsItemValue, workloadBaseName string, nameSuffix string, idx int) string {
	claimName := ptr.Deref(pvc.ClaimName, "")
	if claimName != "" {
		// claim name is set, use it
		return claimName
	}

	suffix := fmt.Sprintf("%d", idx) // default suffix is idx
	if nameSuffix != "" {
		// nameSuffix is given, the result suffix should be nameSuffix-idx
		suffix = fmt.Sprintf("%s-%d", nameSuffix, idx)
	}

	claimName = fmt.Sprintf("pvc-%s-%s", workloadBaseName, suffix)

	return claimName
}

// skipPvcBuilding checks if we should skip creating a PVC, if the PVC is ephemeral or already exists.
// It returns two booleans: the first indicates if we should skip building the PVC, and the second indicates if we should promote the index for the claim name.
// This is required in order for the claim names to be consistent with the ones used in the workload volumes.
func (pb *PvcBuilder) skipPvcBuilding(pvc *flds.PvcsItemValue) (bool, bool) {
	// if Ephemeral is set to true, we will create an ephemeral volume instead of a PVC
	if pvc == nil || ptr.Deref(pvc.Ephemeral, false) {
		return true, false
	}
	// if ExistingPvc is set to true, we will use an existing PVC instead of creating a new one
	// we does not check if the PVC exists here, because we will check it later when create
	if ptr.Deref(pvc.ExistingPvc, false) {
		return true, true
	}

	return false, false
}

// IsValidPvc checks if the PVC exists and if it is owned by the workload
func (pb *PvcBuilder) IsValidPvc(ctx context.Context, namespace string, claimName string, owner client.Object, workloadName string) error {
	logger := log.FromContext(ctx)
	claimKey := client.ObjectKey{
		Name:      claimName,
		Namespace: namespace,
	}

	var existingClaim corev1.PersistentVolumeClaim
	err := pb.k8sClient.Get(ctx, claimKey, &existingClaim)
	// if the PVC does not exist, or we cannot get it
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to get PVC %s: %v", claimName, err)
		}
		logger.V(4).Info("PVC does not exist", "claim", claimKey)
		return nil
	}
	// if the PVC exists, check if it is owned by the Workload
	if owner != nil {

		if !pb.isPvcOwnedBy(existingClaim, owner, workloadName) {
			return fmt.Errorf("PVC %s is owned by another workload", claimName)
		}
		logger.V(4).Info("PVC exists and is owned by the workload", "claim", claimKey)
		return nil

	}
	// PVC exists but is not owned by the Workload
	// this should not happen, but we still check it here
	return fmt.Errorf("pvc owner is nil, but PVC %s is not owned by the workload", claimName)
}

func (pb *PvcBuilder) isPvcOwnedBy(existingClaim corev1.PersistentVolumeClaim, owner client.Object, name string) bool {
	if metav1.IsControlledBy(&existingClaim, owner) {
		return true
	}
	// for backward compatibility, we also check the label

	return pb.backwardCompatibilityIsPVCOwnedByWorkload(existingClaim, name)
}

func (pb *PvcBuilder) backwardCompatibilityIsPVCOwnedByWorkload(existingClaim corev1.PersistentVolumeClaim, name string) bool {
	if existingClaim.Labels == nil {
		return false
	}
	// ancient PVCs may not have the Workload name owner as a label, but they should have the release name
	if jobReleaseName, ok := existingClaim.Labels[constants.LabelRelease]; ok && name != "" {
		if jobReleaseName == name {
			return true
		}
	}
	return false
}
