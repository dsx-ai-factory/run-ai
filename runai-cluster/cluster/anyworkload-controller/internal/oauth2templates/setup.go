/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package oauth2templates

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
)

// NewSetup registers template provisioning independently of optional workload plugins.
func NewSetup() func(context.Context, ctrl.Manager) error {
	return func(_ context.Context, mgr ctrl.Manager) error {
		if err := NewReconciler(mgr.GetClient()).SetupWithManager(mgr); err != nil {
			return fmt.Errorf("setup OAuth2 proxy templates reconciler: %w", err)
		}
		return nil
	}
}
