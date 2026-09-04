/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package karta

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Retry budget for EWI to reconcile a freshly-created Karta before we declare
// the migration broken. 15 attempts x 2s gives EWI ~30s to catch up.
//
// verifierWorkers caps concurrent per-Karta checks so the pre-upgrade hook
// finishes well within the 600s helm activeDeadlineSeconds even with
// hundreds of RIs (per Isan-Rivkin PR review):
//
//	worst case = ceil(N / verifierWorkers) * verifierRetryAttempts * verifierRetryDelay
//
// At N=36 sequential the worst case was ~18min; with verifierWorkers=10 the
// same N is ~120s, and 500 RIs is still ~25min worst-case but only seconds
// in the realistic path where most Kartas are already covered.
const (
	verifierRetryAttempts = 15
	verifierRetryDelay    = 2 * time.Second
	verifierWorkers       = 10
)

// VerifyHealth checks that every Karta covers its RI (every True condition
// on the RI is also True on the Karta). Strict gate for the pre-upgrade
// hook. Per-RI checks run in parallel (up to verifierWorkers at a time) so
// the gate stays under the 600s helm hook deadline.
func VerifyHealth(ctx context.Context, c client.Client) error {
	ris, err := listRIs(ctx, c)
	if err != nil {
		return err
	}

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(verifierWorkers)
	for i := range ris.Items {
		ri := &ris.Items[i]
		g.Go(func() error {
			return verifyOne(gctx, c, ri)
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}
	log.Info().Int("count", len(ris.Items)).Msg("karta health verified")
	return nil
}

// verifyOne polls a single Karta until it covers its source RI, or until
// the retry budget is exhausted.
func verifyOne(ctx context.Context, c client.Client, ri *unstructured.Unstructured) error {
	name, err := generateKartaName(ri)
	if err != nil {
		return err
	}
	var lastWhy string
	for attempt := 1; attempt <= verifierRetryAttempts; attempt++ {
		karta, err := getKarta(ctx, c, name)
		if err != nil {
			return fmt.Errorf("get karta %s: %w", name, err)
		}
		if ok, why := kartaCoversRI(karta, ri); ok {
			return nil
		} else {
			lastWhy = why
		}
		log.Debug().Str("karta", name).Int("attempt", attempt).Str("reason", lastWhy).Msg("karta not yet covering RI, retrying")
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(verifierRetryDelay):
		}
	}
	return fmt.Errorf("karta %s does not cover RI %s after %d attempts: %s", name, ri.GetName(), verifierRetryAttempts, lastWhy)
}
