/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObjectsToReconcile struct {
	ToAdd    []client.Object
	ToRemove []client.Object
	ToUpdate []client.Object
}

func InitEmptyObjectsToReconcile() *ObjectsToReconcile {
	return &ObjectsToReconcile{
		ToAdd:    []client.Object{},
		ToRemove: []client.Object{},
		ToUpdate: []client.Object{},
	}
}

func getObjNameStr(name string, namespace string) string {
	if namespace != "" {
		return fmt.Sprintf("<%s/%s>", namespace, name)
	}
	return name
}

func ObjectLogStr(object client.Object) string {
	gvk := object.GetObjectKind().GroupVersionKind()
	return fmt.Sprintf("%s/%s %s",
		gvk.GroupVersion().String(), gvk.Kind, getObjNameStr(object.GetName(), object.GetNamespace()))
}

func ObjectLogStrWithOwner(object client.Object) string {
	ownerRefs := object.GetOwnerReferences()
	if len(ownerRefs) == 0 {
		return ObjectLogStr(object)
	}
	ownerRef := ownerRefs[0]

	return fmt.Sprintf("<%s> with owner <%s/%s/%s>", ObjectLogStr(object),
		ownerRef.Kind, ownerRef.APIVersion, ownerRef.Name)
}
