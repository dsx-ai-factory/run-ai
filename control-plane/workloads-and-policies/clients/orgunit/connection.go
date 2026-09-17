/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orgunit

import (
	org_unit_service_client "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client"
)

type FactoryInterface interface {
	Client() ClientInterface
}

var factory FactoryInterface = RealFactory{}

func Get() FactoryInterface {
	return factory
}

type RealFactory struct{}

func (c RealFactory) Client() ClientInterface {
	return &Client{
		apiClient: org_unit_service_client.NewOrgUnitClient(orgUnitServiceUrl),
	}
}

func SetMock(clientFactory FactoryInterface) {
	factory = clientFactory
}

var orgUnitServiceUrl string

func InitConnection(orgUnitServiceUrlStr string) {
	orgUnitServiceUrl = orgUnitServiceUrlStr
}
