/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package constants

// NVIDIA Snapshot labels. They live here because the go-operator writes them
// into the restore webhook's objectSelector while the anyworkload-controller
// matches on them, and the two are separate modules. If the selector and the
// matcher ever disagreed, opted-in pods would silently stop being routed to
// the webhook.
const (
	// SnapshotOptInLabel is user-set on the pod template, and is the restore
	// webhook's objectSelector. It is a label rather than an annotation for
	// exactly that reason: an objectSelector cannot match annotations.
	SnapshotOptInLabel = "run.ai/snapshot"

	// SnapshotKeyLabel is stamped by the restore webhook on every opted-in
	// pod, and is how the capture plugin finds each key's first Ready replica.
	SnapshotKeyLabel = "run.ai/snapshot-key"

	// SnapshotRestoreWebhookPath is where the restore webhook is served. The
	// go-operator writes it into the webhook configuration's clientConfig and
	// the anyworkload-controller registers it on its webhook server; if they
	// disagreed the API server would call a path nothing answers.
	SnapshotRestoreWebhookPath = "/mutate-pod-snapshot-restore"
)
