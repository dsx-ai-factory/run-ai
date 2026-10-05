/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validator

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type ServiceAccountValidator struct {
	client.Client
	uidAndFunctions        []ValidatingFunction
	allowedServiceAccounts []string
}

type ValidatingFunction struct {
	Func func(ctx context.Context, client client.Client) (string, error)
	Id   string
}

func NewServiceAccountValidator(k8sClient client.Client, allowedServiceAccounts []string, uidAndFunctions []ValidatingFunction) *ServiceAccountValidator {
	return &ServiceAccountValidator{
		Client:                 k8sClient,
		allowedServiceAccounts: allowedServiceAccounts,
		uidAndFunctions:        uidAndFunctions,
	}
}

func (v *ServiceAccountValidator) Validate(ctx context.Context, req admission.Request) (bool, error) {
	// Explicit identities may be configured for controllers outside this cluster.
	if AllowedServiceAccounts(req.UserInfo.Username, v.allowedServiceAccounts) {
		return true, nil
	}
	for _, f := range v.uidAndFunctions {
		var funcId string
		var err error
		if f.Id != "" {
			funcId = f.Id
		} else {
			funcId, err = f.Func(ctx, v.Client)
			if err != nil {
				return false, err
			}
			f.Id = funcId
		}
		if req.UserInfo.UID == funcId {
			return true, nil
		}
	}
	return false, nil
}

func AllowedServiceAccounts(username string, allowedServiceAccounts []string) bool {
	if username == "" {
		return false
	}
	// Match the full identity, including the namespace for Kubernetes service accounts.
	for _, allowedUsername := range allowedServiceAccounts {
		if username == allowedUsername {
			return true
		}
	}
	return false
}
