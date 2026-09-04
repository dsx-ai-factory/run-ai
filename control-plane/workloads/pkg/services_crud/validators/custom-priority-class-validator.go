/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validators

import (
	"context"
	"fmt"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_integration"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"net/http"
	"strconv"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
)

func ValidateCustomPriorityClass(ctx context.Context, clusterId string, priorityClass workloadsApi.NullableString) (map[string]int32, *workloadsApi.SubmissionErrorV2) {
	if !convutils.HasValue(priorityClass) {
		return nil, nil
	}
	clusterCapabilityChecker := checker.NewClusterCapabilitiesChecker(clusterId)
	clusterSupportMin, err := clusterCapabilityChecker.ValidateClusterCapability(ctx, checker.MinVersionOfFlexiblePriorityClasses)
	if err != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, err.Error())
	}
	if !clusterSupportMin {
		return nil, nil
	}

	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, apiErr.Message)
	}
	tenantIdStr := strconv.Itoa(int(tenantId))

	dbPriorityClassesNames, priorityValues, err := services_integration.GetClusterPriorityNamesAndValues(ctx, tenantIdStr, clusterId)
	if err != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, err.Error())
	}

	priorityClassName := *priorityClass.Get()
	if !dbPriorityClassesNames.Has(priorityClassName) {
		errMessage := composeErrorMessage(priorityClassName, dbPriorityClassesNames)
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, errMessage)
	}

	return priorityValues, nil
}

func composeErrorMessage(priorityClass string, priorityClassesNames datastructs.Set[string]) string {
	validPrioritiesStr := strings.Join(priorityClassesNames.Keys(), ",")
	errMessage := fmt.Sprintf("Error: invalid argument \"%s\" for priorityClass field: valid values are <%s>",
		priorityClass, validPrioritiesStr)
	return errMessage
}
