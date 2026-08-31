/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package events

import (
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WorkloadRecorder struct {
	workloadObj   client.Object
	eventRecorder record.EventRecorder
}

func NewWorkloadRecorder(workloadObj client.Object, eventRecorder record.EventRecorder) *WorkloadRecorder {
	return &WorkloadRecorder{
		workloadObj:   workloadObj,
		eventRecorder: eventRecorder,
	}
}

func (r *WorkloadRecorder) Event(eventtype, reason, message string) {
	r.eventRecorder.Event(r.workloadObj, eventtype, reason, message)
}

func (r *WorkloadRecorder) Eventf(eventtype, reason, messageFmt string, args ...interface{}) {
	r.eventRecorder.Eventf(r.workloadObj, eventtype, reason, messageFmt, args...)
}
