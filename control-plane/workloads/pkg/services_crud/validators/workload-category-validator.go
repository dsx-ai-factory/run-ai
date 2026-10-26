/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validators

import (
	"context"
	"fmt"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_integration"
	"net/http"
	"strings"

	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
)

func ValidateWorkloadCategory(ctx context.Context, category workloadsApi.NullableString) *workloadsApi.SubmissionErrorV2 {
	if !convutils.HasValue(category) {
		return nil
	}

	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, apiErr.Message)
	}

	categories, err := services_integration.GetDBCategoriesNames(ctx, tenantId)
	if err != nil {
		return workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, err.Message)
	}
	categoryName := *category.Get()

	if !categories.Has(categoryName) {
		errMessage := composeCategoriesErrorMessage(categoryName, categories)
		return workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, errMessage)
	}

	return nil
}

func composeCategoriesErrorMessage(category string, categories datastructs.Set[string]) string {
	validCategoriesStr := strings.Join(categories.Keys(), ",")
	errMessage := fmt.Sprintf("Error: invalid argument \"%s\" for categoty field: valid values are <%s>",
		category, validCategoriesStr)
	return errMessage
}
