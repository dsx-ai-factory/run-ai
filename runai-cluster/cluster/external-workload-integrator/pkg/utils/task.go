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
	// CreateFromRunnerOperation builds an ExternalWorkload from the runner alone, for a
	// workload object that was never created
	CreateFromRunnerOperation TaskOperation = "createFromRunner"
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

	// set when the task was enqueued for a known ExternalWorkload, so it can still be found
	// when the workload object it describes is gone
	ExternalWorkloadName string

	// set when the task is keyed on the runner rather than the workload object
	WorkloadRunnerName string
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

func (ew *ExternalWorkloadTask) SetExternalWorkloadName(ewName string) *ExternalWorkloadTask {
	ew.ExternalWorkloadName = ewName
	return ew
}

func (ew *ExternalWorkloadTask) SetWorkloadRunnerName(name string) *ExternalWorkloadTask {
	ew.WorkloadRunnerName = name
	return ew
}

func (ew *ExternalWorkloadTask) GetOperation() string {
	return string(ew.operation)
}
