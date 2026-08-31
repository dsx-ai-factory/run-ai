/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"context"
	"fmt"
	"strings"

	"github.com/run-ai/runai/runai-cluster/common/constants"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetOwnerInstance(ctx context.Context, ownerRef *metav1.OwnerReference, namespace string, cachedReader client.Reader) (
	*metav1.PartialObjectMetadata, error) {
	owner := &metav1.PartialObjectMetadata{}
	groupVersion := strings.Split(ownerRef.APIVersion, "/")
	if len(groupVersion) == 1 {
		groupVersion = []string{"", groupVersion[0]}
	}
	owner.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   groupVersion[0],
		Version: groupVersion[1],
		Kind:    ownerRef.Kind,
	})
	err := cachedReader.Get(ctx, types.NamespacedName{Namespace: namespace, Name: ownerRef.Name}, owner)
	if err != nil {
		return nil, err
	}
	// In some edge cases the owner is already deleted and an identical resource with the same name has been created
	if ownerRef.UID != owner.GetUID() {
		return nil, errors.NewResourceExpired(fmt.Sprintf("Owner resource with uid %s not found", ownerRef.UID))
	}

	return owner, nil
}

// GetTopResourceOwner will go up the OwnerReferences chain, to find the last known top owner.
// If found more than 1 owner reference - will return an error.
// If there was an error retrieving an owner instance,
// will return the error,
// together with the last known owner instance (that is not self).
// The user then needs to check if the error is errors.IsForbidden(err).
func GetTopResourceOwner(ctx context.Context, object client.Object, cachedReader client.Reader) (
	lastKnownOwnerInstance *metav1.PartialObjectMetadata, err error) {

	ownerRefs := object.GetOwnerReferences()
	if len(ownerRefs) == 0 {
		return nil, fmt.Errorf("object doesn't have any owners")
	}

	var nextOwner *metav1.OwnerReference
	var ownerInstance *metav1.PartialObjectMetadata
	var lastOwnerInstance *metav1.PartialObjectMetadata
	for nextOwner = &ownerRefs[0]; nextOwner != nil; {
		ownerInstance, err = GetOwnerInstance(ctx, nextOwner, object.GetNamespace(), cachedReader)
		if err != nil {
			return lastOwnerInstance, err
		}

		nextOwners := ownerInstance.GetOwnerReferences()
		if len(nextOwners) == 0 {
			return ownerInstance, nil
		} else if len(nextOwners) > 1 {
			return nil, fmt.Errorf("found %s <%s in namespace %s> with multiple owners: %v",
				nextOwner.Kind, nextOwner.Name, object.GetNamespace(), nextOwners)
		}

		lastOwnerInstance = ownerInstance
		nextOwner = &nextOwners[0]
	}

	// shouldn't get here
	return nil, fmt.Errorf("failed to find top resource owner")
}

// GetRunaiWorkloadId - calculates the workload id, if exists - then the Control-Plane workload ID;
// if not - then it's a "regular" (non-backendy) workload that was created in the cluster, use the k8s object UID.
func GetRunaiWorkloadId(uid types.UID, annotations map[string]string) string {
	workloadId, found := annotations[constants.ControlPlaneWorkloadIdAnnotationForWorkload]
	if found && workloadId != "" {
		return workloadId
	}
	return string(uid)
}

// CalculateTopOwnerWorkloadId - gets the top-owner of the object, and returns the workload id
// (if needed, calling GetRunaiWorkloadId for runai workloads).
// NOTE: when calling this func, make sure you are using a wrapper for the cached client,
// to avoid the k8s cached-client bug: where "Get" for objects we don't have the
// RBAC for are left waiting forever. Use a wrapper to avoid those issues.
// https://github.com/kubernetes-sigs/controller-runtime/issues/550
func CalculateTopOwnerWorkloadId(object client.Object, cachedReader client.Reader) (string, error) {
	if len(object.GetOwnerReferences()) == 0 {
		return string(object.GetUID()), nil
	}

	topOwner, err := GetTopResourceOwner(context.Background(), object, cachedReader)
	if err != nil {
		return "", err
	}

	return GetRunaiWorkloadId(topOwner.UID, topOwner.Annotations), nil
}

// GetTopOwnerWorkloadId - first checks if the TopOwnerUidLabel exists, and if not - calls CalculateTopOwnerWorkloadId.
// NOTE: when calling this func, make sure you are using a wrapper for the cached client,
// to avoid the k8s cached-client bug: where "Get" for objects we don't have the
// RBAC for are left waiting forever. Use a wrapper to avoid those issues.
// https://github.com/kubernetes-sigs/controller-runtime/issues/550
func GetTopOwnerWorkloadId(object client.Object, cachedReader client.Reader) (string, error) {
	topOwnerUid, found := object.GetLabels()[constants.TopOwnerWorkloadIdLabel]
	if found && topOwnerUid != "" {
		return topOwnerUid, nil
	}

	topOwnerUid, found = object.GetLabels()[constants.TopOwnerUidLabel]
	if found && topOwnerUid != "" {
		return topOwnerUid, nil
	}

	return CalculateTopOwnerWorkloadId(object, cachedReader)
}

func GetTopOwnerWorkloadIdFromLabels(labels map[string]string) (string, error) {
	topOwnerUid, found := labels[constants.TopOwnerWorkloadIdLabel]
	if found && topOwnerUid != "" {
		return topOwnerUid, nil
	}

	topOwnerUid, found = labels[constants.TopOwnerUidLabel]
	if found && topOwnerUid != "" {
		return topOwnerUid, nil
	}

	return "", fmt.Errorf("error finding top-owner-id label <%s> or <%s> on object",
		constants.TopOwnerWorkloadIdLabel, constants.TopOwnerUidLabel)
}
