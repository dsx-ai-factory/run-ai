/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package fake

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth"
)

// FakeValidToken is a synthetic JWT: fake issuer/subject, no real signature ("mock").
const FakeValidToken = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjk5OTk5OTk5OTksImlhdCI6MTYwMDAwMDAwMCwiaXNzIjoiaHR0cHM6Ly9mYWtlLWlzc3Vlci5leGFtcGxlLmNvbS9hdXRoL3JlYWxtcy9ydW5haSIsImF1ZCI6WyJydW5haSIsImJyb2tlciIsImFjY291bnQiXSwic3ViIjoidGVzdEBleGFtcGxlLmNvbSIsInR5cCI6IkJlYXJlciIsImF6cCI6ImNsaSIsImFsbG93ZWRfb3JpZ2lucyI6WyJodHRwczovL2Zha2UtaXNzdWVyLmV4YW1wbGUuY29tIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJvZmZsaW5lX2FjY2VzcyIsInVtYV9hdXRob3JpemF0aW9uIiwiZGVmYXVsdC1yb2xlcy1ydW5haSJdfSwicmVzb3VyY2VfYWNjZXNzIjp7ImJyb2tlciI6eyJyb2xlcyI6WyJyZWFkLXRva2VuIl19LCJhY2NvdW50Ijp7InJvbGVzIjpbIm1hbmFnZS1hY2NvdW50IiwibWFuYWdlLWFjY291bnQtbGlua3MiLCJ2aWV3LXByb2ZpbGUiXX19LCJzY29wZSI6Im9wZW5pZCBvZmZsaW5lX2FjY2VzcyBwcm9maWxlIGVtYWlsIiwidGVuYW50X2lkIjoxMDAxLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwic3ViamVjdF91dWlkIjoiNThiYmY0NmEtMTI0NS00ODYyLWEyZmQtMjdhNWU1MjAyOGI3IiwicHJlZmVycmVkX3VzZXJuYW1lIjoidGVzdEBleGFtcGxlLmNvbSIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSJ9.mock" //nolint:gosec,lll // no security hazard

type AlwaysValidToken struct {
}

func NewValidTokenBackend(getter auth.ConfigGetter) (auth.TokenService, error) {
	return NewAlwaysValidTokenManagerBackend(), nil
}

func NewAlwaysValidTokenManagerBackend() *AlwaysValidToken {
	return &AlwaysValidToken{}
}

func (a *AlwaysValidToken) GetAccessToken(_ context.Context) (string, error) {
	return FakeValidToken, nil
}
