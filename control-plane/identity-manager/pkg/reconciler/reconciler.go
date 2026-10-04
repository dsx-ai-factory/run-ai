/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package reconciler

import (
	"context"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

type Reconciler interface {
	// Reconcile reconciles the resource.
	Reconcile(ctx context.Context, realm string, force bool) error
}

type ReconcileManager interface {
	ShouldReconcile() bool
	ShouldUpdateTenants(ctx context.Context) bool
	RunReconcile(ctx context.Context) error
	AddAllTenantsToLocalSchema(ctx context.Context)
}

type Key string

const (
	LogKey Key = "log"
)

func LogReconcileDuration(ctx context.Context, start time.Time, realmName, reconcileName string) {
	durationTime := time.Since(start)
	logObject := map[string]string{
		NameLabel:      reconcileName,
		RealmNameLabel: realmName,
	}
	metrics.GaugeVec(MetricName).Get().WithLabels(logObject).Set(durationTime.Seconds())
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("reconciling duration for %s in realm: %s took: %s\n", reconcileName, realmName, durationTime)
}

type ReconcilerConfig struct {
	InternalServiceAccountsVersion string
	ServiceAccountsVersion         string
	AccessKeysVersion              string
}

func NewReconcilerConfig(internalServiceAccountsVersion, serviceAccountsVersion, accessKeysVersion string) *ReconcilerConfig {
	return &ReconcilerConfig{
		InternalServiceAccountsVersion: internalServiceAccountsVersion,
		ServiceAccountsVersion:         serviceAccountsVersion,
		AccessKeysVersion:              accessKeysVersion,
	}
}
