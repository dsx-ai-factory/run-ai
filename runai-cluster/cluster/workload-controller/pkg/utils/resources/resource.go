/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	UnknownKind = "Unknown"
)

func GetResourceType(resource client.Object) string {
	return reflect.TypeOf(resource).Elem().Name()
}

func SafeGetKind(object client.Object) string {
	if object != nil && object.GetObjectKind() != nil {
		return object.GetObjectKind().GroupVersionKind().Kind
	}

	return UnknownKind
}
