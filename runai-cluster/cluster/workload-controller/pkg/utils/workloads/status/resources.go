/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package status

import (
	"fmt"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	resourceobject "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/workloads/status/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func PatchOwnedResource(updateResource client.Object, ownedResources *[]common.ObservedResource) error {
	if updateResource == nil {
		return fmt.Errorf("updateResource is nil")
	}
	if ownedResources == nil {
		ownedResources = &[]common.ObservedResource{}
	}
	kindGroup, err := resourceobject.GetObjectResource(updateResource)
	if err != nil {
		return err
	}
	ownedResourcesTmp := make([]common.ObservedResource, len(*ownedResources))
	copy(ownedResourcesTmp, *ownedResources)
	found := false
	for index := range ownedResourcesTmp {
		if ownedResourcesTmp[index].Kind.Kind == kindGroup.Kind &&
			ownedResourcesTmp[index].Name == updateResource.GetName() &&
			ownedResourcesTmp[index].Namespace == updateResource.GetNamespace() {
			ownedResourcesTmp[index].Uid = updateResource.GetUID()
			ownedResourcesTmp[index].ObservedGeneration = updateResource.GetGeneration()
			found = true
		}
	}
	if !found {
		ownedResourcesTmp = append(ownedResourcesTmp, common.ObservedResource{
			Resource: common.Resource{
				Kind:      kindGroup,
				Name:      updateResource.GetName(),
				Namespace: updateResource.GetNamespace(),
			},
			ObservedGeneration: updateResource.GetGeneration(),
			Uid:                updateResource.GetUID(),
		})
	}
	*ownedResources = ownedResourcesTmp
	return nil
}
