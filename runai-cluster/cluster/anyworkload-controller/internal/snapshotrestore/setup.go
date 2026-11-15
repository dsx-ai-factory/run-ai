/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package snapshotrestore serves the restore webhook, which turns an opted-in
// replica into a pod restored from an NVIDIA Snapshot checkpoint. It runs at pod
// CREATE because Snapshot's NRI plugin replaces the container command at
// container creation; no later hook works.
package snapshotrestore

import (
	"context"
	"fmt"

	snapshotv1alpha1 "github.com/ai-dynamo/snapshot/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// NewSetup returns a manager setup that serves the restore webhook. Registered
// only when Snapshot is installed, on the same fact that gates the webhook
// configuration, so the path and its routes appear together.
func NewSetup(agentNodeSelector map[string]string) func(context.Context, ctrl.Manager) error {
	return func(_ context.Context, manager ctrl.Manager) error {
		// This manager only: Snapshot is optional, so its types do not belong
		// in the shared scheme.
		if err := snapshotv1alpha1.AddToScheme(manager.GetScheme()); err != nil {
			return fmt.Errorf("register Snapshot types: %w", err)
		}

		// GetAPIReader, not GetClient: the grant is get only.
		handler := NewMutator(manager.GetAPIReader(), agentNodeSelector)
		manager.GetWebhookServer().Register(constants.SnapshotRestoreWebhookPath,
			&webhook.Admission{Handler: handler})
		return nil
	}
}
