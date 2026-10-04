/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package reconciler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

const (
	NameLabel      = "reconcilerName"
	RealmNameLabel = "realmName"
	MetricName     = "reconcile_duration"
)

type MasterReconcilerI interface {
	AddReconciler(reconcilerType api.ReconcilerType, reconciler Reconciler)
	GetReconcilerByType(recType api.ReconcilerType) Reconciler
	Reconcile(ctx context.Context, realm string, force bool) error
	ReconcileRealms(ctx context.Context, realms *[]models.Realm, ignoredRealms map[string]bool, force bool) error
}

type MasterReconciler struct {
	reconcilers map[api.ReconcilerType]Reconciler
}

func NewMasterReconciler(ctx context.Context) MasterReconcilerI {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Init Master Reconciler")
	logObject := []string{NameLabel, RealmNameLabel}
	metrics.GaugeVec(MetricName).Register("reconciliation time in seconds", logObject)
	return &MasterReconciler{
		reconcilers: map[api.ReconcilerType]Reconciler{},
	}
}

func (r *MasterReconciler) AddReconciler(reconcilerType api.ReconcilerType, reconciler Reconciler) {
	r.reconcilers[reconcilerType] = reconciler
}

func (r *MasterReconciler) GetReconcilerByType(recType api.ReconcilerType) Reconciler {
	if _, ok := r.reconcilers[recType]; ok {
		return r.reconcilers[recType]
	}
	return nil
}

func (r *MasterReconciler) ReconcileRealms(ctx context.Context, realms *[]models.Realm, ignoredRealms map[string]bool, force bool) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Performing reconciliation for realms")
	var failedRealms []string
	start := time.Now()
	defer LogReconcileDuration(ctx, start, "all", "ReconcileRealms")
	for _, realm := range *realms {
		realmName := realm.Realm
		if _, ok := ignoredRealms[realmName]; ok {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Skipping reconciliation for %s realm", realmName)
		} else {
			err := r.Reconcile(ctx, realmName, force)
			if err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to reconcile realm %s, err: %v", realmName, err)
				failedRealms = append(failedRealms, realmName)
			}
		}
	}
	if len(failedRealms) > 0 {
		return fmt.Errorf("error on reconcile for realms: %s", failedRealms)
	}
	return nil
}

func (r *MasterReconciler) Reconcile(ctx context.Context, realm string, force bool) error {
	start := time.Now()
	defer LogReconcileDuration(ctx, start, realm, "MasterReconciler")
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Reconciling realm %s", realm)
	numWorkers := len(r.reconcilers)
	var wg sync.WaitGroup
	errorsChan := make(chan error, numWorkers)
	for reconcilerType, reconciler := range r.reconcilers {
		wg.Add(1)
		go r.runReconcilerAsync(ctx, reconcilerType, &wg, errorsChan, reconciler, realm, force)
	}
	wg.Wait()
	close(errorsChan)
	for err := range errorsChan {
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *MasterReconciler) runReconcilerAsync(ctx context.Context, reconcilerType api.ReconcilerType, wg *sync.WaitGroup, errorsChan chan error, rec Reconciler, realm string, force bool) {
	defer wg.Done()
	err := rec.Reconcile(ctx, realm, force)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("error running %s reconciler", reconcilerType)
		errorsChan <- err
	}
}
