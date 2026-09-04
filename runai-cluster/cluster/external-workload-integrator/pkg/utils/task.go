/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type TaskOperation string

const (
	AddOperation    TaskOperation = "add"
	UpdateOperation TaskOperation = "update"
	DeleteOperation TaskOperation = "delete"
)

type TaskPayload[T client.Object] interface {
	GetOperation() string
}

type HandleFunc func(context.Context, client.Object, schema.GroupVersionKind, TaskOperation) error

type ExternalWorkloadTask struct {
	operation TaskOperation
	Object    client.Object
	Key       types.NamespacedName
	GVK       schema.GroupVersionKind
}

func NewExternalWorkloadTask(operation TaskOperation, key types.NamespacedName, gvk schema.GroupVersionKind) *ExternalWorkloadTask {
	task := &ExternalWorkloadTask{
		operation: operation,
		Key:       key,
		GVK:       gvk,
	}
	return task
}

func (ew *ExternalWorkloadTask) SetObject(obj client.Object) *ExternalWorkloadTask {
	ew.Object = obj
	return ew
}

func (ew *ExternalWorkloadTask) GetOperation() string {
	return string(ew.operation)
}
