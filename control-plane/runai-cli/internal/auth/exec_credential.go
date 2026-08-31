/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	authv1 "k8s.io/client-go/pkg/apis/clientauthentication/v1"
)

const (
	KindExecCredential = "ExecCredential" //nolint:gosec
)

// NewExecCredential returns an ExecCredential for the given token.
func NewExecCredential(token string) authv1.ExecCredential {
	return authv1.ExecCredential{
		TypeMeta: metav1.TypeMeta{
			APIVersion: authv1.SchemeGroupVersion.String(),
			Kind:       KindExecCredential,
		},
		Status: &authv1.ExecCredentialStatus{
			Token: token,
		},
	}
}
