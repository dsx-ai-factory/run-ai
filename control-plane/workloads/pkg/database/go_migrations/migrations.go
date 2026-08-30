package go_migrations

import (
	"context"

	ri_migrations "github.com/run-ai/runai/control-plane/workloads/pkg/database/go_migrations/ri"
)

func Migrate() error {
	ctx := context.Background()

	err := ri_migrations.InitInternalKartas(ctx)
	if err != nil {
		return err
	}

	return nil
}
