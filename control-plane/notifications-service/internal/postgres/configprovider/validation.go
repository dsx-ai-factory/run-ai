/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package configprovider

import (
	"fmt"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/postgres/models"
	"gorm.io/gorm"
)

type ValidationDao interface {
	InsertValidationIfNotExist(encrypted string) (*models.Validation, error)
}

type postgresValidationDao struct {
	db     *gorm.DB
	schema string
}

func (pvd *postgresValidationDao) InsertValidationIfNotExist(encrypted string) (*models.Validation, error) {
	record, err := pvd.getRecord()
	if err != nil {
		return nil, err
	}

	if record != nil {
		return record, nil
	}

	newRec := &models.Validation{Data: encrypted}
	return newRec, pvd.db.Begin().Table(pvd.validationsTable()).
		Assign(models.Validation{Data: encrypted}).
		FirstOrCreate(newRec).Commit().Error
}

func (pvd *postgresValidationDao) getRecord() (*models.Validation, error) {
	validation := make([]*models.Validation, 0)
	if err := pvd.db.
		Table(pvd.validationsTable()).Find(&validation).Error; err != nil {
		return nil, err
	}

	if len(validation) == 0 {
		return nil, nil
	}

	return validation[0], nil
}

func (pvd *postgresValidationDao) validationsTable() string {
	return fmt.Sprintf("%s.%s", pvd.schema, models.ValidationTable)
}
