/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func IsObjectOwnedBy(object client.Object, ownerUid types.UID) bool {
	if object == nil {
		return false
	}
	for _, owner := range object.GetOwnerReferences() {
		if owner.UID == ownerUid {
			return true
		}
	}
	return false
}

// this is the function check if the object is created by the owner
// by checking the label release, if the label release is the same as the owner name, then it is created by the owner
// this method is ony used for backward compatibility, and should be removed in the future
func BackwardCompatibilityIsObjectCreatedBy(object client.Object, ownerName string) bool {
	if object.GetOwnerReferences() != nil || len(object.GetOwnerReferences()) != 0 {
		return false
	}
	if object.GetLabels() == nil {
		return false
	}
	if release, ok := object.GetLabels()[constants.LabelRelease]; ok && release == ownerName {
		return true
	}
	return false
}
