/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package containermetrics

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// EtcdDBSizeRow is etcd's on-disk database size over the run. Kept out of the CPU/memory
// table because it is a disk measurement in its own right, and because it is another thing
// to configure when running at scale: the quota has to be sized against it.
type EtcdDBSizeRow struct {
	Metric   string `json:"metric"`
	MinBytes string `json:"min_bytes"`
	P95Bytes string `json:"p95_bytes"`
	MaxBytes string `json:"max_bytes"`
}

// ExtractEtcdDBSize reads etcd_mvcc_db_total_size_in_bytes — the size the quota is enforced
// against. Returns nil if the metric is not scraped.
func ExtractEtcdDBSize(ctx context.Context, api v1.API, window time.Duration, endTime time.Time) ([]EtcdDBSizeRow, error) {
	dur := fmt.Sprintf("%ds", int(window.Seconds()))
	base := "max(etcd_mvcc_db_total_size_in_bytes)"

	value := func(query string) (float64, error) {
		res, _, _, err := api.Query(ctx, query, endTime)
		if err != nil {
			return 0, fmt.Errorf("etcd db size query failed: %w", err)
		}
		if vec, ok := res.(model.Vector); ok && len(vec) > 0 {
			return float64(vec[0].Value), nil
		}
		return 0, nil
	}

	lo, err := value(fmt.Sprintf(`min_over_time((%s)[%s:1m])`, base, dur))
	if err != nil {
		return nil, err
	}
	mid, err := value(fmt.Sprintf(`quantile_over_time(0.95, (%s)[%s:1m])`, base, dur))
	if err != nil {
		return nil, err
	}
	hi, err := value(fmt.Sprintf(`max_over_time((%s)[%s:1m])`, base, dur))
	if err != nil {
		return nil, err
	}
	if hi == 0 {
		return nil, nil
	}

	return []EtcdDBSizeRow{{
		Metric:   "etcd_db_total_size_bytes",
		MinBytes: bytes(lo),
		P95Bytes: bytes(mid),
		MaxBytes: bytes(hi),
	}}, nil
}

// WriteEtcdDBSize emits its own files alongside the CPU/memory tables.
func WriteEtcdDBSize(rows []EtcdDBSizeRow, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", outputDir, err)
	}

	data, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outputDir, "etcd-db-size.json"), data, 0o644); err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(outputDir, "etcd-db-size.csv"))
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	if err := w.Write([]string{"metric", "min_bytes", "p95_bytes", "max_bytes"}); err != nil {
		return err
	}
	for _, r := range rows {
		if err := w.Write([]string{r.Metric, r.MinBytes, r.P95Bytes, r.MaxBytes}); err != nil {
			return err
		}
	}
	return nil
}
