/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package errors

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/run-ai/runai/runai-common-packages/runaihttp"
)

type DuplicateKeyError struct {
	err     *pgconn.PgError
	Details string
}

type HumanizeError interface {
	Humanize() string
}

func NewDuplicateKeyErrorHumanizer(err *pgconn.PgError) *DuplicateKeyError {
	return &DuplicateKeyError{
		err: err,
	}
}

func TryCreateDuplicateKeyError(err error) *DuplicateKeyError {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr); pgErr.Code == runaihttp.ErrDuplicateKey {
		return NewDuplicateKeyErrorHumanizer(pgErr)
	}
	return nil
}

func (e *DuplicateKeyError) Humanize() (humanized string) {
	err := e.err
	if err == nil {
		return ""
	}
	detail := err.Detail
	tableName := err.TableName
	if detail == "" || tableName == "" {
		return ""
	}
	re := regexp.MustCompile(`Key \((.*)\)=\((.*)\)`)
	matches := re.FindStringSubmatch(detail)
	if len(matches) > 2 {
		keys := strings.Split(matches[1], ",")
		values := strings.Split(matches[2], ",")
		if len(keys) > 0 && len(values) > 0 && len(keys) == len(values) {
			lastKey := strings.TrimSpace(keys[len(keys)-1])
			lastValue := strings.TrimSpace(values[len(values)-1])
			humanized = fmt.Sprintf("%s %s %s already exists", tableName, lastKey, lastValue)
			return
		}
	}
	return ""
}
