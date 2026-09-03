/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package jobs

import (
	"context"
	workloadcommon "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func IndexInvolvedOid(mgr manager.Manager) error {
	err := mgr.GetFieldIndexer().IndexField(
		context.Background(), &corev1.Event{},
		workloadcommon.InvolvedObjectUidField,
		func(obj client.Object) []string {
			ev := obj.(*corev1.Event)
			return []string{string(ev.InvolvedObject.UID)}
		},
	)

	if err != nil {
		return err
	}

	return nil
}
