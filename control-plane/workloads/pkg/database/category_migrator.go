/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package database

import (
	"context"
	"fmt"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models" // Added for models.Workload
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"github.com/google/uuid"
	_ "github.com/lib/pq" // Blank import for PostgreSQL driver registration

	"gorm.io/gorm"
)

const (
	defaultBatchSize = 200
	defaultSleep     = time.Millisecond * 20
	// RunTimeout The timeout of 30 minutes is to allow deletion on up to ~2.5M rows
	RunTimeout = 30 * time.Minute
)

var nativeWorkloadTypes = map[string]string{
	"Training":    "Train",
	"Inference":   "Deploy",
	"Workspace":   "Build",
	"Distributed": "Train",
}

// CategoryMigrator handles the batch update of category_id in the workloads table.
type CategoryMigrator struct {
	client        orm.DB
	batchSize     int
	sleepInterval time.Duration
	stopCh        <-chan struct{}
}

// NewCategoryMigrator creates a new CategoryMigrator instance.
func NewCategoryMigrator(client orm.DB, batchSize int, sleepInterval time.Duration, stopCh <-chan struct{}) *CategoryMigrator {
	return &CategoryMigrator{
		client:        client,
		batchSize:     batchSize,
		sleepInterval: sleepInterval,
		stopCh:        stopCh,
	}
}

// migrateExternalWorkloadsBatch processes a single batch of workloads and updates their category_id.
// Returns the number of workloads updated in this batch and any error that occurred.
func (m *CategoryMigrator) migrateExternalWorkloadsBatch(ctx context.Context) (int64, error) {
	tx := m.client.WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	var currentBatchIDs []string
	err := tx.Table("workloads").
		Select("id").
		Where("category_id IS NULL").
		Limit(m.batchSize).
		Find(&currentBatchIDs)
	if err != nil {
		logging.Errorf("CATEGORY_MIGRATOR: GORM failed to select batch of workload IDs: %v", err)
		return 0, err
	}

	if len(currentBatchIDs) == 0 {
		logging.Infof("CATEGORY_MIGRATOR: No more workloads to update.")
		return 0, nil
	}

	subQuery := tx.Table(models.ExternalWorkloadTypeConfigView{}.TableName()).
		Select("category_id").
		Where("name = workloads.type").
		Limit(1)

	currentBatchUpdated, err := tx.Model(&models.Workload{}).
		Where("id IN ?", currentBatchIDs).
		Update("category_id", subQuery.Actual())
	if err != nil {
		logging.Errorf("CATEGORY_MIGRATOR: GORM failed to update workloads: %v", err)
		return 0, err
	}

	if txErr := tx.Commit(); txErr != nil {
		logging.Errorf("CATEGORY_MIGRATOR: GORM failed to commit transaction: %v", txErr)
		return 0, txErr
	}
	committedTx = true

	if currentBatchUpdated < int64(len(currentBatchIDs)) {
		logging.Warnf("CATEGORY_MIGRATOR: Note: Expected to update %d workloads, but only updated %d.", len(currentBatchIDs), currentBatchUpdated)
	}

	return currentBatchUpdated, nil
}

// migrateNativeWorkloadsBatch processes a single batch of native workloads and updates their category_id.
// Returns the number of workloads updated in this batch and any error that occurred.
func (m *CategoryMigrator) migrateNativeWorkloadsBatch(ctx context.Context, categoryIDs map[string]uuid.UUID) (int64, error) {
	tx := m.client.WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	var currentBatchIDs []string
	err := tx.Model(&models.Workload{}).
		Select("id").
		Where("category_id IS NULL AND type IN ?", getMapKeys(nativeWorkloadTypes)).
		Limit(m.batchSize).
		Find(&currentBatchIDs)
	if err != nil {
		logging.Errorf("CATEGORY_MIGRATOR: GORM failed to select batch of workload IDs: %v", err)
		return 0, err
	}

	if len(currentBatchIDs) == 0 {
		logging.Infof("CATEGORY_MIGRATOR: No more native workloads to update.")
		return 0, nil
	}

	// Create a case statement for the update
	caseStmt := "CASE type "
	for workloadType, categoryName := range nativeWorkloadTypes {
		if categoryID, ok := categoryIDs[categoryName]; ok {
			caseStmt += fmt.Sprintf("WHEN '%s' THEN '%s'::uuid ", workloadType, categoryID)
		}
	}
	caseStmt += "END"

	currentBatchUpdated, err := tx.Model(&models.Workload{}).
		Where("id IN ?", currentBatchIDs).
		Update("category_id", gorm.Expr(caseStmt))
	if err != nil {
		logging.Errorf("CATEGORY_MIGRATOR: GORM failed to update native workloads: %v", err)
		return 0, err
	}

	if txErr := tx.Commit(); txErr != nil {
		logging.Errorf("CATEGORY_MIGRATOR: GORM failed to commit transaction: %v", txErr)
		return 0, txErr
	}
	committedTx = true

	if currentBatchUpdated < int64(len(currentBatchIDs)) {
		logging.Warnf("CATEGORY_MIGRATOR: Note: Expected to update %d native workloads, but only updated %d.", len(currentBatchIDs), currentBatchUpdated)
	}

	return currentBatchUpdated, nil
}

