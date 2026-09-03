/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"k8s.io/apimachinery/pkg/fields"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

var logUtils = logger.Log.WithName("utils")

func ListAccessRules(ctx context.Context, k8sClient client.Client, fieldSelector *fields.Selector) ([]v1alpha1.AccessRule, error) {
	listOptions := &client.ListOptions{}
	fieldSelectorLogStr := ""
	if fieldSelector != nil {
		listOptions.FieldSelector = *fieldSelector
		fieldSelectorLogStr = fmt.Sprintf(" with field selector: %s", (*fieldSelector).String())
	}

	accessRulesList := &v1alpha1.AccessRuleList{}
	err := k8sClient.List(ctx, accessRulesList, listOptions)
	if err != nil {
		logUtils.Error(err, fmt.Sprintf("Failed listing AccessRules%s", fieldSelectorLogStr))
		return []v1alpha1.AccessRule{}, err
	}

	return accessRulesList.Items, nil
}
