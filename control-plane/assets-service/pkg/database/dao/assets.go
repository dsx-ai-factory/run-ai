/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package dao (file assets.go): generic helpers for k8s-object-backed assets
// stored in the shared assets_spec table. Per-kind DAOs (rct.go, and future
// k8s-object asset kinds) build on these helpers.

package dao

import (
	"context"
	"net/http"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"github.com/google/uuid"
)

func createAssetInTransaction(
	ctx context.Context,
	tx orm.DB,
	kind api.AssetKind,
	meta api.AssetCreationRequest,
	spec any,
	nameValidationQuery *database.ParametrizedQuery,
) (*api.AssetMeta, *api.Error) {
	id := uuid.New().String()

	assetMeta, err := createAssetMeta(ctx, tx, id, kind, meta, nameValidationQuery)
	if err != nil {
		if strings.Contains(err.Error(), alreadyExistsError) {
			return nil, database.ErrorWithCode(database.CREATE, kind, http.StatusConflict).WithDetails(err)
		}
		return nil, database.Error(database.CREATE, kind).WithDetails(err)
	}

	specMap, marshalErr := dbmodel.ToJsonBMap(spec)
	if marshalErr != nil {
		return nil, database.ErrorWithCode(database.CREATE, kind, http.StatusBadRequest).WithDetails(marshalErr)
	}

	specRec := dbmodel.AssetSpecRec{
		AssetId:  id,
		TenantId: auth.TenantIdOf(ctx),
		Spec:     specMap,
	}

	if _, err := tx.Create(specRec); err != nil {
		return nil, database.Error(database.CREATE, kind).WithDetails(err)
	}

	return assetMeta, nil
}
