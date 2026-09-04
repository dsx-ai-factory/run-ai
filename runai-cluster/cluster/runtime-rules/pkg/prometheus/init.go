/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package prometheus

import (
	"os"

	"go.uber.org/zap"
)

// GetQuerierClient initializes the prometheus client
func GetQuerierClient(startPromClient bool, promServiceUrl string) Querier {
	zap.S().Info("Initializing prometheus client")

	zap.S().Debugf("Got prometheus service url %s", promServiceUrl)

	var querier Querier
	var err error

	if startPromClient {
		querier, err = NewPromQuerier(promServiceUrl)
		if err != nil {
			zap.S().Errorf("Failed to start prometheus querier. error: %v", err.Error())
			os.Exit(1)
		}
		zap.S().Infof("prometheus client start with url %s", promServiceUrl)
	} else {
		zap.S().Warn("Not starting prometheus client")
	}
	return querier
}
