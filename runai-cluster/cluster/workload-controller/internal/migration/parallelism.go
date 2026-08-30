package migration

import (
	"context"
	"fmt"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/internal/sync"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Executor interface {
	Execute(context.Context) error
}

type Handler[T any] interface {
	Handle(ctx context.Context, workload T) error
}

type ParallelMigrator[T any] struct {
	migrationName       string
	k8sClient           client.Client
	nextBatchGetter     func(continueToken string, ctx context.Context) (nextBatch []T, newContinueToken string, err error)
	handler             Handler[T]
	executorCreateFn    func(int, []T, func(context.Context, T) error) Executor
	migrationMaxWorkers int
}

func NewParallelMigrator[T any](
	migrationName string,
	k8sClient client.Client,
	nextBatchGetter func(string, context.Context) ([]T, string, error),
	handler Handler[T],
	executorCreateFn func(int, []T, func(context.Context, T) error) Executor,
	migrationMaxWorkers int,
) *ParallelMigrator[T] {
	return &ParallelMigrator[T]{
		migrationName:       migrationName,
		k8sClient:           k8sClient,
		nextBatchGetter:     nextBatchGetter,
		handler:             handler,
		executorCreateFn:    executorCreateFn,
		migrationMaxWorkers: migrationMaxWorkers,
	}
}

func (pm *ParallelMigrator[T]) Migrate(ctx context.Context) error {
	var migratedWorkloadsCounter uint64
	var continueToken string
	defer func() {
		log.Info("migration completed", "total workload migrated", migratedWorkloadsCounter)
	}()
	for {
		if ctx.Err() != nil {
			return fmt.Errorf("context is cancelled; %w", ctx.Err())
		}

		nextBatch, newContinueToken, err := pm.nextBatchGetter(continueToken, ctx)
		if err != nil {
			return err
		}

		if len(nextBatch) == 0 {
			log.Info("no workloads to migrate in this batch")
		} else {
			if pm.executorCreateFn == nil {
				pm.executorCreateFn = CreateMigrateExecuteCreateFn[T]
			}
			executor := pm.executorCreateFn(pm.migrationMaxWorkers, nextBatch, pm.handler.Handle)
			err = executor.Execute(ctx)
			if err != nil {
				return fmt.Errorf("failed to execute %s migration: %w", pm.migrationName, err)
			}

			migratedWorkloadsCounter += uint64(len(nextBatch))
		}

		continueToken = newContinueToken
		if continueToken == "" {
			break
		}
	}

	return nil
}

func CreateMigrateExecuteCreateFn[T any](numberOfWorkers int, objects []T, handler func(context.Context, T) error) Executor {
	return sync.NewWorkerPool[T](numberOfWorkers, handler, objects)
}
