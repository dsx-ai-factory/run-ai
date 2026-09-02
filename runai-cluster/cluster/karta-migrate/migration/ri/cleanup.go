/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package ri

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/karta-migrate/migration"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Cleanup strips the legacy RI finalizer and deletes every ResourceInterface
// CR. Called by the post-upgrade hook once the migration has produced
// matching Kartas.
//
// Per natasharomm PR review: errors on a single RI do not abort the loop -
// they are accumulated into a multierror so we delete as many RIs as
// possible in one pass. The post-upgrade hook is idempotent; the next run
// retries any failures.
func Cleanup(ctx context.Context, c client.Client) error {
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(migration.LegacyRIGVK.GroupVersion().WithKind(migration.LegacyRIGVK.Kind + "List"))
	if err := c.List(ctx, list); err != nil {
		return err
	}

	var errs *multierror.Error
	for i := range list.Items {
		ri := &list.Items[i]

		finalizers := ri.GetFinalizers()
		stripped := make([]string, 0, len(finalizers))
		for _, f := range finalizers {
			if f != Finalizer {
				stripped = append(stripped, f)
			}
		}
		if len(stripped) != len(finalizers) {
			ri.SetFinalizers(stripped)
			if err := c.Update(ctx, ri); err != nil && !apierrors.IsNotFound(err) {
				errs = multierror.Append(errs, fmt.Errorf("strip finalizer from %s: %w", ri.GetName(), err))
				// keep going to delete - finalizer strip can fail on conflict
				// while the object is still deletable.
			}
		}

		if err := c.Delete(ctx, ri); err != nil && !apierrors.IsNotFound(err) {
			errs = multierror.Append(errs, fmt.Errorf("delete RI %s: %w", ri.GetName(), err))
			continue
		}
		log.Debug().Str("name", ri.GetName()).Msg("deleted RI")
	}
	return errs.ErrorOrNil()
}
