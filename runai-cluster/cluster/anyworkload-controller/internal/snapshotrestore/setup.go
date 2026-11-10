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

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// NewSetup returns a manager setup that serves the restore webhook. Registered
// only when Snapshot is installed, on the same fact that gates the webhook
// configuration, so the path and its routes appear together.
func NewSetup() func(context.Context, ctrl.Manager) error {
	return func(_ context.Context, manager ctrl.Manager) error {
		manager.GetWebhookServer().Register(constants.SnapshotRestoreWebhookPath, &webhook.Admission{Handler: NewMutator()})
		return nil
	}
}
