/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PvcHandler struct {
	client.Client
}

func NewPvcHandler(client client.Client) *PvcHandler {
	return &PvcHandler{
		Client: client,
	}
}

// ConstructCopiedPvcInNamespace constructs a new PersistentVolumeClaim object in the target namespace;
// the name should already be validated to be unique.
func (pvcHandler *PvcHandler) ConstructCopiedPvcInNamespace(
	originalPvc *v1.PersistentVolumeClaim, datavolume *v1alpha1.DataVolume,
	pvcName, pvcNamespace, pvName string) *v1.PersistentVolumeClaim {
	newPvc := &v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: pvcNamespace,
			Labels: map[string]string{
				v1alpha1.DataVolumeIdLabel:   datavolume.Spec.DatavolumeId,
				v1alpha1.DataVolumeNameLabel: datavolume.Name,
			},
		},
	}
	datavolume.SetOwner(newPvc)

	pvcHandler.deepCopySpec(originalPvc, newPvc)

	newPvc.Spec.VolumeName = pvName
	return newPvc
}

func (pvcHandler *PvcHandler) deepCopySpec(fromPvc, toPvc *v1.PersistentVolumeClaim) {
	fromPvc.Spec.DeepCopyInto(&toPvc.Spec)
}

// GetPvcsOfDatavolume -
// if excludeNs is not nil, it will exclude pvcs in that namespace.
// returns a map of pvName -> pvc,
// and a list of pvcs that were somehow found without a volumeName.
func (pvcHandler *PvcHandler) GetPvcsOfDatavolume(ctx context.Context, datavolume *v1alpha1.DataVolume, excludeNs *string) (
	map[string]*v1.PersistentVolumeClaim, []*v1.PersistentVolumeClaim, error) {
	pvcs, err := pvcHandler.ListPvcsOfDatavolume(ctx, datavolume)
	if err != nil {
		return nil, nil, err
	}

	pvNameToPvc := map[string]*v1.PersistentVolumeClaim{}
	unClaimedPvcs := []*v1.PersistentVolumeClaim{}

	for i := range pvcs {
		pvc := &pvcs[i]

		if datavolume.IsOwner(pvc) == -1 {
			continue
		}
		if excludeNs != nil && pvc.Namespace == *excludeNs {
			continue
		}

		pvName := pvc.Spec.VolumeName
		if pvName == "" {
			unClaimedPvcs = append(unClaimedPvcs, pvc)
			continue
		}
		pvNameToPvc[pvName] = pvc
	}

	return pvNameToPvc, unClaimedPvcs, nil

}

func (pvcHandler *PvcHandler) GetPvc(ctx context.Context, name, namespace string) (*v1.PersistentVolumeClaim, error) {
	pvc := &v1.PersistentVolumeClaim{}
	err := pvcHandler.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}, pvc)
	if err != nil {
		return nil, err
	}

	return pvc, nil
}

func (pvcHandler *PvcHandler) ListPvcsOfDatavolume(ctx context.Context, datavolume *v1alpha1.DataVolume) ([]v1.PersistentVolumeClaim, error) {
	return pvcHandler.listPvcs(ctx, client.MatchingLabels{
		v1alpha1.DataVolumeIdLabel: datavolume.Spec.DatavolumeId,
	})
}

func (pvcHandler *PvcHandler) ListPvcsInNamespace(ctx context.Context, namespace string, opts ...client.ListOption) ([]v1.PersistentVolumeClaim, error) {
	opts = append(opts, client.InNamespace(namespace))
	return pvcHandler.listPvcs(ctx, opts...)
}

func (pvcHandler *PvcHandler) ListPvcsOfDatavolumeInNamespace(ctx context.Context, namespace string, datavolumeId string) ([]v1.PersistentVolumeClaim, error) {
	return pvcHandler.ListPvcsInNamespace(ctx, namespace, client.MatchingLabels{
		v1alpha1.DataVolumeIdLabel: datavolumeId,
	})
}

func (pvcHandler *PvcHandler) listPvcs(ctx context.Context, opts ...client.ListOption) ([]v1.PersistentVolumeClaim, error) {
	pvs := &v1.PersistentVolumeClaimList{}
	err := pvcHandler.List(ctx, pvs, opts...)
	if err != nil {
		return nil, err
	}

	return pvs.Items, nil
}
