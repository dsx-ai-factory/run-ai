/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/services"
)

const quantilePlaceholder = "QUANTILE"

type quantileInjector struct {
	next QueryInjector
}

func NewQuantileInjector(next QueryInjector) *quantileInjector {
	return &quantileInjector{next: next}
}

func (q *quantileInjector) Inject(query string, metricInfo *services.MetricInformation) (string, error) {
	if metricInfo.Quantile == nil {
		return query, nil
	}

	if strings.Count(query, quantilePlaceholder) == 0 {
		return "", fmt.Errorf("quantile metric query must contain %s placeholder", quantilePlaceholder)
	}
	injectedQuery := strings.ReplaceAll(query, quantilePlaceholder, fmt.Sprintf("%.2f", *metricInfo.Quantile))
	if q.next == nil {
		return injectedQuery, nil
	}
	return q.next.Inject(injectedQuery, metricInfo)
}
