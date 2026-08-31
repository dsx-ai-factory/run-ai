/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// This file is a workaround for a bug in vite.
// The bug:
// When you import the WORKSPACE_ROUTE_NAMES from the workspace.route.ts file into
// component, during the work you do some changes to the component (or to his parents),
// HMR (used by vite) show an exception about "initialization" and "syntax errors or importing non-existent modules".
// So I moved the names out of the route file, which fix this error.
export const TRAINING_ROUTE_NAMES = {
  TRAINING_NEW: "training-new",
  TRAINING_ASSETS_EDIT: "training-assets-edit",
  DISTRIBUTED_ASSETS_EDIT: "distributed-assets-edit",
  TRAINING_HYBRID: "training-hybrid",
  DISTRIBUTED_HYBRID: "distributed-hybrid",
};
