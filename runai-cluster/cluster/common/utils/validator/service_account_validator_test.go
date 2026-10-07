/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validator

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func TestAllowedServiceAccounts(t *testing.T) {
	for _, tc := range []struct {
		name     string
		username string
		allowed  []string
		want     bool
	}{
		{name: "unlisted tenant service account", username: "system:serviceaccount:tenant:default"},
		{name: "unlisted system service account", username: "system:serviceaccount:kube-system:coredns"},
		{name: "exact identity", username: "system:serviceaccount:runai:controller", allowed: []string{"other", "system:serviceaccount:runai:controller"}, want: true},
		{name: "same name in another namespace", username: "system:serviceaccount:tenant:controller", allowed: []string{"system:serviceaccount:runai:controller"}},
		{name: "service account name substring", username: "system:serviceaccount:tenant:my-controller-x", allowed: []string{"controller"}},
		{name: "identity suffix", username: "system:serviceaccount:runai:controller-extra", allowed: []string{"system:serviceaccount:runai:controller"}},
		{name: "identity prefix", username: "prefix-system:serviceaccount:runai:controller", allowed: []string{"system:serviceaccount:runai:controller"}},
		{name: "empty entry", username: "alice", allowed: []string{""}},
		{name: "empty identity", allowed: []string{""}},
		{name: "explicit external identity", username: "nmc-controller", allowed: []string{"nmc-controller"}, want: true},
		{name: "external identity substring", username: "prefix-nmc-controller-suffix", allowed: []string{"nmc-controller"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, AllowedServiceAccounts(tc.username, tc.allowed))
		})
	}
}

func TestServiceAccountValidatorValidate(t *testing.T) {
	lookupError := errors.New("service account lookup failed")
	for _, tc := range []struct {
		name      string
		username  string
		uid       string
		allowed   []string
		functions []ValidatingFunction
		want      bool
		wantError error
	}{
		{name: "tenant cannot bypass UID check", username: "system:serviceaccount:tenant:default", uid: "tenant-uid", functions: []ValidatingFunction{{Id: "trusted-uid"}}},
		{name: "trusted UID remains allowed", username: "system:serviceaccount:runai:controller", uid: "trusted-uid", functions: []ValidatingFunction{{Id: "trusted-uid"}}, want: true},
		{name: "resolved UID remains allowed", username: "system:serviceaccount:runai:controller", uid: "trusted-uid", functions: []ValidatingFunction{{Func: func(context.Context, client.Client) (string, error) { return "trusted-uid", nil }}}, want: true},
		{name: "no UID checks denies unlisted identity", username: "system:serviceaccount:tenant:default"},
		{name: "exact allowlist needs no UID checks", username: "system:serviceaccount:runai:controller", allowed: []string{"system:serviceaccount:runai:controller"}, want: true},
		{name: "lookup errors fail closed", username: "system:serviceaccount:tenant:default", functions: []ValidatingFunction{{Func: func(context.Context, client.Client) (string, error) { return "", lookupError }}}, wantError: lookupError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v := NewServiceAccountValidator(nil, tc.allowed, tc.functions)
			allowed, err := v.Validate(context.Background(), admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{
				UserInfo: authenticationv1.UserInfo{Username: tc.username, UID: tc.uid},
			}})
			require.ErrorIs(t, err, tc.wantError)
			assert.Equal(t, tc.want, allowed)
		})
	}
}
