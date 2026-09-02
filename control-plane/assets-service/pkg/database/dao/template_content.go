/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

func writeTemplateContent(tx orm.DB, templateId string, containedAssetIds []string) error {
	templateContent := make([]dbmodel.TemplateContentRec, 0)

	for _, id := range containedAssetIds {
		templateContent = append(templateContent, dbmodel.TemplateContentRec{TemplateAssetId: templateId, ContainedAssetId: id})
	}

	for _, containedAsset := range templateContent {
		containedAsset.TemplateAssetId = templateId
		if _, err := tx.Create(containedAsset); err != nil {
			return err
		}
	}

	return nil
}

func getTemplateContent(tx orm.DB, templateId string, assetsRef *api.AssetsRef) ([]string, int64, *api.Error, error) {
	rowsAffected, containedIds, err := GetTemplateAssetIds(tx, templateId)
	if err != nil {
		return nil, rowsAffected, nil, err
	}

	assets := make([]dbmodel.AssetMetaRec, 0)

	record := dbmodel.AssetMetaRec{}

	rows, scanErr := tx.Model(record).
		Select("*").
		Where("id in ?", containedIds).
		Order(&record.Id).
		Scan(&assets)
	if scanErr != nil {
		return nil, rows, nil, scanErr
	}

	assetRecToAssetRef(assetsRef, assets, nil)

	return containedIds, rows, nil, nil
}

func updateTemplateContent(tx orm.DB, templateId string, requestedIds []string) *api.Error {
	_, existingIds, err := GetTemplateAssetIds(tx, templateId)
	if err != nil {
		return api.NewError(http.StatusServiceUnavailable, "failed to obtain template content", err)
	}

	compareIds := utils.CompareIds(existingIds, requestedIds)

	for containedId, comparison := range compareIds {
		if comparison.Requested && comparison.Existing {
			continue
		}
		record := dbmodel.TemplateContentRec{
			TemplateAssetId:  templateId,
			ContainedAssetId: containedId,
		}
		if comparison.Existing {
			_, err := tx.Where(&record).Delete(&record)
			if err != nil {
				return api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to remove asset %s from template %s", containedId, templateId), err)
			}
		}
		if comparison.Requested {
			_, err := tx.Create(&record)
			if err != nil {
				return api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to add asset %s to template %s", containedId, templateId), err)
			}
		}
	}

	return nil
}
