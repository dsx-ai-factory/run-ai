/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package connections

import (
	"github.com/run-ai/runai/control-plane/metrics-service/internal/generated/pkg/api"
)

type Connection struct {
	Url              string
	Headers          map[string]string
	AuthConfigurator AuthConfigurator
}

func (sc *Connection) UrlOrDefault(defaultUrl string) string {
	if sc.Url == "" {
		return defaultUrl
	}
	return sc.Url
}

func newConnection(metricStore api.MetricStoreEndpoint) (*Connection, error) {
	authConfigurator, err := getAuthConfigurator(metricStore.Auth)
	if err != nil {
		return nil, err
	}
	return &Connection{
		Url:              metricStore.Url,
		Headers:          metricStore.Headers,
		AuthConfigurator: authConfigurator,
	}, nil
}

func getAuthConfigurator(auth *api.MetricStoreEndpointAuth) (AuthConfigurator, error) {
	if auth == nil {
		return &EmptyAuthConfigurator{}, nil
	}
	if auth.Basic == nil || auth.Basic.Username == "" || auth.Basic.Password == "" {
		return &EmptyAuthConfigurator{}, nil
	}

	return &BasicAuthConfigurator{
		Username: auth.Basic.Username,
		Password: auth.Basic.Password,
	}, nil
}
