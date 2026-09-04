/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao_crud

import (
	"errors"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"
)

const (
	presetsSaveUpdate = "UPDATE crud_templates_presets SET id=?,preset_template_id=?,tenant_id=?,is_deleted=?,tenant_template_id=? WHERE id = ?"
)

var _ = Describe("saveTemplatesPresetsOfTenant", func() {
	It("saves the record", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(presetsSaveUpdate).Returns(1)
		m.ExpectTransaction().Commit()

		rec := &models_crud.TemplatesPresets{
			ID:               uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			TenantId:         7,
			PresetTemplateId: "preset-1",
			IsDeleted:        true,
		}

		apiErr := saveTemplatesPresetsOfTenant(testCtx(7), runaidb.DB(), rec)

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the update fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(presetsSaveUpdate).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		rec := &models_crud.TemplatesPresets{
			ID:               uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			TenantId:         7,
			PresetTemplateId: "preset-1",
		}

		apiErr := saveTemplatesPresetsOfTenant(testCtx(7), runaidb.DB(), rec)

		Expect(apiErr).ToNot(BeNil())
	})
})
