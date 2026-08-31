/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/run-ai/runai/runai-common-packages/config"
	dbconfig "github.com/run-ai/runai/runai-common-packages/runaidb/config"
)

const lockRetryInterval = 100 * time.Millisecond

type Locker struct {
	conn    *sql.Conn
	retries int
}

func NewLocker(conn *sql.Conn) *Locker {
	conf := config.Get[*dbconfig.Config]()
	return &Locker{
		conn:    conn,
		retries: conf.Migration.LockRetries,
	}
}

func (pl *Locker) Acquire(schema string) (bool, error) {
	query := "SELECT pg_try_advisory_lock(to_regnamespace($1)::integer)"
	maxAttempts := pl.retries + 1
	var err error
	for i := 0; i < maxAttempts; i++ {
		var acquired sql.NullBool
		err = pl.conn.QueryRowContext(context.Background(), query, schema).Scan(&acquired)
		if err == nil && acquired.Valid && acquired.Bool {
			return true, nil
		}

		if err != nil {
			err = fmt.Errorf("failed to acquire lock (attempt %d/%d): %w", i+1, maxAttempts, err)
		}

		if i+1 < maxAttempts {
			time.Sleep(lockRetryInterval)
		}
	}

	return false, err
}

func (pl *Locker) Release(schema string) (bool, error) {
	var released sql.NullBool
	query := "SELECT pg_advisory_unlock(to_regnamespace($1)::integer)"
	err := pl.conn.QueryRowContext(context.Background(), query, schema).Scan(&released)
	return released.Valid && released.Bool, err
}
