/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package externalworkloads

import (
	"context"
	"errors"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/cachedclient"
	topowner "github.com/dsx-ai-factory/run-ai/runai-cluster/common/top-owner"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// errDroppedByLabel is returned when the object, or an owner in its chain, carries a label that
// cachedclient.ShouldDropByLabel matches.
var errDroppedByLabel = errors.New("object belongs to a workload that must not get an externalworkload of its own")

func isDroppedByLabel(err error) bool {
	return errors.Is(err, errDroppedByLabel)
}

// getTopResourceOwnerWithHandler will go up the OwnerReferences chain,
// to find the last known top owner + the last top owner that has a handler.
// If found more than 1 owner reference - will return an error.
// If the object or an owner in the chain carries a label that must be dropped (see cachedclient.ShouldDropByLabel) -
// will return errDroppedByLabel.
// If there was an error retrieving an owner instance,
// will return the error,
// together with the last known owner instance (that is not self).
// The user then needs to check if the error is errors.IsForbidden(err).
// There could be a case where not all owners in the chain have handlers, but the last one has - great, we will return it.
// If the last known doesn't have a handler - then we will return it in the first return value,
// and the last top owner that has a handler (could be nil) in the second return value.
func (h *ExternalWorkloadHandler) getTopResourceOwnerWithHandler(ctx context.Context, object client.Object) (*metav1.PartialObjectMetadata, *metav1.PartialObjectMetadata, error) {
	if cachedclient.ShouldDropByLabel(object) {
		return nil, nil, errDroppedByLabel
	}
	ownerRefs := object.GetOwnerReferences()
	if len(ownerRefs) == 0 {
		return nil, nil, fmt.Errorf("object doesn't have any owners")
	}
	if len(ownerRefs) > 1 {
		return nil, nil, fmt.Errorf("object has multiple owners")
	}

	var nextOwner *metav1.OwnerReference
	var lastOwnerInstance *metav1.PartialObjectMetadata
	var lastOwnerInstanceWithHandler *metav1.PartialObjectMetadata
	for nextOwner = &ownerRefs[0]; nextOwner != nil; {
		ownerInstance, err := topowner.GetOwnerInstance(ctx, nextOwner, object.GetNamespace(), h.cachedClient)
		if err != nil {
			return lastOwnerInstance, lastOwnerInstanceWithHandler, err
		}
		if cachedclient.ShouldDropByLabel(ownerInstance) {
			return nil, nil, errDroppedByLabel
		}

		nextOwners := ownerInstance.GetOwnerReferences()
		if len(nextOwners) == 0 {
			// reached top owner - check if it has a handler
			if _, found := h.supportedGVKsHandlers[ownerInstance.GroupVersionKind()]; found {
				return ownerInstance, ownerInstance, nil
			}
			// if it doesn't have handler - return last known (ownerInstance) and last known with handler (could be nil)
			return ownerInstance, lastOwnerInstanceWithHandler, nil
		} else if len(nextOwners) > 1 {
			return nil, nil, fmt.Errorf("found %s <%s in namespace %s> with multiple owners: %v",
				nextOwner.Kind, nextOwner.Name, object.GetNamespace(), nextOwners)
		}

		lastOwnerInstance = ownerInstance
		if _, found := h.supportedGVKsHandlers[ownerInstance.GroupVersionKind()]; found {
			lastOwnerInstanceWithHandler = ownerInstance
		}

		nextOwner = &nextOwners[0]
	}

	// shouldn't get here
	return nil, nil, fmt.Errorf("failed to find top resource owner")
}
