/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"context"
	"errors"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const strictOwnerChainMaxDepth = 16

// ErrInvalidOwnerChain marks an ownership graph that cannot prove the expected
// owner. Kubernetes read failures are returned without this marker.
var ErrInvalidOwnerChain = errors.New("invalid owner chain")

// StrictOwner identifies the exact object at which an ownership walk must end.
type StrictOwner struct {
	GroupKind schema.GroupKind
	Namespace string
	Name      string
	UID       types.UID
}

// VerifyStrictOwnerChain follows live owner references from object to expected.
// It verifies every intermediate UID and rejects ambiguous, cyclic, or broken
// chains. A controller owner is authoritative when present; otherwise the
// object must have exactly one owner reference. The supplied reader must perform
// live, uncached API reads when this function is used for authorization.
func VerifyStrictOwnerChain(
	ctx context.Context,
	object client.Object,
	expected StrictOwner,
	reader client.Reader,
) error {
	if object == nil {
		return invalidOwnerChainf("start object is nil")
	}
	if reader == nil {
		return fmt.Errorf("strict owner reader is nil")
	}
	if object.GetNamespace() == "" || object.GetName() == "" || object.GetUID() == "" {
		return invalidOwnerChainf("start object identity is incomplete")
	}
	if expected.GroupKind.Empty() || expected.Namespace == "" || expected.Name == "" || expected.UID == "" {
		return invalidOwnerChainf("expected owner identity is incomplete")
	}
	if object.GetNamespace() != expected.Namespace {
		return invalidOwnerChainf(
			"object namespace %q does not match expected owner namespace %q",
			object.GetNamespace(), expected.Namespace,
		)
	}

	seen := datastructs.NewSet(identityForObject(object))
	current := object
	for range strictOwnerChainMaxDepth {
		ownerRef, err := authoritativeOwnerReference(current.GetOwnerReferences())
		if err != nil {
			return err
		}
		ownerGV, err := schema.ParseGroupVersion(ownerRef.APIVersion)
		if err != nil {
			return invalidOwnerChainf("owner reference API version %q is invalid: %v", ownerRef.APIVersion, err)
		}
		ownerGVK := ownerGV.WithKind(ownerRef.Kind)

		if ownerGVK.GroupKind() == expected.GroupKind && ownerRef.Name == expected.Name {
			if ownerRef.UID != expected.UID {
				return invalidOwnerChainf(
					"owner reference %s/%s has UID %q, expected %q",
					ownerRef.Kind, ownerRef.Name, ownerRef.UID, expected.UID,
				)
			}
			return nil
		}

		identity := ownerIdentity{Name: ownerRef.Name, UID: ownerRef.UID}
		if seen.Has(identity) {
			return invalidOwnerChainf(
				"cycle detected at %s/%s with UID %q",
				ownerRef.Kind, ownerRef.Name, ownerRef.UID,
			)
		}
		seen.Add(identity)

		owner := &metav1.PartialObjectMetadata{}
		owner.SetGroupVersionKind(ownerGVK)
		if err := reader.Get(
			ctx,
			client.ObjectKey{Namespace: expected.Namespace, Name: ownerRef.Name},
			owner,
		); err != nil {
			return fmt.Errorf("read owner %s/%s: %w", ownerRef.Kind, ownerRef.Name, err)
		}
		if owner.GetUID() != ownerRef.UID {
			return invalidOwnerChainf(
				"live owner %s/%s has UID %q, owner reference has %q",
				ownerRef.Kind, ownerRef.Name, owner.GetUID(), ownerRef.UID,
			)
		}
		current = owner
	}

	return invalidOwnerChainf("owner chain exceeds maximum depth %d", strictOwnerChainMaxDepth)
}

func authoritativeOwnerReference(ownerRefs []metav1.OwnerReference) (metav1.OwnerReference, error) {
	controllers := make([]metav1.OwnerReference, 0, 1)
	for _, ownerRef := range ownerRefs {
		if ownerRef.Controller != nil && *ownerRef.Controller {
			controllers = append(controllers, ownerRef)
		}
	}

	var ownerRef metav1.OwnerReference
	if len(controllers) == 1 {
		ownerRef = controllers[0]
	} else if len(controllers) > 1 {
		return metav1.OwnerReference{}, invalidOwnerChainf("object has %d controller owners", len(controllers))
	} else if len(ownerRefs) == 1 {
		ownerRef = ownerRefs[0]
	} else {
		return metav1.OwnerReference{}, invalidOwnerChainf(
			"object has %d owner references and no authoritative controller owner",
			len(ownerRefs),
		)
	}

	if ownerRef.APIVersion == "" || ownerRef.Kind == "" || ownerRef.Name == "" || ownerRef.UID == "" {
		return metav1.OwnerReference{}, invalidOwnerChainf("owner reference is incomplete: %v", ownerRef)
	}
	return ownerRef, nil
}

type ownerIdentity struct {
	Name string
	UID  types.UID
}

func identityForObject(object client.Object) ownerIdentity {
	return ownerIdentity{
		Name: object.GetName(),
		UID:  object.GetUID(),
	}
}

func invalidOwnerChainf(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidOwnerChain, fmt.Sprintf(format, args...))
}
