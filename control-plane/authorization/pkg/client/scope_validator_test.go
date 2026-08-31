/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package authorization_client

import (
	"testing"

	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/pointer"
)

func initTest(t *testing.T) (testSubject *ScopeValidatorImp) {
	testSubject = &ScopeValidatorImp{}
	return
}

func TestIsPermittedScopesEmpty_NilScopes(t *testing.T) {
	testSubject := initTest(t)
	res := testSubject.IsPermittedScopesEmpty(nil)
	assert.True(t, res)
}

func TestIsPermittedScopesEmpty_True(t *testing.T) {
	testSubject := initTest(t)

	res1 := testSubject.IsPermittedScopesEmpty(nil)
	res2 := testSubject.IsPermittedScopesEmpty(&authorization_api.PermittedScopes{
		System:  nil,
		Tenants: []string{},
	})
	res3 := testSubject.IsPermittedScopesEmpty(&authorization_api.PermittedScopes{
		System:  pointer.Bool(false),
		Tenants: []string{},
	})

	assert.True(t, res1)
	assert.True(t, res2)
	assert.True(t, res3)
}

func TestIsPermittedScopesEmpty_False(t *testing.T) {
	testSubject := initTest(t)
	res1 := testSubject.IsPermittedScopesEmpty(&authorization_api.PermittedScopes{
		System:  pointer.Bool(true),
		Tenants: []string{"1", "2"},
	})
	res2 := testSubject.IsPermittedScopesEmpty(&authorization_api.PermittedScopes{
		System:  nil,
		Tenants: []string{"1", "2"},
	})
	res3 := testSubject.IsPermittedScopesEmpty(&authorization_api.PermittedScopes{
		System:  pointer.Bool(true),
		Tenants: []string{},
	})
	res4 := testSubject.IsPermittedScopesEmpty(&authorization_api.PermittedScopes{
		System:  pointer.Bool(false),
		Tenants: []string{"1"},
	})

	assert.False(t, res1)
	assert.False(t, res2)
	assert.False(t, res3)
	assert.False(t, res4)
}
