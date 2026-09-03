/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaihttp

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"github.com/jackc/pgx/v5/pgconn"
)

const ErrForeignKey = "23503"
const ErrDuplicateKey = "23505"
const ErrInvalidSyntax = "22P02"
const ErrTextNotFound = "record not found" // first() in some cases returns plain error
const ErrTriggerFailed = "P0001"

func LoadErrorDetails(code *int32, message *string, details *string, goError error) {
	if goError == nil {
		return
	}

	var pgErr *pgconn.PgError
	hasPgErr := errors.As(goError, &pgErr)

	if hasPgErr {
		switch pgErr.Code {
		case ErrDuplicateKey:
			*code = http.StatusConflict
			*details = pgErr.Detail
			return
		case ErrInvalidSyntax:
			*code = http.StatusBadRequest
			*details = pgErr.Message
			return
		case ErrForeignKey:
			*code = http.StatusPreconditionFailed
			*details = fmt.Sprintf("The request invalidates %s dependency", pgErr.ConstraintName)
			return
		case ErrTriggerFailed:
			*code = http.StatusConflict
			*message = pgErr.Message
			return
		}
	}

	switch {
	case strings.Contains(goError.Error(), ErrTextNotFound), errors.Is(goError, orm.ErrRecordNotFound):
		*code = http.StatusNotFound
		*details = "requested resource does not exist"
	case errors.Is(goError, orm.ErrDuplicatedKey):
		*code = http.StatusConflict
		*details = orm.ErrDuplicatedKey.Error()
	case errors.Is(goError, orm.ErrForeignKeyViolated):
		*code = http.StatusPreconditionFailed
		*details = orm.ErrForeignKeyViolated.Error()
	case errors.Is(goError, orm.ErrCheckConstraintViolated):
		*code = http.StatusConflict
		*details = orm.ErrCheckConstraintViolated.Error()
	case errors.Is(goError, orm.ErrReadOnlyConnection):
		*code = http.StatusServiceUnavailable
		*details = orm.ErrReadOnlyConnection.Error()
	case errors.Is(goError, orm.ErrUndefinedTable):
		*code = http.StatusInternalServerError
		*details = orm.ErrUndefinedTable.Error()
	case hasPgErr:
		// avoid setting details with postgres error, as it exposes internal information
	default:
		*details = goError.Error()
	}
}