// getMapKeys returns a slice of keys from a map
func getMapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// getMapValues returns a slice of values from a map
func getMapValues[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

// handleShutdownSignals handles context cancellation and stop signals.
// Returns true if should continue processing, false if should stop.
func (m *CategoryMigrator) handleShutdownSignals(ctx context.Context) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	case <-m.stopCh:
		return false, nil
	case <-time.After(m.sleepInterval):
		return true, nil
	}
}

func (m *CategoryMigrator) migrateExternalWorkloads(ctx context.Context) error {
	var totalUpdatedOverall int64 = 0

	for {
		shouldContinue, err := m.handleShutdownSignals(ctx)
		if !shouldContinue {
			return err
		}

		currentBatchUpdated, _ := m.migrateExternalWorkloadsBatch(ctx)
		if currentBatchUpdated == 0 {
			logging.Infof("CATEGORY_MIGRATOR: No more workloads to update. Migration complete. Total updated: %d", totalUpdatedOverall)
			return nil
		}

		totalUpdatedOverall += currentBatchUpdated

		shouldContinue, err = m.handleShutdownSignals(ctx)
		if !shouldContinue {
			return err
		}
	}
}

// migrateNativeWorkloads processes all native workloads in batches.
func (m *CategoryMigrator) migrateNativeWorkloads(ctx context.Context) error {
	// Prefetch category IDs
	var categoryIDs = make(map[string]uuid.UUID)
	var categories []models.WorkloadsCategory
	if err := m.client.Model(&models.WorkloadsCategory{}).
		Where("name IN ?", getMapValues(nativeWorkloadTypes)).
		Find(&categories); err != nil {
		logging.Errorf("CATEGORY_MIGRATOR: Failed to fetch category IDs: %v", err)
		return err
	}

	for _, cat := range categories {
		categoryIDs[cat.Name] = cat.ID
	}

	var totalUpdatedOverall int64 = 0

	for {
		shouldContinue, err := m.handleShutdownSignals(ctx)
		if !shouldContinue {
			return err
		}

		currentBatchUpdated, _ := m.migrateNativeWorkloadsBatch(ctx, categoryIDs)
		if currentBatchUpdated == 0 {
			logging.Infof("CATEGORY_MIGRATOR: No more native workloads to update. Migration complete. Total updated: %d", totalUpdatedOverall)
			return nil
		}

		totalUpdatedOverall += currentBatchUpdated

		shouldContinue, err = m.handleShutdownSignals(ctx)
		if !shouldContinue {
			return err
		}
	}
}

// Run executes the category migration process.
// It processes workloads in batches, updating their category_id based on external_workloads_type_config.
func (m *CategoryMigrator) Run(ctx context.Context) error {
	logging.Infof("CATEGORY_MIGRATOR: Starting category migration with GORM, batch size %d and sleep interval %v", m.batchSize, m.sleepInterval)
	var totalUpdatedOverall int64 = 0
	var errs []error

	if err := m.migrateNativeWorkloads(ctx); err != nil {
		errs = append(errs, fmt.Errorf("native workloads migration failed: %w", err))
	}

	if err := m.migrateExternalWorkloads(ctx); err != nil {
		errs = append(errs, fmt.Errorf("external workloads migration failed: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("migration errors: %v", errs)
	}
	logging.Infof("CATEGORY_MIGRATOR: Finished adding category_id to all workloads. Total updated: %d, batch size %d and sleep interval %v", totalUpdatedOverall, m.batchSize, m.sleepInterval)

	return nil
}

func InitCategoryMigrator(stopCh <-chan struct{}) {
	ctx, cancel := context.WithTimeout(context.Background(), RunTimeout)
	go func() {
		defer cancel()
		client := runaidb.DB()

		migrator := NewCategoryMigrator(client, defaultBatchSize, defaultSleep, stopCh)
		err := migrator.Run(ctx)
		if err != nil {
			logging.Errorf("error running InitCategoryMigrator %v", err.Error())
		}
	}()
}
